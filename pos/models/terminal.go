package models

type Terminal struct {
	ID            int    `bson:"id"`
	LastInvoiceID int    `bson:"last_invoice_id"`
	Description   string `bson:"description"`
}
