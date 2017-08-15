package models

import (
	"pos-proxy/config"
	"pos-proxy/db"

	"gopkg.in/mgo.v2/bson"
)

type InvoicePOSTRequest struct {
	ActionTime     string  `json:"action_time"`
	Invoice        Invoice `json:"posinvoice"`
	RCRS           string  `json:"rcrs"`
	TerminalID     int     `json:"terminal_id"`
	TerminalNumber int     `json:"terminal_number"`
	TerminalName   string  `json:"terminal_description"`
	EmployeeID     string  `json:"employee_id"`
	CashierName    string  `json:"cashier_name"`
	CashierNumber  int     `json:"cashier_number"`
	// only used for payment
	Payments     []Payment `json:"postings"`
	ChangeAmount float64   `json:"change_amount"`
	IsClosed     bool      `json:"is_closed,omitempty"`
}

func (req *InvoicePOSTRequest) Submit() (Invoice, error) {

	if req.Invoice.InvoiceNumber == "" {
		// create a new invoice with a new invoice number
		invoiceNumber, err := advanceInvoiceNumber(req.TerminalID)
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

	q := bson.M{"$set": req.Invoice}
	_, err := db.DB.C("posinvoices").Upsert(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, q)
	if err != nil {
		return Invoice{}, err
	}

	if config.Config.IsFDMEnabled == false {
		// log to ej if fdm is not enabled
		go func() {
			// ej.log()
		}()
	}
	return req.Invoice, nil
}
