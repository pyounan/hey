package pos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/fdm"
	"pos-proxy/pos/locks"
	"pos-proxy/pos/models"
	"pos-proxy/proxy"
	"pos-proxy/syncer"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

// ListInvoices lists open invoices
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
	err := db.DB.C("posinvoices").Find(q).Sort("-created_on").All(&invoices)
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
				helpers.ReturnErrorMessage(w, err)
				return
			}
			respbody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err.Error())
				helpers.ReturnErrorMessage(w, err)
				return
			}
			defer resp.Body.Close()
			invoices := []models.Invoice{}
			err = json.Unmarshal(respbody, &invoices)
			if err != nil {
				log.Println(err.Error())
				helpers.ReturnErrorMessage(w, err)
				return
			}
			if len(invoices) > 0 {
				db.DB.C("posinvoices").Upsert(bson.M{"invoice_number": invoices[0].InvoiceNumber}, invoices[0])
			}
			w.Write(respbody)
		} else {
			proxy.ProxyToBackend(w, r)
		}
		return
	}

	helpers.ReturnSuccessMessage(w, invoices)
}

// ListInvoicesPaginated retrieves list of settled or open invoices
// the settled invoices are being proxied to backend
func ListInvoicesPaginated(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	isSettled := r.URL.Query().Get("is_settled")
	if isSettled == "true" {
		proxy.ProxyToBackend(w, r)
		return
	}
	invoices := []models.Invoice{}
	q["is_settled"] = false
	q["store"], _ = strconv.Atoi(r.URL.Query().Get("store"))
	err := db.DB.C("posinvoices").Find(q).Sort("-created_on").All(&invoices)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	resp := bson.M{}
	resp["count"] = len(invoices)
	resp["results"] = invoices
	helpers.ReturnSuccessMessage(w, resp)
}

// GetInvoice fetches invoice from the database by invoice number
func GetInvoice(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["invoice_nubmer"]
	q["invoice_number"] = id

	invoice := models.Invoice{}
	err := db.DB.C("posinvoices").Find(q).One(&invoice)
	if err != nil {
		//helpers.ReturnErrorMessage(w, err.Error())
		proxy.ProxyToBackend(w, r)
		return
	}
	helpers.ReturnSuccessMessage(w, invoice)
}

// SubmitInvoice creates a new invoice or update an old one
func SubmitInvoice(w http.ResponseWriter, r *http.Request) {
	var req models.InvoicePOSTRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

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
		if req.Invoice.VoidReason != "" {
			responses, err = fdm.EmptyPLUHash(conn, req)
			if err != nil {
				log.Println(err)
				helpers.ReturnErrorMessage(w, err.Error())
				return
			}
		}

		fdmResponses = append(fdmResponses, responses...)
		req.Invoice.FDMResponses = fdmResponses
	}

	invoice, err := req.Submit()
	if err != nil {
		log.Println(err)
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)
	req.Invoice = invoice
	req.Invoice.Events = []models.EJEvent{}
	locks.LockInvoices([]models.Invoice{invoice}, invoice.TerminalID)
	helpers.ReturnSuccessMessage(w, req.Invoice)
}

// UnlockInvoice removes invoice_number key from redis and make
// the invoice available to be picked up again by cashiers
func UnlockInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceNumber := vars["invoice_number"]
	invoice := models.Invoice{}
	err := db.DB.C("posinvoices").Find(bson.M{"invoice_number": invoiceNumber}).One(&invoice)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
	}
	locks.UnlockInvoices([]models.Invoice{invoice})
	helpers.ReturnSuccessMessage(w, true)
}

// FolioInvoice sends invoice lineitems to FDM and increase printing counter
func FolioInvoice(w http.ResponseWriter, r *http.Request) {
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
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		defer conn.Close()
		_, err = fdm.Submit(conn, req)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
	}

	req.Invoice, err = req.Submit()
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	fdmResponses := []models.FDMResponse{}

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
	req.Invoice.Events = []models.EJEvent{}

	req.Invoice.PrintCount++

	db.DB.C("posinvoices").Update(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, req.Invoice)

	helpers.ReturnSuccessMessage(w, req.Invoice)
}

// PayInvoice creates pospayments and pospostinginformations on invoice
func PayInvoice(w http.ResponseWriter, r *http.Request) {
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
		log.Println(req.Invoice.FDMResponses)
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)

	for i := 0; i < len(req.Postings); i++ {
		req.Postings[i].PosPostingInformations = []models.Posting{}
		req.Postings[i].PosPostingInformations = append(req.Postings[i].PosPostingInformations, models.Posting{})
		req.Postings[i].PosPostingInformations[0].Comments = ""
	}
	req.Invoice.Postings = append(req.Invoice.Postings, req.Postings...)
	req.Invoice.IsSettled = true
	req.Invoice.PaidAmount = req.Invoice.Total
	req.Invoice.Change = req.ChangeAmount

	err = db.DB.C("posinvoices").Update(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, req.Invoice)
	if err != nil {
		log.Println("failed to find posinvoice with this invoice number")
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	// update table status
	if req.Invoice.TableID != nil {
		table := models.Table{}
		err = db.DB.C("tables").Find(bson.M{"id": req.Invoice.TableID}).One(&table)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		table.UpdateStatus()
	}

	helpers.ReturnSuccessMessage(w, req)
}

