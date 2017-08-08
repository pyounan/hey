package models

type Invoice struct {
	InvoiceNumber string        `json:"invoice_number"`
	Items         []POSLineItem `json:"posinvoicelineitem_set"`
	TableNumber   string        `json:"table"`

	Events []map[string]interface{} `json:"events"`
}
