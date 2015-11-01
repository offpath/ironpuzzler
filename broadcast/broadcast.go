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

type MessageType int

const (
	listenerKind = "Listener"
	UpdatePuzzle MessageType = 0
	RefreshPage
	unauthListener = "unauth"
	adminListener = "admin"
)

type Listener struct {
	Channel string
	TeamID string
	Admin bool
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

func SendUpdatePuzzle(c appengine.Context, h *hunt.Hunt, p *puzzle.Puzzle) {
	send(c, h, fmt.Sprintf("%s", p.ID))
}

func SendRefresh(c appengine.Context, h *hunt.Hunt) {
	send(c, h, "refresh")
}

func send(c appengine.Context, h *hunt.Hunt, str string) {
	var listeners []Listener
	_, err := datastore.NewQuery(listenerKind).Ancestor(h.Key).GetAll(c, &listeners)
	if err != nil {
		c.Errorf("Send: %v", err)
		return
	}
	for _, listener := range listeners {
		err := channel.Send(c, listener.Channel, str)
		if err != nil {
			c.Errorf("Send(%s): %v", listener.Channel, err)
		}
	}
}

func SendConsoleUpdate(c appengine.Context, h *hunt.Hunt, str string) {

}