// CancelPostings cancels payments of a paid invocie based on postings frontend ids
func CancelPostings(w http.ResponseWriter, r *http.Request) {
	type CancelPostingsRequest struct {
		PostingsIDs []string `json:"frontend_ids" bson:"frontend_ids"`
		CashierID   int      `json:"poscashier_id" bson:"poscashier_id"`
	}
	req := CancelPostingsRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

	vars := mux.Vars(r)
	invoiceNumber, _ := vars["invoice_number"]
	invoice := models.Invoice{}
	log.Println("invoice number:", invoiceNumber)
	err = db.DB.C("posinvoices").Find(bson.M{"invoice_number": invoiceNumber}).One(&invoice)
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

	err = db.DB.C("posinvoices").Update(bson.M{"invoice_number": invoice.InvoiceNumber}, invoice)
	if err != nil {
		log.Println("failed to find posinvoice with this invoice number")
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, invoice.Postings)
}

// RefundInvoice handles the refund scenario
func RefundInvoice(w http.ResponseWriter, r *http.Request) {
	type ReqBody struct {
		RCRS                  string         `json:"rcrs" bson:"rcrs"`
		TerminalID            int            `json:"terminal_id" bson:"terminal_id"`
		TerminalNumber        int            `json:"terminal_number" bson:"terminal_number"`
		TerminalName          string         `json:"terminal_description" bson:"terminal_description"`
		EmployeeID            string         `json:"employee_id" bson:"employee_id"`
		Invoice               models.Invoice `json:"posinvoice" bson:"posinvoice"`
		OriginalInvoiceNumber string         `json:"original_invoice_number" bson:"original_invoice_number"`
		DepartmentID          int            `json:"department" bson:"department"`
		Posting               models.Posting `json:"posting" bson:"posting"`
		CashierID             int            `json:"cashier_id" bson:"cashier_id"`
		CashierName           string         `json:"cashier_name" bson:"cashier_name"`
		CashierNumber         int            `json:"cashier_number" bson:"cashier_number"`
		Type                  string         `json:"type" bson:"type"`
		ActionTime            string         `json:"action_time" bson:"action_time"`
	}
	body := ReqBody{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Println("########ERRRORRRRR#####", err.Error())
		helpers.ReturnErrorMessage(w, err)
		return
	}
	defer r.Body.Close()

	terminalIDStr := r.URL.Query().Get("terminal_id")
	terminalID, _ := strconv.Atoi(terminalIDStr)
	invoiceNumber, err := models.AdvanceInvoiceNumber(terminalID)
	if err != nil {
		log.Println("refund error", err.Error())
		helpers.ReturnErrorMessage(w, err)
		return
	}
	body.Invoice.InvoiceNumber = invoiceNumber

	fdmResponses := []models.FDMResponse{}

	req := models.InvoicePOSTRequest{}
	req.RCRS = body.RCRS
	req.EmployeeID = body.EmployeeID
	req.Invoice = body.Invoice
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
		_, err = fdm.Submit(conn, req)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		responses, err := fdm.Payment(conn, req)
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		fdmResponses = append(fdmResponses, responses...)
		body.Invoice.FDMResponses = fdmResponses
	}
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, body)
	invoice, err := req.Submit()
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	req.Invoice = invoice
	req.Invoice.PaidAmount = req.Invoice.Total

	db.DB.C("posinvoices").Upsert(bson.M{"invoice_number": body.Invoice.InvoiceNumber}, body.Invoice)

	type RespBody struct {
		NewInvoice      models.Invoice   `json:"new_invoice" bson:"new_invoice"`
		OriginalInvoice models.Invoice   `json:"original_invoice" bson:"original_invoice"`
		Postings        []models.Posting `json:"postings" bson:"postings"`
	}
	resp := &RespBody{}
	resp.NewInvoice = req.Invoice
	resp.OriginalInvoice = models.Invoice{}
	resp.Postings = []models.Posting{}
	body.Posting.PosPostingInformations = []models.Posting{}
	body.Posting.PosPostingInformations = append(body.Posting.PosPostingInformations, models.Posting{})
	resp.Postings = append(resp.Postings, body.Posting)
	db.DB.C("posinvoices").Find(bson.M{"invoice_number": body.OriginalInvoiceNumber}).One(&resp.OriginalInvoice)
	helpers.ReturnSuccessMessage(w, resp)

}

// Houseuse pay the invoice as house use
func Houseuse(w http.ResponseWriter, r *http.Request) {
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
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	req.Invoice = invoice
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)
	req.Invoice.PaidAmount = req.Invoice.Total
	postings := []models.Posting{}
	posting := models.Posting{PostingType: "houseuse", Amount: req.Invoice.Total}
	postings = append(postings, posting)
	req.Postings = postings

	err = db.DB.C("posinvoices").Update(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, bson.M{"$set": req.Invoice})
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	// update table status
	table := models.Table{}
	err = db.DB.C("tables").Find(bson.M{"id": req.Invoice.TableID}).One(&table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	table.UpdateStatus()

	helpers.ReturnSuccessMessage(w, req)
}

