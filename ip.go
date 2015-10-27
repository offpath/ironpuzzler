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
)

func init() {
	http.HandleFunc("/", huntHandler)
	http.HandleFunc("/api/", api.HuntHandler)
	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/admin/", adminHuntHandler)
	http.HandleFunc("/admin/api/", api.AdminHandler)
	http.HandleFunc("/includes/ingredients.html", ingredientsHandler)
	http.HandleFunc("/_ah/channel/disconnected/", channelDisconnectHandler)
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

func adminHuntHandler(w http.ResponseWriter, r *http.Request) {
	pathHandler(w, r, adminHuntTemplate)
}

