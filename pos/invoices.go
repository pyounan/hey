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
	db.DB.C("tables").Find(bson.M{"number": req.Invoice.TableNumber})
	table.UpdateStatus()

	helpers.ReturnSuccessMessage(w, req)
}

func RefundInvoice(w http.ResponseWriter, r *http.Request) {

}

func Houseuse(w http.ResponseWriter, r *http.Request) {

}

func ChangeTable(w http.ResponseWriter, r *http.Request) {
	body := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	oldTable := body["oldtable"]
	newTable := body["newtable"]
	invoices := body["posinvoices"].([]models.Invoice)
	invoiceNumbers := []string{}
	log.Println("newTable: ", newTable)
	for _, i := range invoices {
		invoiceNumbers = append(invoiceNumbers, i.InvoiceNumber)
	}

	table := models.Table{}
	err = db.DB.C("tables").Find(bson.M{"number": newTable}).One(&table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	// update invoices in db
	_, err = db.DB.C("posinvoices").UpdateAll(bson.M{"invoice_number": bson.M{"$in": invoiceNumbers}}, bson.M{"$set": bson.M{"table": newTable, "table_number": table.ID}})

	// Update Status of new Table
	table.UpdateStatus()
	// Update Status of old Table
	err = db.DB.C("tables").Find(bson.M{"number": oldTable}).One(&table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	table.UpdateStatus()

	// get invoices on the new table
	newInvoices := []models.Invoice{}
	err = db.DB.C("posinvoices").Find(bson.M{"table": newTable}).All(&newInvoices)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	helpers.ReturnSuccessMessage(w, newInvoices)
}

func SplitInvoices(w http.ResponseWriter, r *http.Request) {
	invoices := []models.Invoice{}
	newInvoices := []models.Invoice{}
	for _, i := range invoices {
		req := models.InvoicePOSTRequest{}
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
		}

		invoice, err := req.Submit()
		if err != nil {
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}

		newInvoices = append(newInvoices, invoice)
	}

	helpers.ReturnSuccessMessage(w, newInvoices)
}

func WasteAndVoid(w http.ResponseWriter, r *http.Request) {
	invoice := models.Invoice{}
	err := json.NewDecoder(r.Body).Decode(&invoice)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}

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