// ChangeTable moves the selected invoices from table to another table
func ChangeTable(w http.ResponseWriter, r *http.Request) {
	body := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	oldTable := int(body["oldtable"].(float64))
	newTable := int(body["newtable"].(float64))
	invoices := body["posinvoices"].([]interface{})
	invoiceNumbers := []string{}
	for _, i := range invoices {
		invoiceNumbers = append(invoiceNumbers, (i).(map[string]interface{})["invoice_number"].(string))
	}

	table := models.Table{}
	err = db.DB.C("tables").Find(bson.M{"id": newTable}).One(&table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	// update invoices in db
	_, err = db.DB.C("posinvoices").UpdateAll(bson.M{"invoice_number": bson.M{"$in": invoiceNumbers}}, bson.M{"$set": bson.M{"table": table.Number, "table_number": table.ID}})
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}

	// Update Status of new Table
	table.UpdateStatus()
	// Update Status of old Table
	err = db.DB.C("tables").Find(bson.M{"id": oldTable}).One(&table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	table.UpdateStatus()

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, body)

	// get invoices on the new table
	newInvoices := []models.Invoice{}
	err = db.DB.C("posinvoices").Find(bson.M{"table_number": newTable, "is_settled": false}).All(&newInvoices)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	helpers.ReturnSuccessMessage(w, newInvoices)
}

// SplitInvoices splits invoice to new invoices
func SplitInvoices(w http.ResponseWriter, r *http.Request) {
	type ReqBody struct {
		ActionTime          string           `json:"action_time" bson:"action_time"`
		CashierName         string           `json:"cashier_name" bson:"cashier_name"`
		CashierNumber       int              `json:"cashier_number" bson:"cashier_number"`
		EmployeeID          string           `json:"employee_id" bson:"employee_id"`
		Invoices            []models.Invoice `json:"posinvoices" bson:"posinvoices"`
		RCRS                string           `json:"rcrs" bson:"rcrs"`
		TerminalDescription string           `json:"terminal_description" bson:"terminal_description"`
		TerminalID          int              `json:"terminal_id" bson:"terminal_id"`
		TerminalNumber      int              `json:"terminal_number" bson:"terminal_number"`
	}
	body := ReqBody{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	newInvoices := []models.Invoice{}
	for _, i := range body.Invoices {
		req := models.InvoicePOSTRequest{}
		req.ActionTime = body.ActionTime
		req.CashierName = body.CashierName
		req.CashierNumber = body.CashierNumber
		req.EmployeeID = body.EmployeeID
		req.RCRS = body.RCRS
		req.TerminalName = body.TerminalDescription
		req.TerminalID = body.TerminalID
		req.TerminalNumber = body.TerminalNumber
		req.Invoice = i
		// if fdm is enabled submit items to fdm first
		if config.Config.IsFDMEnabled == true {
			// create fdm connection
			conn, err := fdm.Connect(req.RCRS)
			if err != nil {
				helpers.ReturnErrorMessage(w, err.Error())
				return
			}
			defer conn.Close()
			_, err = fdm.Submit(conn, req)
			if err != nil {
				helpers.ReturnErrorMessage(w, err.Error())
				return
			}
			req.Invoice.Events = []models.EJEvent{}
		}

		invoice, err := req.Submit()
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}

		newInvoices = append(newInvoices, invoice)
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, body)

	helpers.ReturnSuccessMessage(w, newInvoices)
}

// WasteAndVoid wastes a lineitem
func WasteAndVoid(w http.ResponseWriter, r *http.Request) {
	invoice := models.Invoice{}
	err := json.NewDecoder(r.Body).Decode(&invoice)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, invoice)

	lineItem := invoice.Items[len(invoice.Items)-1]
	lineItem.SubmittedQuantity = lineItem.Quantity

	invoice.Items[len(invoice.Items)-1] = lineItem

	err = db.DB.C("posinvoices").Update(bson.M{"invoice_number": invoice.InvoiceNumber}, bson.M{"$set": invoice})
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}

	helpers.ReturnSuccessMessage(w, invoice)
}

// ToggleLocking toggle locking of invoices
func ToggleLocking(w http.ResponseWriter, r *http.Request) {
	numbers := strings.Split(r.URL.Query().Get("id"), ",")
	terminalIDStr := r.URL.Query().Get("terminal_id")
	terminalID, _ := strconv.Atoi(terminalIDStr)
	target := r.URL.Query().Get("target")

	invoices := []models.Invoice{}
	err := db.DB.C("posinvoices").Find(bson.M{"invoice_number": bson.M{"$in": numbers}}).All(&invoices)
	if err != nil {
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

// GetInvoiceLatestChanges gets the invoice and checks if it's locked or not
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
