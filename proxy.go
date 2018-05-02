package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"golang.org/x/net/http2"
	"flag"
	"net"
	"github.com/gorilla/mux"
	"encoding/json"
	"net/url"
	"time"
	"context"
)

const HostAddr = ":8000"

// placeholder for proper error handling.
func handle(err error) {
	if err != nil {
		panic(err)
	}
}

type Proxy struct {
	fileServers map[string]FS
}

type FS struct {
	clientConn	*http2.ClientConn
	conn		net.Conn
	session		string
	fsInfo		FSInfo
}

func (p Proxy) pingFSPeriodically(fs FS) {
	const period = 59 * time.Second // time spent sleeping

	for {
		time.Sleep(period)
		if fs.clientConn.CanTakeNewRequest() {
			ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
			err := fs.clientConn.Ping(ctx)
			cancel()
			if err != nil {
				// ping to FS failed
				fmt.Println("FS disconnected")
				p.removeFS(fs)
				return
			}
		} else {
			// connection has closed
			fmt.Println("FS disconnected")
			p.removeFS(fs)
			return
		}
	}
}

type FSInfo struct {
	Version   string `json:"version"`
	LocalAddr string `json:"local_addr"`
	RelayAddr string `json:"relay_addr"`
	Arch      string `json:"arch"`
}

func (p *Proxy) RequestFromFS(w http.ResponseWriter, r *http.Request) error {
	fs, exist := p.getFS(r)

	if !exist {
		// session token is missing or no such FS is connected
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	u := r.URL
	if u.Host == "" {
		u.Host = HostAddr
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	if _, exist := r.Header["Connection"]; exist {
		r.Header.Del("Connection")
	}

	res, err := fs.clientConn.RoundTrip(r)
	handle(err)

	_, err = io.Copy(w, res.Body)
	handle(err)
	handle(res.Body.Close())

	return err
}

func (p *Proxy) getFS(r *http.Request) (fs FS, exist bool) {
	// read the session token from header
	token := r.Header.Get("Session")
	if token == "" {
		// check for token in url params
		token = r.FormValue("session")
		if token == "" {
			// No session token sent in the request
			return FS{}, false
		}
	}
	fs, exist = p.fileServers[token]
	return
}

func (p *Proxy) removeFS(fs FS) {
	if oldFs, exist := p.fileServers[fs.session]; exist && oldFs == fs {
		fs.conn.Close()
		delete(p.fileServers, fs.session)
		fmt.Println("FS removed")
	}
}

func (p *Proxy) ServeFS(w http.ResponseWriter, r *http.Request) {
	// read the session token
	token := r.Header.Get("Api-Key")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println("FS connected")

	// clean up the old connection if any
	if fs, exist := p.fileServers[token]; exist {
		fs.conn.Close()
		// clientConn.Close() not yet implemented -> https://github.com/golang/go/issues/17292
		delete(p.fileServers, token)
	}

	// Read in the request body.
	fsInfo := FSInfo{}
	handle(json.NewDecoder(r.Body).Decode(&fsInfo))
	handle(r.Body.Close())

	if u, err := url.Parse("//" + fsInfo.RelayAddr); err == nil {
		u.Scheme = "https"
		fsInfo.RelayAddr = u.String()
	}

	if u, err := url.Parse("//" + fsInfo.LocalAddr); err == nil {
		u.Scheme = "http"
		if u.Port() == "" {
			u.Host += ":4563"
		}
		fsInfo.LocalAddr = u.String()
	}

	// re-purpose the connection.
	conn, _, err := w.(http.Hijacker).Hijack()
	handle(err)

	// send the 200 to FS.
	res := &http.Response{
		Status:     "200 Connection Established",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	handle(res.Write(conn))

	// prepare for serving requests from Client.
	transport := new(http2.Transport)
	clientConn, err := transport.NewClientConn(conn)
	handle(err)

	fs := FS{clientConn, conn, token, fsInfo}
	p.fileServers[token] = fs
	fmt.Println("FS added")

	// ping the fs periodically in a goroutine to see if the connection is up
	go p.pingFSPeriodically(fs)
}

func (p *Proxy) ServeClient(w http.ResponseWriter, r *http.Request) {
	fs, exist := p.getFS(r)

	if !exist {
		// session token is missing or no such FS is connected
		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fsInfoJson, err := json.Marshal(fs.fsInfo)
	handle(err)
	w.Write(fsInfoJson)
}

func (p *Proxy) ServeProxyClient(w http.ResponseWriter, r *http.Request) {
	err := p.RequestFromFS(w, r)
	if err != nil {
		log.Println("Encountered an error serving API request:", err)
	}
}

func (p *Proxy) DebugURL(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "goroutines:  %d\n", runtime.NumGoroutine())
	fmt.Fprintf(w, "serving: %d\n", len(p.fileServers))
}

func main() {
	certFile := "cert.pem"
	keyFile := "cert.key"

	tls := flag.Bool("t", false, "enable TLS")
	flag.Parse()

	proxy := new(Proxy)
	proxy.fileServers = make(map[string]FS)

	server := new(http.Server)
	router := mux.NewRouter()
	router.HandleFunc("/fs", proxy.ServeFS).Methods("PUT")
	router.HandleFunc("/client", proxy.ServeClient).Methods("PUT", "GET")
	router.HandleFunc("/debug", proxy.DebugURL).Methods("GET")
	router.PathPrefix("/").HandlerFunc(proxy.ServeProxyClient)
	server.Handler = router

	server.Addr = HostAddr

	if *tls {
		fmt.Println("Serving on", HostAddr, "with TLS")
		handle(server.ListenAndServeTLS(certFile, keyFile))
	} else {
		fmt.Println("Serving on", HostAddr, "*without* TLS")
		handle(server.ListenAndServe())
	}
}
