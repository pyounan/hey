package models

type Event struct {
	Item POSLineItem `json:"item"`
}

type Invoice struct {
	InvoiceNumber string        `json:"invoice_number"`
	Items         []POSLineItem `json:"posinvoicelineitem_set"`
	TableNumber   int        `json:"table"`

	Events []Event `json:"events"`
}
