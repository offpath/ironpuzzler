package broadcast

import (
	"fmt"
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
	leaderboardListenerKind = "LeaderboardListener"
	UpdatePuzzle MessageType = 0
	RefreshPage
)

type LeaderboardListener struct {
	Channel string
}

func AddListener(c appengine.Context, h *hunt.Hunt, t *team.Team) string {
	var channelName string
	if t != nil {
		channelName = fmt.Sprintf("%s.%d", t.Name, time.Now().UnixNano())
	} else {
		channelName = fmt.Sprintf("nil.%d", time.Now().UnixNano())
	}
	token, err := channel.Create(c, channelName)
	if err != nil {
		c.Errorf("Error: %v", err)
		return ""
	}
	listener := LeaderboardListener{channelName}
	_, err = datastore.Put(c, datastore.NewIncompleteKey(c, leaderboardListenerKind, h.Key), &listener)
	if err != nil {
		c.Errorf("Error: %v", err)
		return ""
	}
	return token
}

func RemoveListener(c appengine.Context, str string) {
	keys, err := datastore.NewQuery(leaderboardListenerKind).Filter("Channel =", str).KeysOnly().GetAll(c, nil)
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
	var listeners []LeaderboardListener
	_, err := datastore.NewQuery(leaderboardListenerKind).Ancestor(h.Key).GetAll(c, &listeners)
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
