package ironpuzzler

import (
	"html/template"
	"net/http"
	"strings"

	"appengine"

	"api"
	"broadcast"
	"hunt"
)

var (
	adminTemplate = template.Must(template.New("admin.html").Delims("{(", ")}").ParseFiles("templates/admin.html"))
	adminHuntTemplate = template.Must(template.New("admin_hunt.html").Delims("{(", ")}").ParseFiles("templates/admin_hunt.html"))
	huntTemplate = template.Must(template.New("hunt.html").Delims("{(", ")}").ParseFiles("templates/hunt.html"))
	ingredientsTemplate = template.Must(template.New("ingredients.html").Delims("{(", ")}").ParseFiles("templates/ingredients.html"))
	teamsTemplate = template.Must(template.New("teams.html").Delims("{(", ")}").ParseFiles("templates/teams.html"))
	puzzlesTemplate = template.Must(template.New("puzzles.html").Delims("{(", ")}").ParseFiles("templates/puzzles.html"))
	leaderboardTemplate = template.Must(template.New("leaderboard.html").Delims("{(", ")}").ParseFiles("templates/leaderboard.html"))
	consoleTemplate = template.Must(template.New("console.html").Delims("{(", ")}").ParseFiles("templates/console.html"))
	stateTemplate = template.Must(template.New("state.html").Delims("{(", ")}").ParseFiles("templates/state.html"))
	signinTemplate = template.Must(template.New("signin.html").Delims("{(", ")}").ParseFiles("templates/signin.html"))
)

func init() {
	http.HandleFunc("/", huntHandler)
	http.HandleFunc("/api/", api.HuntHandler)
	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/admin/", adminHuntHandler)
	http.HandleFunc("/admin/api/", api.AdminHandler)
	http.HandleFunc("/includes/ingredients.html", ingredientsHandler)
	http.HandleFunc("/includes/teams.html", teamsHandler)
	http.HandleFunc("/includes/puzzles.html", puzzlesHandler)
	http.HandleFunc("/includes/leaderboard.html", leaderboardHandler)
	http.HandleFunc("/_ah/channel/connected/", channelConnectHandler)
	http.HandleFunc("/_ah/channel/disconnected/", channelDisconnectHandler)
	http.HandleFunc("/includes/console.html", consoleHandler)
	http.HandleFunc("/includes/state.html", stateHandler)
	http.HandleFunc("/includes/signin.html", signinHandler)
}

func channelConnectHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Errorf("client_id = [%s]", r.FormValue("from"))
	broadcast.AddListener(c, r.FormValue("from"))
}

func channelDisconnectHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	broadcast.RemoveListener(c, r.FormValue("from"))
}

func pathHandler(w http.ResponseWriter, r *http.Request, t *template.Template) {
	c := appengine.NewContext(r)
	path := strings.Split(r.URL.Path, "/")
	
	h := hunt.Path(c, path[len(path) - 1])
	if h == nil {
		http.Error(w, "Not found", 404)
		return
	}

	err := t.Execute(w, h.ID)
	if err != nil {
		c.Errorf("template: %v", err)
	}
}

func huntHandler(w http.ResponseWriter, r *http.Request) {
	pathHandler(w, r, huntTemplate)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := adminTemplate.Execute(w, nil)
	if err != nil {
		c.Errorf("adminTemplate: %v", err)
	}
}

func ingredientsHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := ingredientsTemplate.Execute(w, nil)
	if err != nil {
		c.Errorf("ingredientsTemplate: %v", err)
	}
}

func teamsHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := teamsTemplate.Execute(w, nil)
	if err != nil {
		c.Errorf("teamsTemplate: %v", err)
	}
}

func puzzlesHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := puzzlesTemplate.Execute(w, nil)
	if err != nil {
		c.Errorf("puzzlesTemplate: %v", err)
	}
}

func leaderboardHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := leaderboardTemplate.Execute(w, nil)
	if err != nil {
		c.Errorf("leaderboardTemplate: %v", err)
	}
}

func consoleHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := consoleTemplate.Execute(w, nil)
	if err != nil {
		c.Errorf("consoleTemplate: %v", err)
	}
}

func stateHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := stateTemplate.Execute(w, nil)
	if err != nil {
		c.Errorf("stateTemplate: %v", err)
	}
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := signinTemplate.Execute(w, nil)
	if err != nil {
		c.Errorf("signinTemplate: %v", err)
	}
}

func adminHuntHandler(w http.ResponseWriter, r *http.Request) {
	pathHandler(w, r, adminHuntTemplate)
}

