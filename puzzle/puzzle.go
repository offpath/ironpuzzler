package puzzle

import (
	"hunt"
	"team"

	"appengine"
	"appengine/datastore"
)

const (
	puzzleKind = "Puzzle"
)

type Puzzle struct {
	Number int
	Name string
	Answer string
	Paper bool
	Team *datastore.Key `json:"-"`
	
	ID string `datastore:"-"`
	Key *datastore.Key `datastore:"-" json:"-"`
}

func (p *Puzzle) enkey(k *datastore.Key) {
	p.Key = k
	p.ID = k.Encode()
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

func All(c appengine.Context, h *hunt.Hunt, t *team.Team) []*Puzzle {
	var puzzles []*Puzzle
	q := datastore.NewQuery(puzzleKind).Ancestor(h.Key)
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

