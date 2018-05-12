package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"html/template"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"encoding/json"
)

type Dashboard struct {
	dbPath *string
}

func (d *Dashboard) InitDashboardRouter(router *mux.Router) {
	router.HandleFunc("/", d.dashboardHandler).Methods("GET")
	router.HandleFunc("/fs.json", d.fsJson).Methods("GET")
	router.HandleFunc("/connections.json", d.connJson).Methods("GET")
}

func (d *Dashboard) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("dashboard/home.html")
	t.Execute(w, nil)
}

func (d *Dashboard) fsJson(w http.ResponseWriter, r *http.Request) {
	db := d.getDb()
	defer db.Close()
	rows, err := db.Query("SELECT version, local_addr, relay_addr, arch FROM fs")
	if err != nil {
		handle(err)
	}
	defer rows.Close()
	allFS := make([]FSInfo, 0)
	for rows.Next() {
		var fs FSInfo
		err = rows.Scan(&fs.Version, &fs.LocalAddr, &fs.RelayAddr, &fs.Arch)
		if err != nil {
			handle(err)
		}
		allFS = append(allFS, fs)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allFS)
}

func (d *Dashboard) connJson(w http.ResponseWriter, r *http.Request) {
	db := d.getDb()
	defer db.Close()
	rows, err := db.Query("SELECT timestamp, event FROM conn_log LIMIT 1000")
	if err != nil {
		handle(err)
	}
	defer rows.Close()
	connections := make([]ConnectionLog, 0)
	for rows.Next() {
		var conn ConnectionLog
		err = rows.Scan(&conn.Timestamp, &conn.Event)
		if err != nil {
			handle(err)
		}
		connections = append(connections, conn)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connections)
}

func (d *Dashboard) getDb() (db *sql.DB) {
	db, err := sql.Open("sqlite3", *d.dbPath)
	if err != nil {
		handle(err)
	}
	return
}
