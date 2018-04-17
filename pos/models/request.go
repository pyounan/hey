package models

import (
	"log"
	"pos-proxy/db"

	"gopkg.in/mgo.v2/bson"
)

// InvoicePOSTRequest is a model for requests that comes from POST requests
// to make actions on invocies
// swagger:model invoicePOSTRequest
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
			log.Println("ERROR: ", err.Error())
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

	session := db.Session.Copy()
	defer session.Close()
	_, err := db.DB.C("posinvoices").With(session).Upsert(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, req.Invoice)
	if err != nil {
		log.Println("ERROR: ", err.Error())
		return Invoice{}, err
	}

	// update table status
	if req.Invoice.TableID != nil {
		table := Table{}
		err = db.DB.C("tables").With(session).Find(bson.M{"id": *req.Invoice.TableID}).One(&table)
		if err != nil {
			log.Println(err)
		} else {
			table.UpdateStatus()
		}

	}

	return req.Invoice, nil
}

// CancelPostingsRequest swagger:model cancelPostingsRequest
// defines the body of a CancelPosting request
type CancelPostingsRequest struct {
	PostingsIDs []string `json:"frontend_ids" bson:"frontend_ids"`
	CashierID   int      `json:"poscashier_id" bson:"poscashier_id"`
	Posinvoice  Invoice  `json:"posinvoice" bson:"posinvoice"`
}

// BulkSubmitRequest swagger:model bulkSubmitRequest
// defines the body of a BulkSubmit request
type BulkSubmitRequest struct {
	Invoices              []Invoice `json:"posinvoices" bson:"posinvoices"`
	RCRS                  string    `json:"rcrs" bson:"rcrs"`
	TerminalID            int       `json:"terminal_id" bson:"terminal_id"`
	TerminalNumber        int       `json:"terminal_number" bson:"terminal_number"`
	TerminalName          string    `json:"terminal_description" bson:"terminal_description"`
	EmployeeID            string    `json:"employee_id" bson:"employee_id"`
	OriginalInvoiceNumber string    `json:"original_invoice_number" bson:"original_invoice_number"`
	DepartmentID          int       `json:"department" bson:"department"`
	Posting               Posting   `json:"posting" bson:"posting"`
	CashierName           string    `json:"cashier_name" bson:"cashier_name"`
	CashierNumber         int       `json:"cashier_number" bson:"cashier_number"`
	Type                  string    `json:"type" bson:"type"`
	ActionTime            string    `json:"action_time" bson:"action_time"`
}

// BulkSubmitResponse swagger:model bulkSubmitResponse
// defines the body of a BulkSubmit response
type BulkSubmitResponse struct {
	Status int `json:"status"`
}

type OriginalInvoiceData struct {
	ID            int    `json:"id" bson:"id"`
	InvoiceNumber string `json:"invoice_number" bson:"invoice_number"`
}

// RefundInvoiceRequest swagger:model refundInvoiceRequest
// defines the request body of RefundInvoice API
type RefundInvoiceRequest struct {
	RCRS            string              `json:"rcrs" bson:"rcrs"`
	TerminalID      int                 `json:"terminal_id" bson:"terminal_id"`
	TerminalNumber  int                 `json:"terminal_number" bson:"terminal_number"`
	TerminalName    string              `json:"terminal_description" bson:"terminal_description"`
	EmployeeID      string              `json:"employee_id" bson:"employee_id"`
	NewInvoice      Invoice             `json:"new_posinvoice" bson:"new_posinvoice"`
	OriginalInvoice OriginalInvoiceData `json:"posinvoice" bson:"posinvoice"`
	DepartmentID    int                 `json:"department" bson:"department"`
	Posting         Posting             `json:"posting" bson:"posting"`
	CashierName     string              `json:"cashier_name" bson:"cashier_name"`
	CashierNumber   int                 `json:"cashier_number" bson:"cashier_number"`
	Type            string              `json:"type" bson:"type"`
	ActionTime      string              `json:"action_time" bson:"action_time"`
}

// RefundInvoiceResponse swagger:model refundInvoiceResponse
// defines the response body of RefundInvoice API
type RefundInvoiceResponse struct {
	NewInvoice      Invoice   `json:"new_invoice" bson:"new_invoice"`
	OriginalInvoice Invoice   `json:"original_invoice" bson:"original_invoice"`
	Postings        []Posting `json:"postings" bson:"postings"`
}

// ChangeTableRequest swagger:model changeTableRequest
// defines the request body of ChangeTable API
type ChangeTableRequest struct {
	OldTable int       `json:"oldtable" bson:"oldtable"`
	NewTable int       `json:"newtable" bson:"newtable"`
	Invoices []Invoice `json:"posinvoices" bson:"posinvoices"`
}

// SplitInvoiceRequest swagger:model splitInvoiceRequest
// defines the request body of SplitInvoice API
type SplitInvoiceRequest struct {
	ActionTime          string    `json:"action_time" bson:"action_time"`
	CashierName         string    `json:"cashier_name" bson:"cashier_name"`
	CashierNumber       int       `json:"cashier_number" bson:"cashier_number"`
	EmployeeID          string    `json:"employee_id" bson:"employee_id"`
	Invoices            []Invoice `json:"posinvoices" bson:"posinvoices"`
	RCRS                string    `json:"rcrs" bson:"rcrs"`
	TerminalDescription string    `json:"terminal_description" bson:"terminal_description"`
	TerminalID          int       `json:"terminal_id" bson:"terminal_id"`
	TerminalNumber      int       `json:"terminal_number" bson:"terminal_number"`
	Events              []string  `json:"events" bson:"events"`
}
