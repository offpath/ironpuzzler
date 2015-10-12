package team

import (
	"net/http"

	"hunt"

	"appengine"
	"appengine/datastore"
)

const (
	teamKind = "Team"
)

type Team struct {
	Name string
	Password string
	Novice bool
	Hunt *datastore.Key `json:"-"`

	ID string `datastore:"-"`
	Key *datastore.Key `datastore:"-" json:"-"`

}

func (t *Team) enkey(k *datastore.Key) {
	t.Key = k
	t.ID = k.Encode()
}

func ID(c appengine.Context, id string) *Team {
	k, err := datastore.DecodeKey(id)
	if err != nil {
		return nil
	}
	var t Team
	err = datastore.Get(c, k, &t)
	if err != nil {
		return nil
	}
	t.enkey(k)
	return &t
}

func All(c appengine.Context, h *hunt.Hunt) []*Team {
	var teams []*Team
	keys, err := datastore.NewQuery(teamKind).Ancestor(h.Key).Order("Name").GetAll(c, &teams)
	if err != nil {
		c.Errorf("Error: %v", err)
		return nil
	}
	for i, k := range keys {
		teams[i].enkey(k)
	}
	return teams
}

func New(c appengine.Context, h *hunt.Hunt, name string, password string, novice bool) *Team {
	newTeam := &Team{
		Name: name,
		Password: password,
		Novice: novice,
		Hunt: h.Key,
	}
	
	k := datastore.NewIncompleteKey(c, teamKind, h.Key)
	k, err := datastore.Put(c, k, newTeam)
	if err != nil {
		c.Errorf("Error: %v", err)
		return nil
	}
	newTeam.enkey(k)
	return newTeam
}

func SignIn(c appengine.Context, h *hunt.Hunt, r *http.Request) (t *Team, badSignIn bool) {
	teamCookie, _ := r.Cookie("team_id")
	passCookie, _ := r.Cookie("password")

	if teamCookie == nil && passCookie == nil {
		return nil, false
	}

	t = ID(c, teamCookie.Value)
	if t == nil || !t.Hunt.Equal(h.Key) || t.Password != passCookie.Value {
		return nil, true
	}
	return t, false
}

func (t *Team) Delete(c appengine.Context) {
	err := datastore.Delete(c, t.Key)
	if err != nil {
		c.Errorf("Error: %v", err)
		return
	}
}
