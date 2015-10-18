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
	return UpdatableProgressInfo{
		AvailablePoints: p.PointValue,
		Solved: false,
		GrantedPoints: 0,
		SolveTimes: nil,
		Answerable: t != nil && !t.Key.Equal(p.Team),
	}
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
	// TODO(dneal): Add solves to table.
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

