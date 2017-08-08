package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	gh "github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"pos-proxy/auth"
	"pos-proxy/config"
	_ "pos-proxy/db"
	"pos-proxy/income"
	"pos-proxy/pos"
	"pos-proxy/proxy"
)

func init() {
	log.SetFlags(log.LstdFlags)
}

func main() {
	port := flag.String("port", "80", "Port to listen on")
	filePath := flag.String("config", "/etc/cloudinn/pos_config.json", "Configuration for the POS proxy")
	flag.Parse()
	config.Load(*filePath)
	headersOk := gh.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "X-CSRFToken", "Accept", "Accept-Lanuage", "Accept-Encoding"})
	originsOk := gh.AllowedOrigins([]string{"*"})
	methodsOk := gh.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	r := mux.NewRouter()
	r = r.StrictSlash(true)
	r.HandleFunc("/proxy/status/", proxy.Status).Methods("GET")

	// auth
	r.HandleFunc("/api/ensure_tenant_selected/", auth.EnsureTenantSelected).Methods("GET")

	// handle FDM requests
	// r.HandleFunc("/proxy/fdm/status/{rcrs}", handlers.FDMStatus).Methods("GET")
	// r.HandleFunc("/proxy/fdm/invoices", handlers.SubmitInvoice).Methods("POST")
	// r.HandleFunc("/proxy/fdm/folio", handlers.Folio).Methods("POST")
	// r.HandleFunc("/proxy/fdm/payment", handlers.PayInvoice).Methods("POST")

	// handle INCOME requests
	r.HandleFunc("/income/api/currency/", income.ListCurrencies).Methods("GET")
	r.HandleFunc("/income/api/currency/{id}/", income.GetCurrency).Methods("GET")

	r.HandleFunc("/income/api/department/", income.ListDepartments).Methods("GET")
	r.HandleFunc("/income/api/department/{id}/", income.GetDepartment).Methods("GET")

	r.HandleFunc("/income/api/cashier/getposcashier/", income.GetPosCashier).Methods("GET")
	r.HandleFunc("/income/api/poscashierpermissions/", income.GetCashierPermissions).Methods("GET")

	// handle POS requests
	r.HandleFunc("/api/pos/course/{id}/", pos.ListCourses).Methods("GET")

	r.HandleFunc("/api/pos/store/", pos.ListStores).Methods("GET")
	r.HandleFunc("/api/pos/store/{id}/", pos.GetStore).Methods("GET")
	r.HandleFunc("/api/pos/storedetails/{id}/", pos.GetStoreDetails).Methods("GET")
	r.HandleFunc("/api/pos/store/{id}/", pos.UpdateStore).Methods("PUT")

	r.HandleFunc("/api/pos/tables/", pos.ListTables).Methods("GET")
	r.HandleFunc("/api/pos/tables/{id}/getlatestchanges/", pos.GetTableLatestChanges).Methods("POST")
	r.HandleFunc("/api/pos/tables/{number}/", pos.GetTable).Methods("GET")
	r.HandleFunc("/api/pos/tables/{id}/", pos.UpdateTable).Methods("PUT")

	r.HandleFunc("/api/pos/printer/", pos.ListPrinters).Methods("GET")
	r.HandleFunc("/api/pos/printersettings/", pos.ListPrinterSettings).Methods("GET")

	r.HandleFunc("/api/pos/terminal/", pos.ListTerminals).Methods("GET")

	r.HandleFunc("/api/pos/courses/", pos.ListCourses).Methods("GET")

	r.HandleFunc("/api/pos/posinvoices/", pos.ListInvoices).Methods("GET")
	r.HandleFunc("/api/pos/posinvoices/", pos.SubmitInvoice).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/{invoice_nubmer}/", pos.GetInvoice).Methods("GET")
	r.HandleFunc("/api/pos/posinvoices/{invoice_nubmer}/", pos.UpdateInvoice).Methods("PUT")
	r.HandleFunc("/api/pos/posinvoices/{invoice_nubmer}/folio/", pos.FolioInvoice).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/{invoice_nubmer}/payment/", pos.PayInvoice).Methods("POST")

	r.HandleFunc("/api/pos/course/", pos.ListCourses).Methods("GET")

	// // fire a goroutine to send stored electronic journal records
	// // to backend every 10 seconds
	// go func() {
	// 	for true {
	// 		ej.PushToBackend()
	// 		time.Sleep(time.Second * 10)
	// 	}
	// }()

	// go func() {
	// 	for true {
	// 		config.FetchConfiguration()
	// 		time.Sleep(time.Second * 10)
	// 	}
	// }()

	// go func() {
	// 	for true {
	// 		syncer.Load()
	// 		time.Sleep(time.Second * 600)
	// 	}
	// }()

	lr := gh.LoggingHandler(os.Stdout, r)
	// r = gh.RecoveryHandler()(lr)

	log.Printf("Listening on http://localhost:%s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port,
		gh.CORS(originsOk, headersOk, methodsOk)(lr)))

}
