package main

import (
	"flag"
	//"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"

	_ "pos-proxy/db"
	"pos-proxy/handlers"
)

func main() {
	port := flag.String("port", "7000", "Port to listen on")
	flag.Parse()
	r := mux.NewRouter()
	r.HandleFunc("/fdm/status", handlers.FDMStatus).Methods("GET")
	log.Printf("Listening on http://localhost:%s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, r))
}
