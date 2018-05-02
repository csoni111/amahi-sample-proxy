package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
	"fmt"
)

// Logging client structure.
type LogClient struct {
	url 	 string
	hostname string
	channel  chan *ConnectionLog
	client   *http.Client
}

// ConnectionEvent for connection log
type ConnectionEvent string

const  (
	Connection 	  ConnectionEvent = "connect"
	Disconnection ConnectionEvent = "disconnect"
	Streaming 	  ConnectionEvent = "stream"
)

// ConnectionLog structure.
type ConnectionLog struct {
	Timestamp time.Time  	  `json:"@timestamp"`
	Event     ConnectionEvent `json:"event"`
	FSInfo    *FSInfo    	  `json:"fs"`
	Token	  string		  `json:"token"`
}

// New LogClient.
func NewLogClient(index string, typ string) *LogClient {
	hostname, _ := os.Hostname()
	return &LogClient{
		url:  	  fmt.Sprintf("http://localhost:9200/%s/%s", index, typ),
		hostname: hostname,
		client:   &http.Client{Timeout: time.Duration(5) * time.Second},
		channel:  make(chan *ConnectionLog),
	}
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

// Stop log listener.
func (l *LogClient) Stop() {
	close(l.channel)
}

func (l *LogClient) Log(connEvent ConnectionEvent, fsInfo *FSInfo, token string) {
	l.channel <- &ConnectionLog{
		Event:   	connEvent,
		Timestamp:  time.Now(),
		FSInfo:     fsInfo,
		Token:		token,
	}
}

func (l *LogClient) send(c *ConnectionLog) {
	enc, _ := json.Marshal(c)

	resp, err := l.client.Post(l.url,"application/json", bytes.NewBuffer(enc))
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Printf("post request to url: %s failed with status: %d", l.url, resp.StatusCode)
	}
}
