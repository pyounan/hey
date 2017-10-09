package models

import (
	"log"
	"pos-proxy/db"

	"gopkg.in/mgo.v2/bson"
)

// InvoicePOSTRequest is a model for requests that comes from POST requests
// to make actions on invocies
type InvoicePOSTRequest struct {
	ActionTime     string    `json:"action_time" bson:"action_time"`
	Invoice        Invoice   `json:"posinvoice" bson:"posinvoice"`
	RCRS           string    `json:"rcrs" bson:"rcrs"`
	TerminalID     int       `json:"terminal_id" bson:"terminal_id"`
	TerminalNumber int       `json:"terminal_number" bson:"terminal_number"`
	TerminalName   string    `json:"terminal_description" bson:"terminal_description"`
	EmployeeID     string    `json:"employee_id" bson:"employee_id"`
	CashierName    string    `json:"cashier_name" bson:"cashier_name"`
	CashierNumber  int       `json:"cashier_number" bson:"cashier_number"`
	Postings       []Posting `json:"postings" bson:"postings"`
	// only used for payment
	ChangeAmount float64 `json:"change" bson:"change"`
	IsClosed     bool    `json:"is_closed,omitempty" bson:"is_closed,omitempty"`
	ModalName    string  `json:"modalname" bson:"modalname"`
}

// Submit loops over invoice items, and sets the submitted quantity to quantity
// then updates table status if invoice is not a takeout
func (req *InvoicePOSTRequest) Submit() (Invoice, error) {

	if req.Invoice.InvoiceNumber == "" {
		// create a new invoice with a new invoice number
		invoiceNumber, err := AdvanceInvoiceNumber(req.TerminalID)
		if err != nil {
			return Invoice{}, err
		}
		req.Invoice.InvoiceNumber = invoiceNumber
	}

	items := []POSLineItem{}
	for _, item := range req.Invoice.Items {
		item.SubmittedQuantity = item.Quantity
		items = append(items, item)
	}

	req.Invoice.Items = items

	_, err := db.DB.C("posinvoices").Upsert(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, req.Invoice)
	if err != nil {
		return Invoice{}, err
	}

	// update table status
	table := Table{}
	err = db.DB.C("tables").Find(bson.M{"id": req.Invoice.TableID}).One(&table)
	if err != nil {
		log.Println(err)
	} else {
		table.UpdateStatus()
	}

	return req.Invoice, nil
}
