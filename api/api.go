// Package api handles all endpoints that return json.
package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"appengine"

	"hunt"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) != 4 {
		http.Error(w, "Not found", 404)
		return
	}

	c := appengine.NewContext(r)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)

	var h *hunt.Hunt
	if id := r.FormValue("hunt_id"); id != "" {
		h = hunt.ID(c, id)
	}

	// TODO(dneal): Get team id.

	var err error
	switch path[3] {
	case "hunt":
		// Easy case
		err = enc.Encode(h)
	case "hunts":
		hunts := hunt.All(c)
		err = enc.Encode(hunts)
	case "addhunt":
		hunt.New(c, r.FormValue("Name"), r.FormValue("path"))
	case "deletehunt":
		h.Delete(c)
	case "teams":

	case "addteam":

	case "deleteteam":

	}
	if err != nil {
		c.Errorf("Error: %v", err)
		return
	}
}

