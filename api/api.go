// Package api handles all endpoints that return json.
package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"appengine"
	"appengine/datastore"

	"adminconsole"
	"broadcast"
	"hunt"
	"puzzle"
	"tally"
	"team"
)

const (
	leaderboardListenerKind = "LeaderboardListener"
)

type TeamsInfo struct {
	Editable bool
	Teams []*team.Team
}

type TeamInfo struct {
	Name string
	ID string
}

type TeamSelector struct {
	CurrentTeam string
	BadSignIn bool
	Teams []TeamInfo
}

type IngredientInfo struct {
	Display bool
	Editable bool
	Ingredients string
}

type ProgressInfo struct {
	Number int
	Name string
	ID string
	Updatable puzzle.UpdatableProgressInfo
}

type LeaderboardInfo struct {
	Display bool
	Answerable bool
	Progress []*ProgressInfo
}

type PuzzleInfo struct {
	Editable bool
	Puzzles []*puzzle.AdminPuzzle
}

type SurveyInfo struct {
	Display bool
	Done bool
	Puzzles []*ProgressInfo
}

type TeamSurveyInfo struct {
	Name string
	Done bool
}

type AdminSurveyInfo struct {
	Display bool
	Teams []TeamSurveyInfo
}

func HuntHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) != 3 {
		http.Error(w, "Not found", 404)
		return
	}

	c := appengine.NewContext(r)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)

	var h *hunt.Hunt
	if id := r.FormValue("hunt_id"); id != "" {
		h = hunt.ID(c, id)
	}
	if h == nil {
		return
	}

	t, badSignIn := team.SignIn(c, h, r)

	var p *puzzle.Puzzle
	if id := r.FormValue("puzzleid"); id != "" {
		p = puzzle.ID(c, id)
	}

	var err error
	switch path[2] {
	case "teaminfo":
		teams := TeamSelector{
			BadSignIn: badSignIn,
		}
		if t != nil {
			teams.CurrentTeam = t.Name
		} else {
			for _, t := range team.All(c, h) {
				teams.Teams = append(teams.Teams, TeamInfo{t.Name, t.ID})
			}
		}
		err = enc.Encode(teams)
	case "ingredients":
		if h.State >= hunt.StateIngredients || (t != nil && t.Novice && h.State == hunt.StateEarlyAccess) {
			err = enc.Encode(IngredientInfo{true, false, h.Ingredients})
		}
	case "leaderboard":
		if h.State >= hunt.StateSolving {
			var l LeaderboardInfo
			fillLeaderboardInfo(c, h, t, &l)
			err = enc.Encode(l)
		}
	case "leaderboardupdate":
		if p != nil && h.State >= hunt.StateSolving {
			err = enc.Encode(p.UpdatableProgressInfo(c, h, t))
		}
	case "puzzles":
		if t != nil && (h.State >= hunt.StateIngredients || (t.Novice && h.State == hunt.StateEarlyAccess)) && h.State < hunt.StateSolving {
			puzzles := puzzle.All(c, h, t)
			var admin []*puzzle.AdminPuzzle
			for _, p := range puzzles {
				admin = append(admin, p.Admin(c))
			}
			// TODO(dneal): Check state.
			err = enc.Encode(PuzzleInfo{true, admin})
		}
	case "updatepuzzle":
		if p != nil && h.State < hunt.StateSolving{
			p.Name = r.FormValue("name");
			p.Answer = r.FormValue("answer");
			p.Write(c);
			broadcast.SendPuzzlesUpdate(c, h, t)
		}
	case "survey":
		if t != nil && h.State == hunt.StateSurveying {
			info := SurveyInfo{Display: true}
			if t.Survey != "" {
				info.Done = true
			} else {
				puzzles := puzzle.All(c, h, nil)
				for _, p := range puzzles {
					if p.Team.Equal(t.Key) {
						continue
					}
					info.Puzzles = append(info.Puzzles, &ProgressInfo{
						Number: p.Number,
						Name: p.Name,
						ID: p.ID,
					})
				}
			}
			err = enc.Encode(info)
		}
	case "submitsurvey":
		if t != nil && t.Survey == "" {
			t.Survey = r.FormValue("result")
			t.Write(c)
			broadcast.SendSurveyUpdate(c, h, t)
		}
	case "channel":
		err = enc.Encode(broadcast.GetToken(c, h, t, false))
	case "submitanswer":
		if t == nil || p == nil || h.State != hunt.StateSolving{
			break
		}
		var throttled, correct bool
		err = datastore.RunInTransaction(c, func (c appengine.Context) error {
			t := t.ReRead(c)
			p := p.ReRead(c)
			if t.Throttle(c) {
				throttled = true
			} else {
				correct = p.SubmitAnswer(c, h, t, r.FormValue("answer"))
			}
			return nil
		}, nil)
		var outcome string
		if err != nil {
			outcome = "Error, try again!"
		} else if throttled {
			outcome = "Too many answer attempts, please wait a minute and try again"
			adminconsole.Log(c, h, fmt.Sprintf("%s attempts to answer but is throttled", t.Name))
		} else if correct {
			outcome = "Correct!"
			broadcast.SendLeaderboardUpdate(c, h, p)
			adminconsole.Log(c, h, fmt.Sprintf("%s correctly answers (%d) %s", t.Name, p.Number, p.Name))
		} else {
			outcome = fmt.Sprintf("\"%s\" is an incorrect guess for %s", r.FormValue("answer"), p.Name)
			adminconsole.Log(c, h, fmt.Sprintf("%s incorrectly answers [%s] for (%d) %s", t.Name, r.FormValue("answer"), p.Number, p.Name))
		}
		err = enc.Encode(outcome)
	case "finalscores":
		finalScores := tally.Get(c, h, false)
		err = enc.Encode(finalScores)
	}


	if err != nil {
		c.Errorf("Error: %v", err)
	}
}

