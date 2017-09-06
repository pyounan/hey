package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	gh "github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"pos-proxy/auth"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/income"
	"pos-proxy/pos"
	"pos-proxy/proxy"
	"pos-proxy/syncer"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// read encryption key from environment variables
	key := os.Getenv("CLOUDINN_ENC_KEY")
	log.Println(key)
	err := config.ParseAuthCredentials(key)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	port := flag.String("port", "80", "Port to listen on")
	filePath := flag.String("config", "/etc/cloudinn/pos_config.json", "Configuration for the POS proxy")
	flag.Parse()
	config.Load(*filePath)
	headersOk := gh.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "X-CSRFToken", "Accept", "Accept-Lanuage", "Accept-Encoding", "Authorization"})
	originsOk := gh.AllowedOrigins([]string{"*"})
	methodsOk := gh.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	r := mux.NewRouter()
	r = r.StrictSlash(true)
	r.HandleFunc("/proxy/test/", proxy.Status).Methods("GET")

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
	r.HandleFunc("/shadowinn/api/auditdate/", income.GetAuditDate).Methods("GET")
	r.HandleFunc("/shadowinn/api/company/", income.GetCompany).Methods("GET")

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
	r.HandleFunc("/api/pos/terminal/{id}/", pos.GetTerminal).Methods("GET")
	r.HandleFunc("/api/pos/terminal/{id}/unlockterminal/", pos.UnlockTerminal).Methods("GET")

	r.HandleFunc("/api/pos/course/", pos.ListCourses).Methods("GET")

	r.HandleFunc("/api/pos/posinvoices/", pos.ListInvoicesPaginated).Methods("GET").Queries("is_settled", "")
	r.HandleFunc("/api/pos/posinvoices/", pos.ListInvoices).Methods("GET")
	r.HandleFunc("/api/pos/posinvoices/", pos.SubmitInvoice).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/folio/", pos.FolioInvoice).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/changetable/", pos.ChangeTable).Methods("PUT")
	r.HandleFunc("/api/pos/posinvoices/split/", pos.SplitInvoices).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/togglelocking/", pos.ToggleLocking).Methods("GET")
	r.HandleFunc("/api/pos/posinvoicelineitems/wasteandvoid/", pos.WasteAndVoid).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/{invoice_nubmer}/", pos.GetInvoice).Methods("GET")
	r.HandleFunc("/api/pos/posinvoices/{invoice_nubmer}/", pos.UpdateInvoice).Methods("PUT")
	r.HandleFunc("/api/pos/posinvoices/{invoice_nubmer}/createpostings/", pos.PayInvoice).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/{invoice_number}/unlock/", pos.UnlockInvoice).Methods("GET")

	r.HandleFunc("/api/pos/condiment/", pos.ListCondiments).Methods("GET")

	r.HandleFunc("/api/pos/fixeddiscount/", pos.ListFixedDiscounts).Methods("GET")

	r.HandleFunc("/api/pos/fdm/", pos.IsFDMEnabled).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(proxy.ProxyToBackend)

	go func() {
		for true {
			syncer.FetchConfiguration()
			time.Sleep(time.Second * 60)
		}
	}()

	go func() {
		for true {
			syncer.PushToBackend()
			time.Sleep(time.Second * 20)
		}
	}()

	go func() {
		for true {
			syncer.Load()
			time.Sleep(time.Second * 60)
		}
	}()


	lr := gh.LoggingHandler(os.Stdout, r)
	// r = gh.RecoveryHandler()(lr)

	db.ConnectRedis()

	log.Printf("Listening on http://localhost:%s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port,
		gh.CORS(originsOk, headersOk, methodsOk)(lr)))

}
