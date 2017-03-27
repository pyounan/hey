package main

import (
	"flag"
	gh "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"

	_ "pos-proxy/db"
	"pos-proxy/ej"
	"pos-proxy/handlers"
)

func main() {
	port := flag.String("port", "7000", "Port to listen on")
	server_crt := flag.String("server_crt", "server.crt", "Certificate path")
	server_key := flag.String("server_key", "server.key", "Certificate key path")
	flag.Parse()
	originsOk := gh.AllowedOrigins([]string{"*"})
	headersOk := gh.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	methodsOk := gh.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	r := mux.NewRouter()
	r.HandleFunc("/proxy/fdm/status", handlers.FDMStatus).Methods("GET")
	r.HandleFunc("/proxy/fdm/invoices", handlers.SubmitInvoice).Methods("POST")
	r.HandleFunc("/proxy/fdm/folio", handlers.Folio).Methods("POST")
	r.HandleFunc("/proxy/fdm/payment", handlers.PayInvoice).Methods("POST")

	go func() {
		for true {
			ej.PushToBackend()
			time.Sleep(time.Second * 5)
		}
	}()

	log.Printf("Listening on http://localhost:%s\n", *port)
	log.Fatal(http.ListenAndServeTLS(":"+*port, *server_crt, *server_key, gh.CORS(originsOk, headersOk, methodsOk)(r)))

}
