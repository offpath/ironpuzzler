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
)

type Listener struct {
	Channel string
	TeamID string
	Admin bool
}

type Message struct {
	K string
	V string
}

func GetToken(c appengine.Context, h *hunt.Hunt, t *team.Team, admin bool) string {
	// Channel name format: huntID.teamID/admin/unauth.timestamp
	var channelName string
	middle := unauthListener
	if admin {
		middle = adminListener
	} else if t != nil {
		middle = t.ID
	}
	channelName = fmt.Sprintf("%s.%s.%d", h.ID, middle, time.Now().UnixNano())
	token, err := channel.Create(c, channelName)
	if err != nil {
		c.Errorf("Error: %v", err)
		return ""
	}
	return token
}

func AddListener(c appengine.Context, channelName string) {
	split := strings.Split(channelName, ".")
	if len(split) != 3 {
		c.Errorf("Unexpected channel name: %s", channelName)
		return
	}
	h := hunt.ID(c, split[0])
	if h == nil {
		c.Errorf("Channel without matching hunt: %s", channelName)
		return
	}
	listener := Listener{
		Channel: channelName,
	}
	if split[1] != unauthListener {
		if split[1] == adminListener {
			listener.Admin = true
		} else {
			t := team.ID(c, split[1])
			if t == nil {
				c.Errorf("Channel without matching team: %s", channelName)
				return
			}
			listener.TeamID = t.ID
		}
	}
	_, err := datastore.Put(c, datastore.NewIncompleteKey(c, listenerKind, h.Key), &listener)
	if err != nil {
		c.Errorf("Error: %v", err)
		return
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

func SendSurveyUpdate(c appengine.Context, h *hunt.Hunt) {
	sendAdmin(c, h, Message{K:surveyUpdate})
}

func SendTeamsUpdate(c appengine.Context, h *hunt.Hunt) {
	sendAdmin(c, h, Message{K:teamsUpdate})
}

func SendConsoleUpdate(c appengine.Context, h *hunt.Hunt, str string) {
	sendAdmin(c, h, Message{K:consoleUpdate,V:str})
}

func sendAll(c appengine.Context, h *hunt.Hunt, m Message) {
	var listeners []Listener
	_, err := datastore.NewQuery(listenerKind).Ancestor(h.Key).GetAll(c, &listeners)
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
	_, err := datastore.NewQuery(listenerKind).Ancestor(h.Key).Filter("TeamID =", t.ID).GetAll(c, &listeners)
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
	_, err := datastore.NewQuery(listenerKind).Ancestor(h.Key).Filter("Admin =", true).GetAll(c, &listeners)
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

