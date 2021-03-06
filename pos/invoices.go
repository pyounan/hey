package pos

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	incomemodels "pos-proxy/income/models"
	"pos-proxy/libs/libfdm"
	"pos-proxy/opera"
	"pos-proxy/payment"
	"pos-proxy/pos/fdm"
	"pos-proxy/pos/locks"
	"pos-proxy/pos/models"
	"pos-proxy/proxy"
	"pos-proxy/syncer"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/copier"

	"github.com/gorilla/mux"
	"github.com/novalagung/golpal"

	"gopkg.in/mgo.v2/bson"
)

// ListInvoicesLite swagger:route GET /api/pos/posinvoices/ invoices listInvoicesLite
//
// List Invoices (simplified)
//
// serves a list of simplified invoices with basic
// information
//
// Parameters:
// + name: simplified
//   in: query
//   required: true
//
// Responses:
//   200: []invoiceLite
func ListInvoicesLite(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	for key, val := range r.URL.Query() {
		if key == "store" {
			num, err := strconv.Atoi(val[0])
			if err != nil {
				helpers.ReturnErrorMessage(w, err.Error())
				return
			}
			q[key] = num
		} else if key == "updated_on" {
			t, err := time.Parse("", val[0])
			if err != nil {
				helpers.ReturnErrorMessage(w, err.Error())
				return
			}
			q[key] = bson.M{"$gt": t}
		} else if key == "invoice_number" {
			q[key] = val[0]
		}
	}
	invoices := []models.InvoiceLite{}
	session := db.Session.Copy()
	defer session.Close()

	err := db.DB.C("posinvoices").With(session).Find(q).Sort("-created_on").All(&invoices)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	helpers.ReturnSuccessMessage(w, invoices)
}

// ListInvoices swagger:route GET /api/pos/posinvoices/ invoices listInvoices
//
// List Invoices (full)
//
// serves a list of full detailed invoices
//
// Responses:
//   200: []invoice
func ListInvoices(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	for key, val := range r.URL.Query() {
		if key == "store" {
			num, err := strconv.Atoi(val[0])
			if err != nil {
				helpers.ReturnErrorMessage(w, err.Error())
				return
			}
			q[key] = num
		} else if key == "updated_on" {
			t, err := time.Parse("", val[0])
			if err != nil {
				helpers.ReturnErrorMessage(w, err.Error())
				return
			}
			q[key] = bson.M{"$gt": t}
		} else if key == "invoice_number" {
			q[key] = val[0]
		}
	}
	invoices := []models.Invoice{}
	session := db.Session.Copy()
	defer session.Close()

	err := db.DB.C("posinvoices").With(session).Find(q).Sort("-created_on").All(&invoices)
	if err != nil || len(invoices) == 0 {
		// if invoice is settled, get it from the backend & save it to mongo
		if _, ok := q["invoice_number"]; ok {
			netClient := helpers.NewNetClient()

			uri := fmt.Sprintf("%s%s", config.Config.BackendURI, r.RequestURI)
			req, err := http.NewRequest(r.Method, uri, r.Body)
			req = helpers.PrepareRequestHeaders(req)
			resp, err := netClient.Do(req)
			if err != nil {
				log.Println(err.Error())
				helpers.ReturnErrorMessage(w, err.Error())
				return
			}
			defer resp.Body.Close()
			respbody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err.Error())
				helpers.ReturnErrorMessage(w, err.Error())
				return
			}
			invoices := []models.Invoice{}
			err = json.Unmarshal(respbody, &invoices)
			if err != nil {
				log.Println(err.Error())
				helpers.ReturnErrorMessage(w, err.Error())
				return
			}
			if len(invoices) > 0 {
				db.DB.C("posinvoices").With(session).Upsert(bson.M{"invoice_number": invoices[0].InvoiceNumber}, invoices[0])
			}
			w.Write(respbody)
		} else {
			proxy.ProxyToBackend(w, r)
		}
		return
	}

	helpers.ReturnSuccessMessage(w, invoices)
}

// PaginatedResponse swagger:model paginatedResponse
// models a paginated response for list of invoices
type PaginatedResponse struct {
	Count   int                  `json:"count" bson:"count"`
	Results []models.InvoiceLite `json:"results" bson:"results"`
}

// ListInvoicesPaginated swagger:route GET /api/pos/posinvoices/ invoices listInvoicesPaginated
//
// List Invoices (simplified)
//
// retrieves list of settled or open invoices
// the settled invoices are being proxied to backend.
// the result is returned in a paginated style
//
// Parameters:
// + name: is_settled
//   in: query
//   required: true
//   schema:
//     type: boolean
//
// Responses:
//   200: paginatedResponse
func ListInvoicesPaginated(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	isSettled := r.URL.Query().Get("is_settled")
	if isSettled == "true" {
		proxy.ProxyToBackend(w, r)
		return
	}
	invoices := []models.InvoiceLite{}
	q["is_settled"] = false
	store := r.URL.Query().Get("store")
	if store != "" {
		q["store"], _ = strconv.Atoi(store)
	}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("posinvoices").With(session).Find(q).Sort("-created_on").All(&invoices)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	resp := PaginatedResponse{
		Count:   len(invoices),
		Results: invoices,
	}
	helpers.ReturnSuccessMessage(w, resp)
}

// GetInvoice swagger:route GET /api/pos/posinvoices/{invoice_number}/ invoices getInvoice
//
// Get Invoice
//
// fetches invoice from the database by invoice number
//
// Parameters:
// + name: invoice_number
//   in: path
//   required: true
//   schema:
//      type: string
//
// Responses:
//   200: invoice
func GetInvoice(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["invoice_number"]
	q["invoice_number"] = id

	invoice := models.Invoice{}
	session := db.Session.Copy()
	defer session.Close()

	err := db.DB.C("posinvoices").With(session).Find(q).One(&invoice)
	if err != nil {
		//helpers.ReturnErrorMessage(w, err.Error())
		proxy.ProxyToBackend(w, r)
		return
	}
	helpers.ReturnSuccessMessage(w, invoice)
}

