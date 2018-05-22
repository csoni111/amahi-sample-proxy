package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"html/template"
	_ "github.com/mattn/go-sqlite3"
)

type Dashboard struct {
	dbPath *string
	fileServers *map[string]*FS
}

func (d *Dashboard) InitDashboardRouter(router *mux.Router) {
	router.HandleFunc("/", d.pfeHandler).Methods("GET")
	router.HandleFunc("/fs/", d.fsHandler).Methods("GET")
}

func (d *Dashboard) pfeHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/dashboard/base.html", "templates/dashboard/pfe.html")
	t.ExecuteTemplate(w, "layout", nil)
}

func (d *Dashboard) fsHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/dashboard/base.html", "templates/dashboard/fs.html")
	t.ExecuteTemplate(w, "layout", nil)
}
