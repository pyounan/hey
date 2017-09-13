package models

type Terminal struct {
	ID               int    `json:"id" bson:"id"`
	LastInvoiceID    int    `json:"last_invoice_id" bson:"last_invoice_id"`
	Description      string `json:"description" bson:"description"`
	Number           int    `json:"number" bson:"number"`
	RCRS             string `json:"rcrs_number" bson:"rcrs_number"`
	Store            int    `json:"store" bson:"store"`
	StoreDescription string `json:"store_description" bson:"store_description"`
	IsLocked         bool   `json:"is_locked" bson:"is_locked"`
}
