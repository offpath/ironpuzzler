package puzzle

import (
	"hunt"
	"team"

	"appengine"
	"appengine/datastore"

	"strings"
	"time"
)

const (
	puzzleKind = "Puzzle"
	solveKind = "Solve"
	maxPointValue = 15
	minPointValue = 10
)

type Puzzle struct {
	Number int
	Name string
	Answer string
	Paper bool
	Team *datastore.Key `json:"-"`
	PointValue int
	
	ID string `datastore:"-"`
	Key *datastore.Key `datastore:"-" json:"-"`
}

type Solve struct {
	Team *datastore.Key
	Puzzle *datastore.Key
	Time time.Time
	Points int
}

type AdminPuzzle struct {
	Puzzle
	TeamName string
}

type UpdatableProgressInfo struct {
	AvailablePoints int
	Solved bool
	GrantedPoints int
	SolveTimes []string
	Answerable bool
}

func (p *Puzzle) enkey(k *datastore.Key) {
	p.Key = k
	p.ID = k.Encode()
}

func (p *Puzzle) ReRead(c appengine.Context) *Puzzle {
	return Key(c, p.Key)
}

func (p *Puzzle) Write(c appengine.Context) {
	_, err := datastore.Put(c, p.Key, p)
	if err != nil {
		c.Errorf("Write: %v", err)
	}
}

func (p *Puzzle) UpdatableProgressInfo(c appengine.Context, h *hunt.Hunt, t *team.Team) UpdatableProgressInfo {
	u := UpdatableProgressInfo{
		AvailablePoints: p.PointValue,
		Solved: false,
		GrantedPoints: 0,
		SolveTimes: nil,
		Answerable: t != nil && !t.Key.Equal(p.Team),
	}
	var solves []Solve
	_, err := datastore.NewQuery(solveKind).Ancestor(h.Key).Filter("Puzzle =", p.Key).Filter("Team =", t.Key).Limit(1).GetAll(c, &solves)
	if err != nil {
		c.Errorf("Error: %v", err)
	}
	if len(solves) > 0 {
		u.Solved = true
		u.GrantedPoints = solves[0].Points
		u.Answerable = false
	}
	solves = nil
	_, err = datastore.NewQuery(solveKind).Ancestor(h.Key).Filter("Puzzle =", p.Key).Order("Time").GetAll(c, &solves)
	if err != nil {
		c.Errorf("Error: %v", err)
	}
	for _, s := range solves {
		u.SolveTimes = append(u.SolveTimes, s.Time.In(h.GetTimezone(c)).Format("15:04:05"))
	}
	return u
}

func (p *Puzzle) Admin(c appengine.Context) *AdminPuzzle {
	var teamName string
	if t := team.Key(c, p.Team); t != nil {
		teamName = t.Name
	}
	return &AdminPuzzle{
		Puzzle: *p,
		TeamName: teamName,
	}
}

func normalize(str string) string {
	tmp := ""
	for _, c := range strings.ToLower(str) {
		if c >= 'a' && c <= 'z' {
			tmp += string(c)
		}
	}
	return tmp
}

func (p *Puzzle) SubmitAnswer(c appengine.Context, h *hunt.Hunt, t *team.Team, answer string) bool {
	if normalize(p.Answer) != normalize(answer) {
		return false
	}

	if t.Key.Equal(p.Team) {
		// The answer is correct, but it's for the team's own puzzle!
		return true
	}

	keys, err := datastore.NewQuery(solveKind).Ancestor(h.Key).Filter("Puzzle =", p.Key).Filter("Team =", t.Key).Limit(1).KeysOnly().GetAll(c, nil)
	if err != nil {
		return false
	}
	if len(keys) > 0 {
		// The answer is correct, but the team has already solved the puzzle.
		return true
	}
	solve := &Solve{
		Team: t.Key,
		Puzzle: p.Key,
		Time: time.Now(),
		Points: p.PointValue,
	}
	k := datastore.NewIncompleteKey(c, solveKind, h.Key)
	k, err = datastore.Put(c, k, solve)
	if err != nil {
		return false
	}
	if p.PointValue > minPointValue {
		p.PointValue--
		p.Write(c)
	}
	return true
}

func ID(c appengine.Context, id string) *Puzzle {
	k, err := datastore.DecodeKey(id)
	if err != nil {
		return nil
	}
	var p Puzzle
	err = datastore.Get(c, k, &p)
	if err != nil {
		return nil
	}
	p.enkey(k)
	return &p
}

func Key(c appengine.Context, k *datastore.Key) *Puzzle {
	var p Puzzle
	err := datastore.Get(c, k, &p)
	if err != nil {
		return nil
	}
	p.enkey(k)
	return &p
}

func All(c appengine.Context, h *hunt.Hunt, t *team.Team) []*Puzzle {
	var puzzles []*Puzzle
	q := datastore.NewQuery(puzzleKind).Ancestor(h.Key).Order("Number")
	if t != nil {
		q = q.Filter("Team =", t.Key)
	}
	keys, err := q.GetAll(c, &puzzles)
	if err != nil {
		c.Errorf("Error: %v", err)
		return nil
	}
	for i, k := range keys {
		puzzles[i].enkey(k)
	}
	return puzzles
}

func New(c appengine.Context, h *hunt.Hunt, t *team.Team, number int, paper bool) *Puzzle {
	newPuzzle := &Puzzle{
		Number: number,
		Paper: paper,
		Team: t.Key,
		PointValue: maxPointValue,
	}
	
	k := datastore.NewIncompleteKey(c, puzzleKind, h.Key)
	k, err := datastore.Put(c, k, newPuzzle)
	if err != nil {
		c.Errorf("Error: %v", err)
		return nil
	}
	newPuzzle.enkey(k)
	return newPuzzle
}

