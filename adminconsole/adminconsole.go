package adminconsole

import (
	"fmt"
	"time"

	"appengine"
	"appengine/datastore"

	"broadcast"
	"hunt"
)

const (
	messageKind = "Message"
)

type Message struct {
	Contents string
	Time time.Time
}

func (m *Message) ToString(c appengine.Context, loc *time.Location) string {
	return fmt.Sprintf("%s: %s", m.Time.In(loc).Format("15:04:05"), m.Contents)
}

func Logs(c appengine.Context, h *hunt.Hunt) []string {
	var messages []*Message
	_, err := datastore.NewQuery(messageKind).Ancestor(h.Key).Order("-Time").Limit(1000).GetAll(c, &messages)
	if err != nil {
		c.Errorf("Error: %v", err)
		return nil
	}
	var result []string
	for _, m := range messages {
		result = append(result, m.ToString(c, h.GetTimezone(c)))
	}
	return result
}

func Log(c appengine.Context, h *hunt.Hunt, message string) {
	m := &Message{message, time.Now()}
	k := datastore.NewIncompleteKey(c, messageKind, h.Key)
	k, err := datastore.Put(c, k, m)
	if err != nil {
		c.Errorf("Error: %v", err)
	}
	broadcast.SendConsoleUpdate(c, h, m.ToString(c, h.GetTimezone(c)))
}
