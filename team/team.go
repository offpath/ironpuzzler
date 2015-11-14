package team

import (
	"net/http"
	"time"

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
	Attempts []time.Time
	Survey string

	ID string `datastore:"-"`
	Key *datastore.Key `datastore:"-" json:"-"`

}

func (t *Team) enkey(k *datastore.Key) {
	t.Key = k
	t.ID = k.Encode()
}

func (t *Team) Write(c appengine.Context) {
	_, err := datastore.Put(c, t.Key, t)
	if err != nil {
		c.Errorf("Write: %v", err)
	}
}

func (t *Team) Throttle(c appengine.Context) bool {
	now := time.Now()
	if len(t.Attempts) < 4 {
		t.Attempts = append(t.Attempts, now)
		t.Write(c)
		return false
	}
	if now.Sub(t.Attempts[0]).Minutes() > 1.0 {
		t.Attempts[0], t.Attempts[1], t.Attempts[2], t.Attempts[3] = t.Attempts[1], t.Attempts[2], t.Attempts[3], now
		t.Write(c)
		return false
	}
	return true
}

func (t *Team) ReRead(c appengine.Context) *Team {
	return Key(c, t.Key)
}

func ID(c appengine.Context, id string) *Team {
	k, err := datastore.DecodeKey(id)
	if err != nil {
		return nil
	}
	return Key(c, k)
}

func Key(c appengine.Context, k *datastore.Key) *Team {
	var t Team
	err := datastore.Get(c, k, &t)
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
