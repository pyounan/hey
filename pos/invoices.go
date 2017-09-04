package pos

import (
	"encoding/json"
	"log"
	"net/http"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/fdm"
	"pos-proxy/pos/models"
	"pos-proxy/syncer"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

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
		} else {
			// q[key] = val
		}
	}
	invoices := []models.Invoice{}
	err := db.DB.C("posinvoices").Find(q).All(&invoices)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	resp := bson.M{}
	resp["count"] = len(invoices)
	resp["results"] = invoices
	helpers.ReturnSuccessMessage(w, resp)
}

func GetInvoice(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["invoice_nubmer"]
	q["invoice_number"] = id

	invoice := make(map[string]interface{})
	err := db.DB.C("posinvoices").Find(q).One(&invoice)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
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

		fdmResponses = append(fdmResponses, responses...)
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)
	req.Invoice.Events = []models.Event{}
	invoice, err := req.Submit()
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	invoice.FDMResponses = fdmResponses
	req.Invoice = invoice
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)

	helpers.ReturnSuccessMessage(w, invoice)
}

func UpdateInvoice(w http.ResponseWriter, r *http.Request) {

}

func LockInvoice(w http.ResponseWriter, r *http.Request) {

}

func UnlockInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceNumber := vars["invoice_number"]
	invoice := &models.Invoice{}
	err := db.DB.C("posinvoices").Find(bson.M{"invoice_number": invoiceNumber}).One(invoice)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
	}
	helpers.ReturnSuccessMessage(w, true)
}

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

	req.Invoice.Events = []models.Event{}
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)

	invoice, err := req.Submit()
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
	}

	invoice.FDMResponses = fdmResponses

	helpers.ReturnSuccessMessage(w, invoice)
}

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
	}

	req.Invoice.Events = []models.Event{}
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, req)

	req.Postings[0].PosPostingInformations = []models.Posting{}
	req.Postings[0].PosPostingInformations = append(req.Postings[0].PosPostingInformations, models.Posting{})
	req.Postings[0].PosPostingInformations[0].Comments = ""
	req.Invoice.Postings = req.Postings
	req.Invoice.FDMResponses = fdmResponses
	req.Invoice.IsSettled = true
	req.Invoice.PaidAmount = req.Invoice.Total

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

func RefundInvoice(w http.ResponseWriter, r *http.Request) {
	type ReqBody struct {
		Invoice models.Invoice `json:"posinvoice" bson:"posinvoice"`
		DepartmentID int `json:"department" bson:"department"`
		Posintg models.Posting `json:"posting" bson:"posting"`
		OldInvoiceID int `json:"old_posinvoice" bson:"old_posinvoice"`
		CashierID int `json:"cashier_id" bson:"cashier_id"`
		Type string `json:"type" bson:"type"`
	}
	body := ReqBody{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	defer r.Body.Close()

	terminalIDStr := r.URL.Query().Get("terminal_id")
	terminalID, _ := strconv.Atoi(terminalIDStr)
	invoiceNumber, err := models.AdvanceInvoiceNumber(terminalID)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	body.Invoice.InvoiceNumber = invoiceNumber

	fdmResponses := []models.FDMResponse{}

	if config.Config.IsFDMEnabled == true {
		// create fdm connection
		req := models.InvoicePOSTRequest{}
		// TOFIX: fix req body values
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
	body.Invoice.Events = []models.Event{}
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, body)

	type RespBody struct {
		NewInvoice models.Invoice `json:"new_invoice" bson:"new_invoice"`
		OldInvoice models.Invoice `json:"old_invoice" bson:"old_invoice"`
		Postings []models.Posting `json:"postings" bson:"postings"`
	}
	resp := &RespBody{}
	resp.NewInvoice = body.Invoice
	resp.OldInvoice = models.Invoice{}
	db.DB.C("posinvoices").Find(bson.M{"invoice_number": body.OldInvoiceID}).One(&resp.OldInvoice)
	helpers.ReturnSuccessMessage(w, resp)

}

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

func SplitInvoices(w http.ResponseWriter, r *http.Request) {
	type ReqBody struct {
		ActionTime string `json:"action_time" bson:"action_time"`
		CashierName string `json:"cashier_name" bson:"cashier_name"`
		CashierNumber int `json:"cashier_number" bson:"cashier_number"`
		EmployeeID string `json:"employee_id" bson:"employee_id"`
		Invoices []models.Invoice `json:"posinvoices" bson:"posinvoices"`
		RCRS string `json:"rcrs" bson:"rcrs"`
		TerminalDescription string `json:"terminal_description" bson:"terminal_description"`
		TerminalID int `json:"terminal_id" bson:"terminal_id"`
		TerminalNumber int `json:"terminal_number" bson:"terminal_number"`
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
			req.Invoice.Events = []models.Event{}
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

func ToggleLocking(w http.ResponseWriter, r *http.Request) {
	helpers.ReturnSuccessMessage(w, true)
}
