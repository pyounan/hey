package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	gh "github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"pos-proxy/auth"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/income"
	"pos-proxy/libs/libfdm"
	"pos-proxy/opera"
	"pos-proxy/pos"
	"pos-proxy/pos/fdm"
	"pos-proxy/proxy"
	"pos-proxy/sun"
	"pos-proxy/syncer"
	"pos-proxy/templateexport"

	"github.com/TV4/graceful"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	// Check command line arguments, if askings for version, print version then exit
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(config.Version)
		os.Exit(0)
	}
	if len(os.Args) > 1 && os.Args[1] == "versions" {
		fmt.Println(fmt.Sprintf("Version: %s", config.Version))
		fmt.Println(fmt.Sprintf("Build number: %s", config.BuildNumber))
		fmt.Println(fmt.Sprintf("Virtual host: %s", config.VirtualHost))
		os.Exit(0)
	}
	// read encryption key from environment variables
	key := os.Getenv("CLOUDINN_ENC_KEY")
	err := config.ParseAuthCredentials(key)
	if err != nil {
		log.Println(err)
	}
	port := flag.String("port", "80", "Port to listen on")
	templatesPath := flag.String("templates", "templates/*", "Path of templates directory")
	filePath := flag.String("config", "/etc/cloudinn/pos_config.json", "Configuration for the POS proxy")
	flag.Parse()
	err = config.Load(*filePath)
	if err != nil {
		log.Fatal(err)
	}
	// Load templates
	templateexport.ParseTemplates(*templatesPath)
	// Connect to Database
	db.Connect()

	fmt.Println("Loading data & configuration from Cloudinn servers...")
	if config.Config.IsFDMEnabled {
		syncer.PullSequences()
	}
	syncer.Load(syncer.SingleLoadApis)

	go func() {
		for true {
			syncer.FetchConfiguration()
			time.Sleep(time.Second * 60)
		}
	}()

	go func() {
		for true {
			syncer.PushToBackend()
			time.Sleep(time.Second * 5)
		}
	}()

	go func() {
		for true {
			syncer.Load(syncer.ConfApis)
			time.Sleep(time.Second * 30)
		}
	}()

	go proxy.CheckForupdates()

	err = db.ConnectRedis()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection to redis has been stablished successfully...")

	if config.Config.IsOperaEnabled {
		opera.Connect()
	}

	handler := createRouter()

	graceful.Timeout = 30 * time.Second
	graceful.LogListenAndServe(
		&http.Server{
			Addr:    ":" + *port,
			Handler: handler,
		},
	)

}

func createRouter() http.Handler {
	headersOk := gh.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "X-CSRFToken", "Accept", "Accept-Lanuage", "Accept-Encoding", "Authorization"})
	originsOk := gh.AllowedOrigins([]string{"*"})
	methodsOk := gh.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	r := mux.NewRouter()
	r = r.StrictSlash(true)

	// Define routes
	r.HandleFunc("/proxy/test/", proxy.Status).Methods("GET")
	r.HandleFunc("/proxy/version/", proxy.Version).Methods("GET")

	// auth
	r.HandleFunc("/api/ensure_tenant_selected/", auth.EnsureTenantSelected).Methods("GET")
	r.HandleFunc("/core/getloggedinusergroups/{id}/", auth.GetUserPermissions).Methods("GET")

	// handle FDM requests
	r.HandleFunc("/proxy/fdms/status/{rcrs}/", fDMStatus).Methods("GET")
	r.HandleFunc("/api/fdms/pins/", pos.FDMSetPin).Methods("POST")

	// handle INCOME requests
	r.HandleFunc("/income/api/currency/", income.ListCurrencies).Methods("GET")
	r.HandleFunc("/income/api/currency/{id}/", income.GetCurrency).Methods("GET")

	r.HandleFunc("/income/api/department/", income.ListDepartments).Methods("GET")
	r.HandleFunc("/income/api/department/{id}/", income.GetDepartment).Methods("GET")

	r.HandleFunc("/income/api/cashier/getposcashier/", income.GetPosCashier).Methods("POST")
	r.HandleFunc("/income/api/cashier/clockout/", income.Clockout).Methods("POST")
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
	r.HandleFunc("/api/pos/terminal/{id}/", pos.UpdateTerminal).Methods("PUT")
	r.HandleFunc("/api/pos/terminal/{id}/unlockterminal/", pos.UnlockTerminal).Methods("POST")

	r.HandleFunc("/api/pos/course/", pos.ListCourses).Methods("GET")

	r.HandleFunc("/api/pos/posinvoices/", pos.ListInvoicesPaginated).Methods("GET").Queries("is_settled", "")
	r.HandleFunc("/api/pos/posinvoices/", pos.ListInvoicesLite).Methods("GET").Queries("simplified", "")
	r.HandleFunc("/api/pos/posinvoices/", pos.ListInvoices).Methods("GET")
	r.HandleFunc("/api/pos/posinvoices/", pos.SubmitInvoice).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/bulksubmit/", pos.BulkSubmitInvoices).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/folio/", pos.FolioInvoice).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/refund/", pos.RefundInvoice).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/houseuse/", pos.Houseuse).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/changetable/", pos.ChangeTable).Methods("PUT")
	r.HandleFunc("/api/pos/posinvoices/split/", pos.SplitInvoices).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/createpaymentej/", pos.CreatePaymentEJ).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/togglelocking/", pos.ToggleLocking).Methods("GET")
	r.HandleFunc("/api/pos/posinvoicelineitems/wasteandvoid/", pos.WasteAndVoid).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/{invoice_number}/", pos.GetInvoice).Methods("GET")
	r.HandleFunc("/api/pos/posinvoices/{invoice_number}/createpostings/", pos.PayInvoice).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/{invoice_number}/cancelpostings/", pos.CancelPostings).Methods("POST")
	r.HandleFunc("/api/pos/posinvoices/{invoice_number}/unlock/", pos.UnlockInvoice).Methods("GET")
	r.HandleFunc("/api/pos/posinvoices/{invoice_number}/getlatestchanges/", pos.GetInvoiceLatestChanges).Methods("POST")
	// fixed discounts
	r.HandleFunc("/api/pos/fixeddiscount/", pos.ListFixedDiscounts).Methods("GET")
	r.HandleFunc("/api/pos/fixeddiscount/", pos.CreateFixedDiscount).Methods("POST")
	r.HandleFunc("/api/pos/fixeddiscount/{id}/", pos.GetFixedDiscount).Methods("GET")
	r.HandleFunc("/api/pos/fixeddiscount/{id}/", pos.UpdateFixedDiscount).Methods("PUT")
	r.HandleFunc("/api/pos/fixeddiscount/{id}/", pos.DeleteFixedDiscount).Methods("DELETE")

	r.HandleFunc("/api/pos/condiment/", pos.ListCondiments).Methods("GET")

	// HTML Views
	r.HandleFunc("/", homeView).Methods("GET")
	r.HandleFunc("/syncer/logs", requestsLogView).Methods("GET")
	r.HandleFunc("/syncer/logs/request/{id}", syncerRequest).Methods("GET")
	r.HandleFunc("/syncer/logs/response/{id}", syncerResponse).Methods("GET")
	r.HandleFunc("/api/opera/rooms/", opera.ListOperaRooms).Methods("GET")
	r.HandleFunc("/api/opera/roomdepartment/", opera.GetRoomDepartment).Methods("GET")
	r.HandleFunc("/api/pos/opera/{id}/", opera.DeleteConfig).Methods("DELETE")
	r.HandleFunc("/jv/", sun.ImportJournalVouchers).Methods("GET", "POST")

	//r.HandleFunc("/api/pos/fdm/", pos.IsFDMEnabled).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(proxy.ProxyToBackend)

	//lr := gh.LoggingHandler(os.Stdout, r)
	mr := proxy.StatusMiddleware(r)
	cors := gh.CORS(originsOk, headersOk, methodsOk)(mr)
	return cors
}