func updatableProgressInfo(c appengine.Context, h *hunt.Hunt, t *team.Team, p *puzzle.Puzzle) map[string]puzzle.UpdatableProgressInfo {
	result := map[string]puzzle.UpdatableProgressInfo{}
	if p != nil {
		result[p.ID] = p.UpdatableProgressInfo(c, h, t)
	} else {
		for _, p := range puzzle.All(c, h, nil) {
			result[p.ID] = p.UpdatableProgressInfo(c, h, t)
		}
	}
	return result
}

func fillLeaderboardInfo(c appengine.Context, h *hunt.Hunt, t *team.Team, l *LeaderboardInfo) {
	puzzles := puzzle.All(c, h, nil) 
	for _, p := range puzzles {
		l.Progress = append(l.Progress, &ProgressInfo{
			Number: p.Number,
			Name: p.Name,
			ID: p.ID,
			Updatable: p.UpdatableProgressInfo(c, h, t),
		})
	}
	l.Display = true
	l.Answerable = t != nil && h.State == hunt.StateSolving
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) != 4 {
		http.Error(w, "Not found", 404)
		return
	}

	c := appengine.NewContext(r)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)

	var h *hunt.Hunt
	if id := r.FormValue("hunt_id"); id != "" {
		h = hunt.ID(c, id)
	}

	var t *team.Team
	if id := r.FormValue("team_id"); h != nil && id != "" {
		t = team.ID(c, id)
	}

	var p *puzzle.Puzzle
	if id := r.FormValue("puzzleid"); h != nil && id != "" {
		p = puzzle.ID(c, id)
	}

	var err error
	switch path[3] {
	case "hunts":
		hunts := hunt.All(c)
		err = enc.Encode(hunts)
	case "addhunt":
		hunt.New(c, r.FormValue("name"), r.FormValue("path"))
	case "deletehunt":
		h.Delete(c)
	case "updateingredients":
		if h != nil && h.State < hunt.StateEarlyAccess {
			h.Ingredients = r.FormValue("ingredients")
			h.Write(c)
			broadcast.SendIngredientsUpdate(c, h)
		}
	case "teams":
		if h != nil {
			err = enc.Encode(TeamsInfo{
				Editable: true,
				Teams: team.All(c, h),
			})
		}
	case "addteam":
		if h != nil && h.State < hunt.StateEarlyAccess {
			team.New(c, h, r.FormValue("name"), r.FormValue("password"), r.FormValue("novice") == "true")
			broadcast.SendTeamsUpdate(c, h)
		}
	case "deleteteam":
		if t != nil && h.State < hunt.StateEarlyAccess {
			t.Delete(c)
			broadcast.SendTeamsUpdate(c, h)
		}
	case "state":
		if h != nil {
			err = enc.Encode(h.State)
		}
	case "advancestate":
		if h != nil {
			currentState, err := strconv.Atoi(r.FormValue("currentstate"))
			if err == nil {
				advanceState(c, h, currentState)
				broadcast.SendRefresh(c, h)
			}
		}
	case "puzzles":
		if h != nil {
			puzzles := puzzle.All(c, h, nil)
			var admin []*puzzle.AdminPuzzle
			for _, p := range puzzles {
				admin = append(admin, p.Admin(c))
			}
			err = enc.Encode(PuzzleInfo{false, admin})
		}
	case "ingredients":
		err = enc.Encode(IngredientInfo{
			Display: true,
			Editable: true,
			Ingredients: h.Ingredients,
		})
	case "leaderboard":
		if h.State >= hunt.StateSolving {
			var l LeaderboardInfo
			fillLeaderboardInfo(c, h, nil, &l)
			err = enc.Encode(l)
		}
	case "leaderboardupdate":
		if p != nil {
			err = enc.Encode(p.UpdatableProgressInfo(c, h, nil))
		}
	case "channel":
		if h != nil {
			err = enc.Encode(broadcast.GetToken(c, h, nil, true))
		}
	case "console":
		if h != nil {
			err = enc.Encode(adminconsole.Logs(c, h))
		}
	case "adminsurvey":
		if h != nil && h.State == hunt.StateSurveying {
			var info AdminSurveyInfo
			info.Display = true
			for _, t := range team.All(c, h) {
				info.Teams = append(info.Teams, TeamSurveyInfo{t.Name, t.Survey != ""})
			}
			err = enc.Encode(info)
		}
	case "finalscores":
		finalScores := tally.Get(c, h, true)
		err = enc.Encode(finalScores)
	}
	
	if err != nil {
		c.Errorf("Error: %v", err)
		return
	}
}

func advanceState(c appengine.Context, h *hunt.Hunt, currentState int) {
	err := datastore.RunInTransaction(c, func (c appengine.Context) error {
		h := hunt.ID(c, h.ID)
		if h == nil  || h.State != currentState {
			// TODO(dneal): Return a real error.
			return nil
		}
		switch h.State {
		case hunt.StatePreLaunch:
			teams := team.All(c, h)
			nonPaperOrder := rand.Perm(len(teams))
			paperOrder := rand.Perm(len(teams))
			for i := range teams {
				puzzle.New(c, h, teams[i], nonPaperOrder[i] + 1, false)
				puzzle.New(c, h, teams[i], len(teams) + paperOrder[i] + 1, true)
			}
		case hunt.StateSurveying:
			tally.BuildFinalTally(c, h)
		}
		h.State++
		h.Write(c)
		broadcast.SendRefresh(c, h)
		return nil
	}, nil)
	if err != nil {
		c.Errorf("Error: %v", err)
	}
}

