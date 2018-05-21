package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"encoding/json"
	"database/sql"
	"time"
	"strconv"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/cpu"
	"log"
)

func (d *Dashboard) InitApiRouter(router *mux.Router) {
	router.HandleFunc("/connections/", d.connJson).Methods("GET")
	router.HandleFunc("/stats/", d.systemStatsJson).Methods("GET")
}

func (d *Dashboard) parseTimeParam(r *http.Request) (t int64) {
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
	return
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
	t := d.parseTimeParam(r)
	var rows *sql.Rows
	var err error
	if t == 0 {
		rows, err = db.Query("SELECT timestamp, event, fs_id FROM conn_log ORDER BY id DESC")
	} else {
		rows, err = db.Query("SELECT timestamp, event, fs_id FROM conn_log WHERE timestamp > ? ORDER BY id DESC", t)
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

type SystemStat struct {
	TotalMemory uint64 `json:"total_memory"`
	TotalDisk   uint64 `json:"total_disk"`
	NumCores    int32  `json:"num_cores"`
	ModelName   string `json:"model_name"`
	Stats       []Stat `json:"stats"`
}

func (d *Dashboard) systemStatsJson(w http.ResponseWriter, r *http.Request) {
	db := d.getDb()
	defer db.Close()
	t := d.parseTimeParam(r)
	var rows *sql.Rows
	var err error
	if t == 0 {
		rows, err = db.Query("SELECT timestamp, ram_free, disk_free, mem_alloc, cpu_usage FROM stats " +
			"ORDER BY id DESC")
	} else {
		rows, err = db.Query("SELECT timestamp, ram_free, disk_free, mem_alloc, cpu_usage FROM stats "+
			"WHERE timestamp > ? ORDER BY id DESC", t)
	}
	if err != nil {
		handle(err)
	}
	defer rows.Close()
	stats := make([]Stat, 0)
	for rows.Next() {
		var s Stat
		err = rows.Scan(&s.Timestamp, &s.RamFree, &s.DiskFree, &s.MemAlloc, &s.CpuUsage)
		if err != nil {
			handle(err)
		}
		stats = append(stats, s)
	}
	vMemStat, _ := mem.VirtualMemory()
	usageStat, _ := disk.Usage("/")
	infoStats, err := cpu.Info()
	var infoStat cpu.InfoStat
	if err != nil {
		log.Fatal(err)
		infoStat = cpu.InfoStat{}
	} else {
		infoStat = infoStats[0]
	}
	systemStats := SystemStat{
		TotalMemory: vMemStat.Total,
		TotalDisk:   usageStat.Total,
		NumCores:    infoStat.Cores,
		ModelName:   infoStat.ModelName,
		Stats:       stats,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(systemStats)
	return
}

func (d *Dashboard) getDb() (db *sql.DB) {
	db, err := sql.Open("sqlite3", *d.dbPath)
	if err != nil {
		handle(err)
	}
	return
}

