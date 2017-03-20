package main

import (
	"flag"
	gh "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"

	_ "pos-proxy/db"
	"pos-proxy/handlers"
)

func main() {
	port := flag.String("port", "7000", "Port to listen on")
	flag.Parse()
	originsOk := gh.AllowedOrigins([]string{"*"})
	headersOk := gh.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	methodsOk := gh.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	r := mux.NewRouter()
	r.HandleFunc("/proxy/fdm/status", handlers.FDMStatus).Methods("GET")
	r.HandleFunc("/proxy/fdm/invoices", handlers.SubmitInvoice).Methods("POST")
	r.HandleFunc("/proxy/fdm/folio", handlers.Folio).Methods("POST")
	r.HandleFunc("/proxy/fdm/payment", handlers.PayInvoice).Methods("POST")
	log.Printf("Listening on http://localhost:%s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, gh.CORS(originsOk, headersOk, methodsOk)(r)))
}