// SubmitInvoice swagger:route POST /api/pos/posinvoices/ invoices submitInvoice
//
// Submit Invoice
//
// creates a new invoice or update an old one
//
// Parameters:
//  + name: body
//    in: body
//    type: invoicePOSTRequest
//    required: true
//
// Responses:
// 200: invoice
func SubmitInvoice(w http.ResponseWriter, r *http.Request) {
	fmt.Println("SubmitInvoice")
	var req models.InvoicePOSTRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	fmt.Printf("Body %v\n", r.Body)
	defer r.Body.Close()
	// validate that all the items has and item id, otherwise return error
	// It's a safe guard for the bug of created item without any info
	for _, item := range req.Invoice.Items {
		if item.Item == 0 {
			helpers.ReturnErrorMessage(w, "One of the items is corrupted, please discard and try again")
			return
		}
	}

	fdmResponses := []models.FDMResponse{}
	// if fdm is enabled submit items to fdm first
	if config.Config.IsFDMEnabled == true {
		// create fdm connection
		conn, err := fdm.Connect(req.RCRS)
		if err != nil {
			log.Println(err)
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		defer conn.Close()

		responses, err := fdm.Submit(conn, req)
		if err != nil {
			log.Println(err)
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		fdmResponses = append(fdmResponses, responses...)
		req.Invoice.FDMResponses = fdmResponses
	}
	session := db.Session.Copy()
	defer session.Close()

	//Add audit_date if it was nil
	if req.Invoice.AuditDate == nil {
		type auditDate struct {
			Date string `json:"audit_date" bson:"audit_date"`
		}
		a := auditDate{}
		db.DB.C("audit_date").With(session).Find(nil).One(&a)
		req.Invoice.AuditDate = &a.Date
	}

	invoice, err := req.Submit()
	if err != nil {
		log.Println("ERROR:", err.Error())
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)

	fmt.Printf("Printing Enabled %v\n", checkProxyPrintingEnabled())
	//TODO :: Print on kitchen Printer
	if checkProxyPrintingEnabled() {
		printReq := PrintRequest{}
		printReq.PrinterType = kitchenPrinter
		printReq.Invoice = req.Invoice
		printReq.Items = req.Invoice.Items
		printReq.OrderedItems = req.Invoice.OrderedItems
		go sendToPrint(printReq)
	}
	req.Invoice = invoice
	req.Invoice.Events = []models.EJEvent{}
	req.Invoice.OrderedItems = []models.EJEvent{}
	err = db.DB.C("posinvoices").With(session).Update(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, req.Invoice)
	if err != nil {
		log.Println(err)
	}
	if req.Invoice.CreateLock == true {
		locks.LockInvoices([]models.Invoice{invoice}, invoice.TerminalID)
	}
	req.Invoice.CreateLock = false
	err = db.DB.C("posinvoices").With(session).Update(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, req.Invoice)
	if err != nil {
		log.Println(err)
	}

	helpers.ReturnSuccessMessage(w, req.Invoice)
}

// VoidInvoice swagger:route POST /api/pos/posinvoices/{invoice_number}/void/ invoices voidInvoice
//
// Void Invoice
//
// closes invoice and sets the void reason
//
// Parameters:
// + name: invoice_number
//   in: path
//   required: true
//   schema:
//      type: string
//
// + name: body
//   in: body
//   type: invoicePOSTRequest
//   required: true
//
// Responses:
// 200: invoice
func VoidInvoice(w http.ResponseWriter, r *http.Request) {
	var req models.InvoicePOSTRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

	// if fdm is enabled submit items to fdm first
	if config.Config.IsFDMEnabled == true {
		// create fdm connection
		conn, err := fdm.Connect(req.RCRS)
		if err != nil {
			log.Println(err)
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		defer conn.Close()

		responses, err := fdm.EmptyPLUHash(conn, req)
		if err != nil {
			log.Println(err)
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		fdmResponses := []models.FDMResponse{}
		fdmResponses = append(fdmResponses, responses...)
		req.Invoice.FDMResponses = fdmResponses
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)

	session := db.Session.Copy()
	defer session.Close()

	req.Invoice.IsSettled = true
	err = db.DB.C("posinvoices").With(session).Update(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, req.Invoice)
	if err != nil {
		log.Println(err)
	}
	// update table status
	if req.Invoice.TableID != nil {
		table := models.Table{}
		err = db.DB.C("tables").With(session).Find(bson.M{"id": req.Invoice.TableID}).One(&table)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		table.UpdateStatus()
	}
	helpers.ReturnSuccessMessage(w, req.Invoice)
}

// BulkSubmitInvoices swagger:route POST /api/pos/posinvoices/bulksubmit/ invoices bulkSubmitInvoices
//
// Bulk Submit Invoices
//
// loops over list of invoices, creates or updates
// them, then release the terminal that is locked.
//
// Parameters:
//  + name: body
//    in: body
//    type: bulkSubmitRequest
//    required: true
//
// Responses:
// 200: bulkSubmitResponse
func BulkSubmitInvoices(w http.ResponseWriter, r *http.Request) {
	req := models.BulkSubmitRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

	var conn *libfdm.FDM
	if config.Config.IsFDMEnabled == true {
		// create fdm connection
		var err error
		conn, err = fdm.Connect(req.RCRS)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		defer conn.Close()
	}

	for _, invoice := range req.Invoices {
		if config.Config.IsFDMEnabled == true {
			// build a normal request model
			invoiceReq := models.InvoicePOSTRequest{}
			invoiceReq.RCRS = req.RCRS
			invoiceReq.TerminalID = req.TerminalID
			invoiceReq.TerminalName = req.TerminalName
			invoiceReq.ActionTime = req.ActionTime
			invoiceReq.Invoice = invoice
			invoiceReq.TerminalNumber = req.TerminalNumber
			invoiceReq.EmployeeID = req.EmployeeID
			invoiceReq.CashierName = req.CashierName
			invoiceReq.CashierNumber = req.CashierNumber
			responses, err := fdm.Submit(conn, invoiceReq)
			if err != nil {
				helpers.ReturnErrorMessage(w, err.Error())
				return
			}
			invoice.FDMResponses = responses
		}
		invoice.Submit(req.TerminalID)
	}
	// release terminal lock
	locks.UnlockTerminal(req.TerminalID)
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)
	// clear events
	session := db.Session.Copy()
	defer session.Close()

	for _, invoice := range req.Invoices {
		invoice.Events = []models.EJEvent{}
		db.DB.C("posinvoices").With(session).Upsert(bson.M{"invoice_number": invoice.InvoiceNumber}, invoice)
	}
	resp := models.BulkSubmitResponse{
		Status: 200,
	}
	helpers.ReturnSuccessMessage(w, resp)
}

// UnlockInvoice swagger:route GET /api/pos/posinvoices/{invoice_number}/unlock/ invoices unlockInvoice
//
// Unlock Invoice
//
// removes invoice_number key from redis and make
// the invoice available to be picked up again by cashiers
//
// Parameters:
// + name: invoice_number
//   in: path
//   required: true
//   schema:
//      type: string
func UnlockInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceNumber := vars["invoice_number"]
	invoice := models.Invoice{}
	session := db.Session.Copy()
	defer session.Close()

	err := db.DB.C("posinvoices").With(session).Find(bson.M{"invoice_number": invoiceNumber}).One(&invoice)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
	}
	locks.UnlockInvoices([]models.Invoice{invoice})
	helpers.ReturnSuccessMessage(w, true)
}

// FolioInvoice swagger:route POST /api/pos/posinvoices/folio/ invoices folioInvoice
//
// Folio
//
// sends invoice lineitems to FDM and increase printing counter
//
// Parameters:
// + name: body
//   in: body
//   type: invoicePOSTRequest
//   required: true
//
// Responses:
// 200: invoice
func FolioInvoice(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n\tFolioInvoice")
	var req models.InvoicePOSTRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

	fdmResponses := []models.FDMResponse{}
	fdmSubmitResCount := 0
	// if fdm is enabled submit items to fdm first
	if config.Config.IsFDMEnabled == true {
		// create fdm connection
		conn, err := fdm.Connect(req.RCRS)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		defer conn.Close()
		responses, err := fdm.Submit(conn, req)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		fdmSubmitResCount += len(responses)
		fdmResponses = append(fdmResponses, responses...)
	}

	req.Invoice, err = req.Submit()
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	if config.Config.IsFDMEnabled == true {
		// create fdm connection
		conn, err := fdm.Connect(req.RCRS)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		defer conn.Close()
		responses, err := fdm.Folio(conn, req)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}

		fdmResponses = append(fdmResponses, responses...)
		req.Invoice.FDMResponses = fdmResponses
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)
	// remove the FDM submit response from fdm responses
	// so that the folio would render correct data. Since
	// the EJ should know about the submit + folio transactions
	// but the UI only needs the FDM folio response
	if config.Config.IsFDMEnabled {
		req.Invoice.FDMResponses = req.Invoice.FDMResponses[fdmSubmitResCount:]
	}
	fmt.Printf("Printing Enabled %v\n", checkProxyPrintingEnabled())
	//Print on Folio  && Kitchen Printer
	if checkProxyPrintingEnabled() {
		printFolioReq := PrintRequest{}
		printFolioReq.PrinterType = folioPrinter
		printFolioReq.Invoice = req.Invoice
		printFolioReq.Items = req.Invoice.Items
		printFolioReq.OrderedItems = req.Invoice.OrderedItems
		go sendToPrint(printFolioReq)
		printKitReq := PrintRequest{}
		printKitReq.PrinterType = kitchenPrinter
		printKitReq.Invoice = req.Invoice
		printKitReq.Items = req.Invoice.Items
		printKitReq.OrderedItems = req.Invoice.OrderedItems
		go sendToPrint(printKitReq)
	}
	req.Invoice.Events = []models.EJEvent{}
	req.Invoice.OrderedItems = []models.EJEvent{}

	req.Invoice.PrintCount++

	session := db.Session.Copy()
	defer session.Close()

	db.DB.C("posinvoices").With(session).Update(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, req.Invoice)

	helpers.ReturnSuccessMessage(w, req.Invoice)
}

