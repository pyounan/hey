package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

var port = flag.String("port", "", "Port to listen on")
var running = false

func main() {
	flag.Parse()
	if *port != "" {
		// do stuff
		log.Printf("listening on port %s", *port)
		running = true
	} else {
		log.Println("Please provide a port number")
	}
	srv := &http.Server{
		Addr:           ":" + *port,
		Handler:        http.HandlerFunc(handleReq),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	for {
		log.Println(srv.ListenAndServe())
	}
}

func handleReq(w http.ResponseWriter, r *http.Request) {
	log.Println("recieved a connection from ", r.RemoteAddr)
	if r.Method == "GET" {
		log.Println(r.Header)
	} else if r.Method == "POST" {
		log.Println(r.Body)
	}
}
