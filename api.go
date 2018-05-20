package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"encoding/json"
	"database/sql"
	"time"
	"strconv"
)

func (d *Dashboard) InitApiRouter(router *mux.Router) {
	//router.HandleFunc("/fs/", d.fsJson).Methods("GET")
	router.HandleFunc("/connections/", d.connJson).Methods("GET")
}

func (d *Dashboard) getFSInfo(db *sql.DB, fsId int64) (fs FSInfo) {
	err := db.QueryRow("SELECT version, local_addr, relay_addr, arch FROM fs Where id = ?",
		fsId).Scan(&fs.Version, &fs.LocalAddr, &fs.RelayAddr, &fs.Arch)
	if err != nil {
		handle(err)
	}
	return
}

func (d *Dashboard) connJson(w http.ResponseWriter, r *http.Request) {
	db := d.getDb()
	defer db.Close()
	var t int64 = 0
	vals, exist := r.URL.Query()["t"]
	if !exist || len(vals) < 1 {
		t = time.Now().Unix() - (6 * 60 * 60)
	} else {
		val := vals[0]
		if val != "all" {
			i, _ := strconv.Atoi(val)
			t = time.Now().Unix() - (int64(i) * 60 * 60)
		}
	}
	var rows *sql.Rows
	var err error
	if t == 0 {
		rows, err = db.Query("SELECT timestamp, event, fs_id FROM conn_log")
	} else {
		rows, err = db.Query("SELECT timestamp, event, fs_id FROM conn_log WHERE timestamp > ?", t)
	}
	if err != nil {
		handle(err)
	}
	defer rows.Close()
	connections := make([]ConnectionLog, 0)
	allFS := make(map[int64]FSInfo)
	for rows.Next() {
		var conn ConnectionLog
		var fsId int64
		err = rows.Scan(&conn.Timestamp, &conn.Event, &fsId)
		if err != nil {
			handle(err)
		}
		if fsInfo, exist := allFS[fsId]; exist {
			conn.FSInfo = &fsInfo
		} else {
			fsInfo = d.getFSInfo(db, fsId)
			allFS[fsId] = fsInfo
			conn.FSInfo = &fsInfo
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

