package ironpuzzler

import (
	"html/template"
	"net/http"
	"strings"

	"appengine"

	"api"
	"hunt"
)

var (
	adminTemplate = template.Must(template.New("admin.html").Delims("{(", ")}").ParseFiles("templates/admin.html"))
	adminHuntTemplate = template.Must(template.New("admin_hunt.html").Delims("{(", ")}").ParseFiles("templates/admin_hunt.html"))
)

func init() {
	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/admin/", adminHuntHandler)
	http.HandleFunc("/admin/api/", api.AdminHandler)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := adminTemplate.Execute(w, nil)
	if err != nil {
		c.Errorf("adminTemplate: %v", err)
	}
}

func adminHuntHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	path := strings.Split(r.URL.Path, "/")
	if len(path) != 3 {
		http.Error(w, "Not found", 404)
		return
	}
	
	h := hunt.Path(c, path[2])
	if h == nil {
		http.Error(w, "Not found", 404)
		return
	}
	
	err := adminHuntTemplate.Execute(w, h.ID)
	if err != nil {
		c.Errorf("adminHuntTemplate: %v", err)
	}
}

