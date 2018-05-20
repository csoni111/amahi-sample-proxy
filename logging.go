package main

import (
	"log"
	"time"
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"os"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/disk"
	"runtime"
)

// Logging client structure.
type LogClient struct {
	channel chan *ConnectionLog
	db      *sql.DB
}

// ConnectionEvent for connection log
type ConnectionEvent int

const (
	Connection    ConnectionEvent = 1
	Disconnection ConnectionEvent = 2
	Streaming     ConnectionEvent = 3
)

// ConnectionLog structure.
type ConnectionLog struct {
	Timestamp int64           `json:"timestamp"`
	Event     ConnectionEvent `json:"event"`
	FSInfo    *FSInfo         `json:"fs_info"`
	token     string
}

// System Stats structure.
type Stat struct {
	Timestamp int64  `json:"timestamp"`
	RamFree   uint64 `json:"ram_free"`
	DiskFree  uint64 `json:"disk_free"`
	MemAlloc  uint64 `json:"mem_alloc"`
}

// New LogClient.
func NewLogClient(dbPath string) (*LogClient, error) {
	db, err := initDb(dbPath)
	if err == nil {
		return &LogClient{
			db:      db,
			channel: make(chan *ConnectionLog),
		}, nil
	} else {
		log.Fatal(err)
		return &LogClient{}, err
	}
}

func initDb(dbPath string) (*sql.DB, error) {
	// open a new db connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// check sqlite db file exists, if not then create
	_, err = os.Open(dbPath)
	if err != nil {
		fsTable := `CREATE TABLE fs (id INTEGER NOT NULL PRIMARY KEY, token VARCHAR(50) NOT NULL UNIQUE, 
					version NUMERIC, local_addr VARCHAR(30), relay_addr VARCHAR(30), arch VARCHAR(30));`
		_, err = db.Exec(fsTable)
		if err != nil {
			return nil, err
		}

		connTable := `CREATE TABLE conn_log (id INTEGER NOT NULL PRIMARY KEY, timestamp INTEGER NOT NULL, 
					  event SMALLINT NOT NULL, fs_id INTEGER, FOREIGN KEY (fs_id) REFERENCES fs(id));`
		_, err = db.Exec(connTable)
		if err != nil {
			return nil, err
		}

		statsTable := `CREATE TABLE stats (id INTEGER NOT NULL PRIMARY KEY, timestamp INTEGER NOT NULL, 
					   ram_free BIGINT, disk_free BIGINT, mem_alloc BIGINT);`
		_, err = db.Exec(statsTable)
		if err != nil {
			return nil, err
		}
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Start log listener.
func (l *LogClient) Start() {
	for {
		c, more := <-l.channel
		if more {
			l.send(c)
		} else {
			return
		}
	}
}

func (l *LogClient) StatsMonitor() {
	for {
		vMemStat, _ := mem.VirtualMemory()
		usageStat, _ := disk.Usage("/")
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		_, err := l.db.Exec("INSERT INTO stats(timestamp, ram_free, disk_free, mem_alloc) VALUES(?, ?, ?, ?)",
			time.Now().Unix(), vMemStat.Available, usageStat.Free, ms.Alloc)
		if err != nil {
			return
		}
		time.Sleep(1 * time.Minute)
	}
}

// Stop log listener.
func (l *LogClient) Stop() {
	close(l.channel)
	l.db.Close()
}

func (l *LogClient) Log(connEvent ConnectionEvent, fsInfo *FSInfo, token string) {
	l.channel <- &ConnectionLog{
		Event:     connEvent,
		Timestamp: time.Now().Unix(),
		FSInfo:    fsInfo,
		token:     token,
	}
}

func (l *LogClient) send(c *ConnectionLog) {
	var fsId int64

	// query the db for fs given the session token
	err := l.db.QueryRow("SELECT id FROM fs WHERE token = ?", c.token).Scan(&fsId)
	if err != nil {
		// add the new fs to db
		res, err := l.db.Exec("INSERT INTO fs(token, version, local_addr, relay_addr, arch) VALUES (?, ?, ?, ?, ?)",
			c.token, c.FSInfo.Version, c.FSInfo.LocalAddr, c.FSInfo.RelayAddr, c.FSInfo.Arch)
		if err != nil {
			log.Fatal(err)
			return
		}
		fsId, err = res.LastInsertId()
	}

	// insert the connection log entry to the db
	_, err = l.db.Exec("INSERT INTO conn_log(timestamp, event, fs_id) VALUES(?, ?, ?)",
		c.Timestamp, c.Event, fsId)
	if err != nil {
		log.Fatal(err)
	}
}
