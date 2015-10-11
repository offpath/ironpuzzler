// Package api handles all endpoints that return json.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"appengine"

	"hunt"
	"puzzle"
	"team"
)

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
	Ingredients string
}
	
type PuzzleInfo struct {
	Display bool
	Puzzles []*puzzle.Puzzle
}

type PageInfo struct {
	Name string
	Teams TeamSelector
	Ingredients IngredientInfo
	Puzzles PuzzleInfo
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

	var pageInfo PageInfo
	
	pageInfo.Name = h.Name

	if t != nil {
		pageInfo.Teams.CurrentTeam = t.Name
	} else {
		for _, t := range team.All(c, h) {
			pageInfo.Teams.Teams = append(pageInfo.Teams.Teams, TeamInfo{t.Name, t.ID})
		}
	}
	pageInfo.Teams.BadSignIn = badSignIn

	if h.State >= hunt.StateIngredients ||
		(h.State == hunt.StateEarlyAccess && t != nil && t.Novice) {
		pageInfo.Ingredients.Display = true
		pageInfo.Ingredients.Ingredients = h.Ingredients
		if h.State < hunt.StateSolving && t != nil {
			pageInfo.Puzzles.Display = true
			pageInfo.Puzzles.Puzzles = puzzle.All(c, h, t)
		}
	}

	enc.Encode(pageInfo)
}

func fillTeamSelector(c appengine.Context, h *hunt.Hunt, t *team.Team) {

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

	var err error
	switch path[3] {
	case "hunt":
		// Easy case
		err = enc.Encode(h)
	case "hunts":
		hunts := hunt.All(c)
		err = enc.Encode(hunts)
	case "addhunt":
		hunt.New(c, r.FormValue("name"), r.FormValue("path"))
	case "deletehunt":
		h.Delete(c)
	case "updateingredients":
		if h != nil {
			h.Ingredients = r.FormValue("ingredients")
			c.Errorf("indredients: %v", r.FormValue("ingredients"))
			h.Write(c)
		}
	case "teams":
		if h != nil {
			teams := team.All(c, h)
			err = enc.Encode(teams)
		}
	case "addteam":
		if h != nil {
			team.New(c, h, r.FormValue("name"), r.FormValue("password"), r.FormValue("novice") == "true")
		}
	case "deleteteam":
		if t != nil {
			t.Delete(c)
		}
	case "advancestate":
		if h != nil {
			if strconv.Itoa(h.State) == r.FormValue("currentstate") && h.State < hunt.StateDone {
				h.State++
				h.Write(c)
			}
		}
	}
	if err != nil {
		c.Errorf("Error: %v", err)
		return
	}
}

