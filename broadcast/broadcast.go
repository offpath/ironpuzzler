package broadcast

import (
	"fmt"
	"strings"
	"time"

	"appengine"
	"appengine/channel"
	"appengine/datastore"

	"hunt"
	"puzzle"
	"team"
)

const (
	listenerKind = "Listener"
	unauthListener = "unauth"
	adminListener = "admin"
	
	refresh = "refresh"
	leaderboardUpdate = "leaderboardupdate"
	puzzlesUpdate = "puzzlesupdate"
	surveyUpdate = "surveyupdate"
	teamsUpdate = "teamsupdate"
	consoleUpdate = "consoleupdate"
	ingredientsUpdate = "ingredientsupdate"
)

type Listener struct {
	Channel string
	Timestamp string
	TeamID string
	Admin bool
	Open bool
}

type Message struct {
	K string
	V string
}

func GetToken(c appengine.Context, h *hunt.Hunt, t *team.Team, admin bool) string {
	// Channel name format: huntID.timestamp
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	channelName := fmt.Sprintf("%s.%s", h.ID, timestamp)
	token, err := channel.Create(c, channelName)
	if err != nil {
		c.Errorf("Error: %v", err)
		return ""
	}
	listener := Listener{
		Channel: channelName,
		Timestamp: timestamp,
		Admin: admin,
		Open: false,
	}
	if t != nil {
		listener.TeamID = t.ID
	}
	_, err = datastore.Put(c, datastore.NewIncompleteKey(c, listenerKind, h.Key), &listener)
	if err != nil {
		c.Errorf("Error: %v", err)
		return ""
	}
	return token
}

func AddListener(c appengine.Context, channelName string) {
	split := strings.Split(channelName, ".")
	if len(split) != 2 {
		c.Errorf("Unexpected channel name: %s", channelName)
		return
	}
	h := hunt.ID(c, split[0])
	if h == nil {
		c.Errorf("Channel without matching hunt: %s", channelName)
		return
	}
	var listeners []*Listener
	keys, err := datastore.NewQuery(listenerKind).Ancestor(h.Key).Filter("Channel =", channelName).Limit(1).GetAll(c, &listeners)
	if err != nil {
		c.Errorf("Error: %v", err)
		return
	}
	if len(listeners) > 0 {
		listeners[0].Open = true
		_, err := datastore.Put(c, keys[0], listeners[0])
		if err != nil {
			c.Errorf("Write: %v", err)
		}
	}
}

func RemoveListener(c appengine.Context, str string) {
	keys, err := datastore.NewQuery(listenerKind).Filter("Channel =", str).KeysOnly().GetAll(c, nil)
	if err != nil {
		c.Errorf("RemoveListener: %v", err)
		return
	}
	err = datastore.DeleteMulti(c, keys)
	if err != nil {
		c.Errorf("RemoveListener: %v", err)
		return
	}
}

func SendRefresh(c appengine.Context, h *hunt.Hunt) {
	sendAll(c, h, Message{K:refresh})
}

func SendLeaderboardUpdate(c appengine.Context, h *hunt.Hunt, p *puzzle.Puzzle) {
	sendAll(c, h, Message{K:leaderboardUpdate,V:fmt.Sprintf("%s", p.ID)})
}

func SendPuzzlesUpdate(c appengine.Context, h *hunt.Hunt, t *team.Team) {
	m := Message{K:puzzlesUpdate}
	sendAdmin(c, h, m)
	sendTeam(c, h, t, m)
}

func SendSurveyUpdate(c appengine.Context, h *hunt.Hunt, t *team.Team) {
	m := Message{K:surveyUpdate}
	sendAdmin(c, h, m)
	sendTeam(c, h, t, m)
}

func SendTeamsUpdate(c appengine.Context, h *hunt.Hunt) {
	sendAdmin(c, h, Message{K:teamsUpdate})
}

func SendConsoleUpdate(c appengine.Context, h *hunt.Hunt, str string) {
	sendAdmin(c, h, Message{K:consoleUpdate,V:str})
}

func SendIngredientsUpdate(c appengine.Context, h *hunt.Hunt) {
	sendAdmin(c, h, Message{K:ingredientsUpdate})
}

func sendAll(c appengine.Context, h *hunt.Hunt, m Message) {
	var listeners []Listener
	_, err := datastore.NewQuery(listenerKind).Ancestor(h.Key).Filter("Open =", true).GetAll(c, &listeners)
	if err != nil {
		c.Errorf("Send: %v", err)
		return
	}
	for _, listener := range listeners {
		err := channel.SendJSON(c, listener.Channel, m)
		if err != nil {
			c.Errorf("Send(%s): %v", listener.Channel, err)
		}
	}
}

func sendTeam(c appengine.Context, h *hunt.Hunt, t *team.Team, m Message) {
	var listeners []Listener
	_, err := datastore.NewQuery(listenerKind).Ancestor(h.Key).Filter("TeamID =", t.ID).Filter("Open =", true).GetAll(c, &listeners)
	if err != nil {
		c.Errorf("Send: %v", err)
		return
	}
	for _, listener := range listeners {
		err := channel.SendJSON(c, listener.Channel, m)
		if err != nil {
			c.Errorf("Send(%s): %v", listener.Channel, err)
		}
	}

}

func sendAdmin(c appengine.Context, h *hunt.Hunt, m Message) {
	var listeners []Listener
	_, err := datastore.NewQuery(listenerKind).Ancestor(h.Key).Filter("Admin =", true).Filter("Open =", true).GetAll(c, &listeners)
	if err != nil {
		c.Errorf("Send: %v", err)
		return
	}
	for _, listener := range listeners {
		err := channel.SendJSON(c, listener.Channel, m)
		if err != nil {
			c.Errorf("Send(%s): %v", listener.Channel, err)
		}
	}

}