// PayInvoice  swagger:route POST /api/pos/posinvoices/{invoice_number}/createpostings/ invoices createPostings
//
// Create Postings (pay invoice)
//
// creates pospayments and pospostinginformations on invoice
//
// Parameters:
// + name: invoice_number
//   in: path
//   required: true
//   schema:
//      type: integer
//
// + name: body
//   in: body
//   type: invoicePOSTRequest
//   required: true
//
// Responses:
// 200: invoicePOSTRequest
func PayInvoice(w http.ResponseWriter, r *http.Request) {
	var req models.InvoicePOSTRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

	fdmResponses := []models.FDMResponse{}

	if config.Config.IsFDMEnabled == true && req.ModalName == "payment" {
		// create fdm connection
		conn, err := fdm.Connect(req.RCRS)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		defer conn.Close()
		responses, err := fdm.Payment(conn, req)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		fdmResponses = append(fdmResponses, responses...)
		req.Invoice.FDMResponses = fdmResponses
	}

	paidOnOpera := true
	if config.Config.IsOperaEnabled && len(req.Postings) > 0 {
		paidOnOpera = HandleOperaPayments(req.Invoice, req.Postings[0].Department)
		if !paidOnOpera {
			helpers.ReturnErrorMessage(w, "Failed to pay on Opera")
			return
		}
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)

	session := db.Session.Copy()
	defer session.Close()
	dbInvoice := models.Invoice{}
	err = db.DB.C("posinvoices").
		With(session).
		Find(bson.M{"invoice_number": req.Invoice.InvoiceNumber}).
		One(&dbInvoice)
	if err != nil {
		log.Println("failed to find posinvoice with this invoice number")
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	for i := 0; i < len(req.Postings); i++ {
		req.Postings[i].PosPostingInformations = []models.Posting{}
		req.Postings[i].PosPostingInformations = append(req.Postings[i].PosPostingInformations, models.Posting{})
		req.Postings[i].PosPostingInformations[0].Comments = ""
	}
	dbInvoice.Postings = append(dbInvoice.Postings, req.Postings...)
	req.Invoice.Postings = dbInvoice.Postings
	req.Invoice.IsSettled = true
	req.Invoice.PaidAmount = req.Invoice.Total
	req.Invoice.Change = req.ChangeAmount

	err = db.DB.C("posinvoices").With(session).Update(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, req.Invoice)
	if err != nil {
		log.Println("failed to find posinvoice with this invoice number")
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	// update table status
	if req.Invoice.TableID != nil {
		table := models.Table{}
		err = db.DB.C("tables").With(session).Find(bson.M{"id": req.Invoice.TableID}).One(&table)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		table.UpdateStatus()
	}

	fmt.Printf("Printing Enabled %v\n", checkProxyPrintingEnabled())
	//TODO :: Print on Folio Printer
	if checkProxyPrintingEnabled() {
		printReq := PrintRequest{}
		printReq.PrinterType = folioPrinter
		printReq.Invoice = req.Invoice
		printReq.OrderedItems = req.Invoice.OrderedItems
		printReq.Items = req.Invoice.Items
		go sendToPrint(printReq)
	}
	helpers.ReturnSuccessMessage(w, req)
}

// CreatePaymentEJ swagger:route POST /api/pos/posinvoices/createpaymentej/ invoices createPaymentEJ
//
// Create Payment EJ
//
// takes a map of values and proxy it to backend in order to create ej for payment
func CreatePaymentEJ(w http.ResponseWriter, r *http.Request) {
	body := make(map[string]interface{})
	json.NewDecoder(r.Body).Decode(&body)
	err := syncer.QueueRequest(r.RequestURI, r.Method, r.Header, body)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, true)
}

// CancelPostings swagger:route POST /api/pos/posinvoices/{invoice_number}/cancelpostings/ invoices cancelPostings
//
// Cancel Postings
//
// cancels payments of a paid invocie based on postings frontend ids
//
// Parameters:
// + name: invoice_number
//   in: path
//   required: true
//   schema:
//      type: integer
//
// + name: body
//   in: body
//   type: cancelPostingsRequest
//   required: true
//
// Responses:
// 200: []posting
func CancelPostings(w http.ResponseWriter, r *http.Request) {
	req := models.CancelPostingsRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

	vars := mux.Vars(r)
	invoiceNumber, _ := vars["invoice_number"]
	invoice := models.Invoice{}
	session := db.Session.Copy()
	defer session.Close()
	err = db.DB.C("posinvoices").With(session).Find(bson.M{"invoice_number": invoiceNumber}).One(&invoice)
	if err != nil {
		log.Println("failed to find posinvoice with this invoice number")
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)

	newPayments := []models.Posting{}
	for i := 0; i < len(req.PostingsIDs); i++ {
		for j := 0; j < len(invoice.Postings); j++ {
			if req.PostingsIDs[i] == invoice.Postings[j].FrontendID {
				newPayment := models.Posting{}
				newPayment.Amount = -1 * invoice.Postings[j].Amount
				newPayment.ForeignAmount = -1 * invoice.Postings[j].ForeignAmount
				newPayment.AuditDate = invoice.Postings[j].AuditDate
				newPayment.PostingType = invoice.Postings[j].PostingType
				newPayment.DepartmentDetails = invoice.Postings[j].DepartmentDetails
				if invoice.Postings[j].Room != nil {
					newPayment.Room = invoice.Postings[j].Room
					newPayment.RoomNumber = invoice.Postings[j].RoomNumber
					newPayment.RoomDetails = invoice.Postings[j].RoomDetails
				}
				newPayment.PaymentLog = invoice.Postings[j].PaymentLog
				if invoice.Postings[j].Sign == "+" {
					newPayment.Sign = "-"
				} else {
					newPayment.Sign = "+"
				}
				newPayment.Type = invoice.Postings[j].Type

				newPayment.PosPostingInformations = []models.Posting{}
				newPayment.PosPostingInformations = append(newPayment.PosPostingInformations, models.Posting{Comments: ""})
				newPayments = append(newPayments, newPayment)
				break
			}
		}
	}
	invoice.Postings = append(invoice.Postings, newPayments...)

	err = db.DB.C("posinvoices").With(session).Update(bson.M{"invoice_number": invoice.InvoiceNumber}, invoice)
	if err != nil {
		log.Println("failed to find posinvoice with this invoice number")
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, invoice)
}

// RefundInvoice swagger:route POST /api/pos/posinvoices/refund/ invoices refundInvoice
//
// Refund Invoice
//
// handles the refund scenario and cancel postings.
// Creating a new invoice with the canceled postings.
//
// Parameters:
// + name: body
//   in: body
//   type: refundInvoiceRequest
//   required: true
//
// Responses:
// 200: refundInvoiceResponse
func RefundInvoice(w http.ResponseWriter, r *http.Request) {
	body := models.RefundInvoiceRequest{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Println("Error:", err.Error())
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

	terminalIDStr := r.URL.Query().Get("terminal_id")
	terminalID, _ := strconv.Atoi(terminalIDStr)
	invoiceNumber, err := models.AdvanceInvoiceNumber(terminalID)
	if err != nil {
		log.Println("refund error", err.Error())
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	body.NewInvoice.InvoiceNumber = invoiceNumber

	fdmResponses := []models.FDMResponse{}

	req := models.InvoicePOSTRequest{}
	req.RCRS = body.RCRS
	req.EmployeeID = body.EmployeeID
	req.Invoice = body.NewInvoice
	req.TerminalID = body.TerminalID
	req.TerminalNumber = body.TerminalNumber
	req.TerminalName = body.TerminalName
	req.CashierName = body.CashierName
	req.CashierNumber = body.CashierNumber
	req.IsClosed = true
	req.ChangeAmount = 0
	req.ActionTime = body.ActionTime
	if config.Config.IsFDMEnabled == true {
		// create fdm connection
		conn, err := fdm.Connect(req.RCRS)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		defer conn.Close()
		responses, err := fdm.Payment(conn, req)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		fdmResponses = append(fdmResponses, responses...)
		req.Invoice.FDMResponses = fdmResponses
		body.NewInvoice.FDMResponses = fdmResponses
	}

	paidOnOpera := true
	if config.Config.IsOperaEnabled && !req.Invoice.HouseUse {
		paidOnOpera = HandleOperaPayments(req.Invoice, body.DepartmentID)
		if !paidOnOpera {
			helpers.ReturnErrorMessage(w, "Failed to refund on Opera")
			return
		}
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, body)
	body.NewInvoice, err = req.Submit()
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	req.Invoice = body.NewInvoice
	req.Invoice.IsSettled = true
	req.Invoice.PaidAmount = req.Invoice.Total

	session := db.Session.Copy()
	defer session.Close()

	db.DB.C("posinvoices").With(session).Upsert(bson.M{"invoice_number": body.NewInvoice.InvoiceNumber}, req.Invoice)

	resp := &models.RefundInvoiceResponse{}
	resp.NewInvoice = req.Invoice
	resp.OriginalInvoice = models.Invoice{}
	if body.NewInvoice.HouseUse == false {
		resp.Postings = []models.Posting{}
		body.Posting.PosPostingInformations = []models.Posting{}
		body.Posting.PosPostingInformations = append(body.Posting.PosPostingInformations, models.Posting{})
		resp.Postings = append(resp.Postings, body.Posting)
		resp.NewInvoice.Postings = resp.Postings
	}
	db.DB.C("posinvoices").With(session).Find(bson.M{"invoice_number": body.OriginalInvoice.InvoiceNumber}).One(&resp.OriginalInvoice)
	// change the returned_qty of the line items that haven refunded
	for _, item := range body.NewInvoice.Items {
		for i, oldItem := range resp.OriginalInvoice.Items {
			if oldItem.FrontendID == *item.OriginalFrontendID {
				resp.OriginalInvoice.Items[i].ReturnedQuantity += -1 * item.Quantity
				break
			}
		}
	}
	db.DB.C("posinvoices").With(session).Update(bson.M{"invoice_number": body.OriginalInvoice.InvoiceNumber}, resp.OriginalInvoice)
	helpers.ReturnSuccessMessage(w, resp)

}

// Houseuse  swagger:route POST /api/pos/posinvoices/houseuse/ invoices houseuse
//
// Houseuse
//
// pay the invoice and settle it as house use
//
// Parameters:
// + name: body
//   in: body
//   type: invoicePOSTRequest
//   required: true
//
// Responses:
// 200: invoice
func Houseuse(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n\tHouseUse\t\n")
	var req models.InvoicePOSTRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

	fdmResponses := []models.FDMResponse{}

	if config.Config.IsFDMEnabled == true {
		// create fdm connection
		conn, err := fdm.Connect(req.RCRS)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		defer conn.Close()
		responses, err := fdm.Payment(conn, req)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		fdmResponses = append(fdmResponses, responses...)
		req.Invoice.FDMResponses = fdmResponses
	}

	invoice, err := req.Submit()
	if err != nil {
		log.Println("Submit error:", err.Error())
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	req.Invoice = invoice
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)
	req.Invoice.HouseUse = true
	req.Invoice.PaidAmount = req.Invoice.Total
	postings := []models.Posting{}
	posting := models.Posting{PostingType: "houseuse", Amount: req.Invoice.Total}
	postings = append(postings, posting)
	req.Postings = postings

	session := db.Session.Copy()
	defer session.Close()
	err = db.DB.C("posinvoices").With(session).Update(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, req.Invoice)
	if err != nil {
		log.Println("Mongodb error:", err.Error())
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	// update table status
	if req.Invoice.TableID != nil && *req.Invoice.TableID != 0 {
		table := models.Table{}
		err = db.DB.C("tables").With(session).Find(bson.M{"id": req.Invoice.TableID}).One(&table)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		table.UpdateStatus()
	}

	fmt.Printf("Printing Enabled %v\n", checkProxyPrintingEnabled())
	//TODO :: Print on Folio Printer
	if checkProxyPrintingEnabled() {
		printReq := PrintRequest{}
		printReq.PrinterType = folioPrinter
		printReq.Invoice = req.Invoice
		printReq.Items = req.Invoice.Items
		printReq.OrderedItems = req.Invoice.OrderedItems
		copier.Copy(printReq.OrderedItems, req.Invoice.OrderedItems)
		go sendToPrint(printReq)
	}
	helpers.ReturnSuccessMessage(w, req.Invoice)
}

// ChangeTable swagger:route POST /api/pos/posinvoices/changetable/ invoices changeTable
//
// Change Table
//
// moves the selected invoices from table to another table
//
// Parameters:
// + name: body
//   in: body
//   type: changeTableRequest
//   required: true
//
// Responses:
// 200: []invoice
func ChangeTable(w http.ResponseWriter, r *http.Request) {
	body := models.ChangeTableRequest{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

	// update invoices in db
	session := db.Session.Copy()
	defer session.Close()

	newTable := models.Table{}
	log.Println("new table id", body.NewTable)
	err = db.DB.C("tables").With(session).Find(bson.M{"id": body.NewTable}).One(&newTable)
	if err != nil {
		log.Println(err)
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	for _, i := range body.Invoices {
		selector := bson.M{"invoice_number": i.InvoiceNumber}
		*i.TableID = int64(body.NewTable)
		*i.TableDetails = newTable.Description
		*i.TableNumber = int64(newTable.Number)
		err = db.DB.C("posinvoices").With(session).Update(selector, i)
		if err != nil {
			log.Println(err)
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, body)

	// Update Status of new Table
	newTable.UpdateStatus()

	// Update Status of old Table
	oldTable := models.Table{}
	err = db.DB.C("tables").With(session).Find(bson.M{"id": body.OldTable}).One(&oldTable)
	if err != nil {
		log.Println(err)
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	oldTable.UpdateStatus()

	// get invoices on the new table
	newInvoices := []models.Invoice{}
	err = db.DB.C("posinvoices").With(session).Find(bson.M{"table_number": body.NewTable, "is_settled": false}).All(&newInvoices)
	if err != nil {
		log.Println(err)
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, newInvoices)
}

// SplitInvoices swagger:route POST /api/pos/posinvoices/split/ invoices split
//
// Split
//
// splits invoice to new invoices
//
// Parameters:
// + name: body
//   in: body
//   type: splitInvoiceRequest
//   required: true
//
// Responses:
// 200: []invoice
func SplitInvoices(w http.ResponseWriter, r *http.Request) {
	body := models.SplitInvoiceRequest{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

	session := db.Session.Copy()
	defer session.Close()

	for _, i := range body.Invoices {

		err := db.DB.C("posinvoices").With(session).Update(bson.M{"invoice_number": i.InvoiceNumber}, i)
		if err != nil {
			log.Println("DB ERROR", err)
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}

	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, body)

	helpers.ReturnSuccessMessage(w, body.Invoices)
}

// WasteAndVoid wastes a  swagger:route POST /api/pos/posinvoices/wasteandvoid/ invoices wasteAndVoid
//
// Waste & Void
//
// Wastes a lineitem
//
// Parameters:
// + name: body
//   in: body
//   type: invoicePOSTRequest
//   required: true
//
// Responses:
// 200: invoice
func WasteAndVoid(w http.ResponseWriter, r *http.Request) {
	req := models.InvoicePOSTRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

	terminalIDStr := r.URL.Query().Get("terminal_id")
	terminalID, _ := strconv.Atoi(terminalIDStr)
	terminal := models.Terminal{}
	session := db.Session.Copy()
	defer session.Close()
	err = db.DB.C("terminals").With(session).Find(bson.M{"id": terminalID}).One(&terminal)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	lineitem := req.Invoice.Items[len(req.Invoice.Items)-1]
	lineitem.SubmittedQuantity = lineitem.Quantity

	req.Invoice.Items[len(req.Invoice.Items)-1] = lineitem

	if config.Config.IsFDMEnabled == true {
		// create fdm connection
		conn, err := fdm.Connect(terminal.RCRS)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		defer conn.Close()
		responses, err := fdm.Submit(conn, req)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		req.Invoice.FDMResponses = responses
	}
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)
	req.Invoice.Events = []models.EJEvent{}

	err = db.DB.C("posinvoices").With(session).Update(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, req.Invoice)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	helpers.ReturnSuccessMessage(w, req.Invoice)
}

// ToggleLocking swagger:route GET /api/pos/posinvoices/togglelocking/
//
// Toggle Locking
//
// toggle locking of invoices
//
// Parameters:
// + name: id
//   in: query
//   required: true
//
// + name: terminal_id
//   in: query
//   required: true
//   schema:
//      type: integer
//
// + name: target
//   in: query
//   schema:
//     type: string
//   required: true
func ToggleLocking(w http.ResponseWriter, r *http.Request) {
	numbers := strings.Split(r.URL.Query().Get("id"), ",")
	terminalIDStr := r.URL.Query().Get("terminal_id")
	terminalID, _ := strconv.Atoi(terminalIDStr)
	target := r.URL.Query().Get("target")

	invoices := []models.Invoice{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("posinvoices").With(session).Find(bson.M{"invoice_number": bson.M{"$in": numbers}}).All(&invoices)
	if err != nil {
		log.Println(err.Error())
		helpers.ReturnSuccessMessage(w, err.Error())
		return
	}

	if target == "lock" {
		otherTerminal, err := locks.LockInvoices(invoices, terminalID)
		if err != nil {
			helpers.ReturnErrorMessageWithStatus(w, 409, fmt.Sprintf("Invoices locked by Terminal %d", otherTerminal))
			return
		}
	} else {
		locks.UnlockInvoices(invoices)
	}

	helpers.ReturnSuccessMessage(w, true)
}

// GetInvoiceLatestChanges swagger:route POST /api/pos/posinvoices/{invoice_number}/getlatestchanges/ invoices getInvoiceLatestChanges
//
// Get Invoice Latest Changes
//
// gets the invoice and checks if it's locked or not
//
// Parameters:
// + name: invoice_number
//   in: path
//   required: true
//   schema:
//      type: integer
//
// Responses:
// 200: invoice
func GetInvoiceLatestChanges(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	terminalID, _ := strconv.Atoi(params["terminal_id"][0])
	invoiceNumber, _ := mux.Vars(r)["invoice_number"]
	invoice := models.Invoice{InvoiceNumber: invoiceNumber}
	otherTerminal, err := locks.LockInvoices([]models.Invoice{invoice}, terminalID)
	if err != nil {
		log.Println(err)
		res := bson.M{"terminal": otherTerminal, "lockedposinvoices": true, "posinvoice": invoice}
		helpers.ReturnErrorMessageWithStatus(w, 409, res)
		return
	}
	res := bson.M{"terminal": nil, "lockedposinvoices": false, "posinvoice": invoice}
	helpers.ReturnSuccessMessage(w, res)
}

func AddToPostRequest(postRequest *opera.PostRequest,
	revenueConfig map[int]string, taxConfig map[int]string,
	serviceConfig map[int]string, department int,
	taxes map[int]int64, service map[int]int64, discounts int64,
	subtotal int64) {
	if revenueConfig[department] == "1" {
		postRequest.Subtotal1 += subtotal - discounts
		postRequest.Discount1 += discounts
		postRequest.TotalAmount += subtotal
	} else if revenueConfig[department] == "2" {
		postRequest.Subtotal2 += subtotal - discounts
		postRequest.Discount2 += discounts
		postRequest.TotalAmount += subtotal
	} else if revenueConfig[department] == "3" {
		postRequest.Subtotal3 += subtotal - discounts
		postRequest.Discount3 += discounts
		postRequest.TotalAmount += subtotal
	} else if revenueConfig[department] == "4" {
		postRequest.Subtotal4 += subtotal - discounts
		postRequest.Discount4 += discounts
		postRequest.TotalAmount += subtotal
	} else if revenueConfig[department] == "5" {
		postRequest.Subtotal5 += subtotal - discounts
		postRequest.Discount5 += discounts
		postRequest.TotalAmount += subtotal
	} else if revenueConfig[department] == "6" {
		postRequest.Subtotal6 += subtotal - discounts
		postRequest.Discount6 += discounts
		postRequest.TotalAmount += subtotal
	}
	for key, value := range taxes {
		if taxConfig[key] == "1" {
			postRequest.Tax1 += value
		} else if taxConfig[key] == "2" {
			postRequest.Tax2 += value
		} else if taxConfig[key] == "3" {
			postRequest.Tax3 += value
		} else if taxConfig[key] == "4" {
			postRequest.Tax4 += value
		}
		postRequest.TotalAmount += value
	}
	for key, value := range service {
		if serviceConfig[key] == "1" {
			postRequest.ServiceCharge1 += value
		} else if serviceConfig[key] == "2" {
			postRequest.ServiceCharge2 += value
		} else if serviceConfig[key] == "3" {
			postRequest.ServiceCharge3 += value
		} else if serviceConfig[key] == "4" {
			postRequest.ServiceCharge4 += value
		}
		postRequest.TotalAmount += value
	}
}

func HandleOperaPayments(invoice models.Invoice, department int) bool {
	postRequest := opera.PostRequest{}
	taxConfig := []opera.OperaConfig{}
	revenueConfig := []opera.OperaConfig{}
	serviceConfig := []opera.OperaConfig{}
	session := db.Session.Copy()
	defer session.Close()

	_ = db.DB.C("operasettings").With(session).Find(bson.M{"config_name": "tax"}).All(&taxConfig)
	_ = db.DB.C("operasettings").With(session).Find(bson.M{"config_name": "revenue_department"}).All(&revenueConfig)
	_ = db.DB.C("operasettings").With(session).Find(bson.M{"config_name": "service_charge"}).All(&serviceConfig)

	taxFlattenedMap := opera.FlattenToMap(taxConfig)
	revenueFlattenedMap := opera.FlattenToMap(revenueConfig)
	serviceFlattenedMap := opera.FlattenToMap(serviceConfig)
	for _, lineitem := range invoice.Items {
		var taxes map[int]int64
		var discounts float64
		var service map[int]int64

		departmentID := lineitem.AttachedAttributes["revenue_department"]
		department := incomemodels.Department{}
		_ = db.DB.C("departments").With(session).Find(bson.M{"id": departmentID}).One(&department)

		price := float64(lineitem.Price)
		roundedPrice := helpers.ConvertToInt(helpers.Round(price, 0.05))
		discounts = ComputeDiscounts(price, lineitem.AppliedDiscounts)
		discountsInt := helpers.ConvertToInt(discounts)
		subtotalFloat := price + discounts
		subtotal := roundedPrice + discountsInt

		taxes, service = ComputeTaxes(subtotalFloat, department.TaxDefs, invoice.TakeOut)
		AddToPostRequest(&postRequest, revenueFlattenedMap,
			taxFlattenedMap, serviceFlattenedMap,
			department.ID, taxes, service, discountsInt, subtotal)

		for _, condimentlineitem := range lineitem.CondimentLineItems {
			condimentDepartment := incomemodels.Department{}
			departmentID := condimentlineitem.AttachedAttributes["revenue_department"]
			_ = db.DB.C("departments").With(session).Find(bson.M{"id": departmentID}).One(&condimentDepartment)

			condimentPrice := float64(lineitem.Quantity * condimentlineitem.Price)
			roundedPrice := helpers.ConvertToInt(helpers.Round(condimentPrice, 0.05))
			discounts = ComputeDiscounts(condimentPrice, lineitem.AppliedDiscounts)
			subtotalFloat := condimentPrice + discounts
			discountsInt := helpers.ConvertToInt(discounts)
			subtotal := roundedPrice + discountsInt

			taxes, service = ComputeTaxes(subtotalFloat, condimentDepartment.TaxDefs, invoice.TakeOut)
			AddToPostRequest(&postRequest, revenueFlattenedMap,
				taxFlattenedMap, serviceFlattenedMap,
				department.ID, taxes, service, discountsInt, subtotal)
		}
	}

	invoiceDate := strings.Split(invoice.LastPaymentDate, "-")
	postRequest.Date = fmt.Sprintf("%02s%02s%02s", invoiceDate[0][2:], invoiceDate[1], invoiceDate[2])

	t := time.Now()
	val := fmt.Sprintf("%02d%02d%02d", t.Hour(), t.Minute(), t.Second())
	postRequest.Time = val

	postRequest.CheckNumber = strings.Replace(invoice.InvoiceNumber, "-", "", -1)
	postRequest.RevenueCenter = invoice.Store
	postRequest.WorkstationId = fmt.Sprintf("%d", invoice.TerminalID)
	paymentMethodInt, _ := opera.GetPaymentMethod(department)
	postRequest.PaymentMethod = paymentMethodInt
	seqNumber, _ := opera.GetNextSequence()
	postRequest.SequenceNumber = seqNumber
	postRequest.RequestType = 1
	postRequest.Covers = invoice.Pax
	if invoice.OperaReservation != "" {
		postRequest.ReservationId = invoice.OperaReservation
		postRequest.RoomNumber = invoice.OperaRoomNumber
		postRequest.LastName = strings.Split(invoice.WalkinName, "/")[1]
	}

	buf := bytes.NewBufferString("")
	if err := xml.NewEncoder(buf).Encode(postRequest); err != nil {
		log.Println(err)
		return false
	}
	msg, _ := opera.SendRequest([]byte(buf.String()))

	if len(msg) > 1 {
		msg = msg[1 : len(msg)-1]
	}
	postAnswer := opera.PostAnswer{}
	responseBuf := bytes.NewBufferString(msg)
	if err := xml.NewDecoder(responseBuf).Decode(&postAnswer); err != nil {
		log.Println("error parsing", err)
		return false
	}
	if postAnswer.AnswerStatus != "OK" {
		log.Println("post answer not OK")
		return false
	}
	return true
}

func ComputeTaxes(amount float64, tax_defs map[string][]incomemodels.TaxDef,
	takeout bool) (map[int]int64, map[int]int64) {
	serviceConfig := opera.OperaConfig{}
	taxConfig := opera.OperaConfig{}
	taxMap := map[int]float64{}
	serviceMap := map[int]float64{}
	taxMapInt := map[int]int64{}
	serviceMapInt := map[int]int64{}
	session := db.Session.Copy()
	defer session.Close()
	_ = db.DB.C("operasettings").With(session).Find(bson.M{"config_name": "service_charge"}).One(&serviceConfig)
	_ = db.DB.C("operasettings").With(session).Find(bson.M{"config_name": "tax"}).One(&taxConfig)
	w := float64(1.0)
	requiredTax := ""
	if takeout {
		requiredTax = "takeout"
	} else {
		requiredTax = "dinein"
	}
	for key, tax_types := range tax_defs {
		for _, tax_def := range tax_types {
			if tax_def.POS == "all" || tax_def.POS == requiredTax {
				if key == "fix" {
					newFormula := strings.Replace(tax_def.Formula, "x", "1", -1)
					output, _ := golpal.New().ExecuteSimple(newFormula)
					value, _ := strconv.ParseFloat(output, 32)
					amount -= value
				} else if key == "in" {
					newFormula := strings.Replace(tax_def.Formula, "x", "1", -1)
					output, _ := golpal.New().ExecuteSimple(newFormula)
					value, _ := strconv.ParseFloat(output, 32)
					w += value
				}
			}
		}
	}
	net_amount := amount / w

	for key, tax_types := range tax_defs {
		for _, tax_def := range tax_types {
			if (key == "ex" || key == "fix_ex") && (tax_def.POS == "all" || tax_def.POS == requiredTax) {
				net_amount_str := fmt.Sprintf("%v", net_amount)
				newFormula := strings.Replace(tax_def.Formula, "x", net_amount_str, -1)
				output, _ := golpal.New().ExecuteSimple(newFormula)
				value, _ := strconv.ParseFloat(output, 32)
				serviceFound := opera.CheckInArray(tax_def.DepartmentID, serviceConfig.Value.Departments)
				taxFound := opera.CheckInArray(tax_def.DepartmentID, taxConfig.Value.Departments)
				if serviceFound {
					if _, ok := serviceMap[tax_def.DepartmentID]; ok {
						serviceMap[tax_def.DepartmentID] += value
					} else {
						serviceMap[tax_def.DepartmentID] = value
					}

				}
				if taxFound {
					if _, ok := taxMap[tax_def.DepartmentID]; ok {
						taxMap[tax_def.DepartmentID] += value
					} else {
						taxMap[tax_def.DepartmentID] = value
					}

				}
			}
		}
	}

	for key, value := range serviceMap {
		serviceMapInt[key] = helpers.ConvertToInt(helpers.Round(value, 0.05))
	}
	for key, value := range taxMap {
		taxMapInt[key] = helpers.ConvertToInt(helpers.Round(value, 0.05))
	}

	return taxMapInt, serviceMapInt
}

func ComputeDiscounts(amount float64, discounts []models.AppliedDiscount) float64 {
	discountsValue := 0.0
	for _, discount := range discounts {
		value := amount * discount.Percentage / 100.0
		discountsValue += value
		amount -= value
	}
	return -helpers.Round(discountsValue, 0.05)
}

// CancelLastPayment swagger:route POST /api/pos/posinvoices/{invoice_number}/cancellastpayment/ invoices createPostings
//
// Cancel Last Payment
//
// loop through posting and cancel them in reverse order
//
// Parameters:
// + name: invoice_number
//   in: path
//   required: true
//   schema:
//      type: integer
//
// + name: body
//   in: body
//   type: invoicePOSTRequest
//   required: true
//
// Responses:
// 200: string
// 500: string
func CancelLastPayment(w http.ResponseWriter, r *http.Request) {
	var req models.InvoicePOSTRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()
	err = payment.CancelLastPayment(req)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, "Cancel payment request sent successfully")
	return
}
