package main

import (
	"flag"
	gh "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"

	"pos-proxy/config"
	_ "pos-proxy/db"
	// "pos-proxy/ej"
	"pos-proxy/handlers"
)

func init() {
	log.SetFlags(log.LstdFlags)
}

func main() {
	port := flag.String("port", "7000", "Port to listen on")
	server_crt := flag.String("server_crt", "server.crt", "Certificate path")
	server_key := flag.String("server_key", "server.key", "Certificate key path")
	file_path := flag.String("config", "/etc/cloudinn/pos_config.json", "Configuration for the POS proxy")
	flag.Parse()
	config.Load(*file_path)
	originsOk := gh.AllowedOrigins([]string{"*"})
	headersOk := gh.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	methodsOk := gh.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	r := mux.NewRouter()
	r.HandleFunc("/proxy/test", handlers.ProxyTest).Methods("GET")
	r.HandleFunc("/proxy/fdm/status/{rcrs}", handlers.FDMStatus).Methods("GET")
	r.HandleFunc("/proxy/fdm/invoices", handlers.SubmitInvoice).Methods("POST")
	r.HandleFunc("/proxy/fdm/folio", handlers.Folio).Methods("POST")
	r.HandleFunc("/proxy/fdm/payment", handlers.PayInvoice).Methods("POST")

	// fire a goroutine to send stored electronic journal records
	// to backend every 10 seconds
	/*go func() {
		for true {
			ej.PushToBackend()
			time.Sleep(time.Second * 10)
		}
	}()*/

	go func() {
		for true {
			config.FetchConfiguration()
			time.Sleep(time.Second * 10)
		}
	}()

	log.Printf("Listening on http://localhost:%s\n", *port)
	log.Fatal(http.ListenAndServeTLS(":"+*port, *server_crt, *server_key, gh.CORS(originsOk, headersOk, methodsOk)(r)))

}
