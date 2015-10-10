package hunt

import (
	"regexp"

	"appengine"
	"appengine/datastore"
)

const (
	StatePreLaunch = 0
	StateEarlyAccess = 1
	StateIngredients = 2
	StateSolving = 3
	StateGrading = 4
	StateTallying = 5
	StateDone = 6

	huntKind = "Hunt"
)

var (
	pathRegexp = regexp.MustCompile("[a-z0-9]+")
)

type Hunt struct {
	Name string
	Path string
	Ingredients string
	State int

	ID string `datastore:"-"`
	Key *datastore.Key `datastore:"-" json:"-"`
}

func (h *Hunt) enkey(k *datastore.Key) {
	h.Key = k
	h.ID = k.Encode()
}

func (h *Hunt) Delete(c appengine.Context) {
	err := datastore.Delete(c, h.Key)
	if err != nil {
		c.Errorf("Error: %v", err)
	}
}

func Path(c appengine.Context, path string) *Hunt {
	var hunts []*Hunt
	keys, err := datastore.NewQuery(huntKind).Filter("Path =", path).GetAll(c, &hunts)
	if err != nil {
		c.Errorf("GetHunt: %v", err)
		return nil
	}
	if len(hunts) == 0 {
		return nil
	}
	hunts[0].enkey(keys[0])
	return hunts[0]
}

func ID(c appengine.Context, id string) *Hunt {
	k, err := datastore.DecodeKey(id)
	if err != nil {
		return nil
	}
	var h Hunt
	err = datastore.Get(c, k, &h)
	if err != nil {
		return nil
	}
	h.enkey(k)
	return &h
}

func All(c appengine.Context) []*Hunt {
	var hunts []*Hunt
	keys, err := datastore.NewQuery(huntKind).GetAll(c, &hunts)
	if err != nil {
		c.Errorf("Error: %v", err)
		return nil
	}
	for i := range keys {
		hunts[i].enkey(keys[i])
	}
	return hunts
}

func New(c appengine.Context, name string, path string) *Hunt {
	if !pathRegexp.MatchString(path) {
		return nil
	}
	newHunt := &Hunt{
		Name: name,
		Path: path,
		State: StatePreLaunch,
	}

	// TODO(dneal): Ensure no collisions on path.

	k := datastore.NewIncompleteKey(c, huntKind, nil)
	k, err := datastore.Put(c, k, newHunt)
	if err != nil {
		c.Errorf("Error: %v", err)
		return nil
	}
	newHunt.enkey(k)
	return newHunt
}