func homeView(w http.ResponseWriter, r *http.Request) {
	ctx := bson.M{}
	ctx["version"] = config.Version
	templateexport.ExportedTemplates.ExecuteTemplate(w, "home", ctx)
}

func requestsLogView(w http.ResponseWriter, r *http.Request) {
	offset := 0
	limit := 100
	qParams := r.URL.Query()
	page := 0
	if v, ok := qParams["page"]; ok {
		page, _ = strconv.Atoi(v[0])
		offset = (limit * page)
	}
	ctx := bson.M{}
	logs := []syncer.RequestLog{}
	recordCount, _ := db.DB.C("requests_log").With(db.Session.Copy()).Count()
	db.DB.C("requests_log").With(db.Session.Copy()).Find(nil).Sort("-created_at").Limit(limit).Skip(offset).All(&logs)
	ctx["logs"] = logs
	if offset+limit >= recordCount {
		ctx["hasNext"] = false
	} else {
		ctx["hasNext"] = true
		ctx["nextPage"] = page + 1
	}
	if page == 0 {
		ctx["hasPrevious"] = false
	} else {
		ctx["hasPrevious"] = true
		ctx["prevPage"] = page - 1
	}
	ctx["page"] = page + 1
	ctx["totalRecords"] = recordCount
	ctx["offset"] = offset + 1
	ctx["lastRecord"] = offset + limit
	if recordCount < offset+limit {
		ctx["lastRecord"] = recordCount
	}

	templateexport.ExportedTemplates.ExecuteTemplate(w, "syncer_logs", ctx)
}

func syncerRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	logRecord := syncer.RequestLog{}
	err := db.DB.C("requests_log").With(db.Session.Copy()).FindId(bson.ObjectIdHex(id)).One(&logRecord)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	helpers.ReturnSuccessMessage(w, logRecord.Request)
}

func syncerResponse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	logRecord := syncer.RequestLog{}
	err := db.DB.C("requests_log").With(db.Session.Copy()).FindId(bson.ObjectIdHex(id)).One(&logRecord)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	helpers.ReturnSuccessMessage(w, logRecord.ResponseBody)
}

func fDMStatus(w http.ResponseWriter, r *http.Request) {
	rcrs := mux.Vars(r)["rcrs"]
	ctx := bson.M{
		"has_error": false,
		"error":     "",
		"RCRS":      rcrs,
		"version":   config.Version,
	}
	// create FDM connection
	conn, err := fdm.Connect(rcrs)
	if err != nil {
		ctx["error"] = err.Error()
		ctx["has_error"] = true
		templateexport.ExportedTemplates.ExecuteTemplate(w, "fdm_status", ctx)
		return
	}
	defer conn.Close()
	// send status message to FDM
	ns, err := fdm.GetNextSequence(rcrs)
	if err != nil {
		ctx["error"] = err.Error()
		ctx["has_error"] = true
		templateexport.ExportedTemplates.ExecuteTemplate(w, "fdm_status", ctx)
		return
	}
	resp, err := libfdm.Identification(conn, ns)
	if err != nil {
		ctx["error"] = err.Error()
		ctx["has_error"] = true
		templateexport.ExportedTemplates.ExecuteTemplate(w, "fdm_status", ctx)
		return
	}
	ctx["response"] = resp
	templateexport.ExportedTemplates.ExecuteTemplate(w, "fdm_status", ctx)
}
