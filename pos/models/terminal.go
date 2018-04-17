package models

// Terminal swagger:model terminal
// defines attributes of a Terminal entity
type Terminal struct {
	ID int `json:"id" bson:"id"`
	// last number used to generate invoice_number for new invoices
	LastInvoiceID int    `json:"last_invoice_id" bson:"last_invoice_id"`
	Description   string `json:"description" bson:"description"`
	// unique number users see to identify a Terminal
	Number int `json:"number" bson:"number"`
	// the license number set by the Belgian government
	RCRS string `json:"rcrs_number" bson:"rcrs_number"`
	// ID of the store the terminal is linked too
	Store            int    `json:"store" bson:"store"`
	StoreDescription string `json:"store_description" bson:"store_description"`
	IsLocked         bool   `json:"is_locked" bson:"is_locked"`
}
