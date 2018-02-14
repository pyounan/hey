package models

import (
	"pos-proxy/db"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Table struct {
	ID            int       `json:"id" bson:"id"`
	Number        int       `json:"number" bson:"number"`
	StoreID       int       `json:"store_id" bson:"store_id"`
	Status        string    `json:"status" bson:"status"`
	HasInvoice    bool      `json:"has_invoice" bson:"has_invoice"`
	UpdatedOn     time.Time `json:"updated_on" bson:"updated_on"`
	InvoicesCount int       `json:"invoices_count" bson:"invoices_count"`
	Description   string    `json:"description" bson:"description"`
	IsActive      bool      `json:"is_active" bson:"is_active"`
}

func (t *Table) UpdateStatus() error {
	invoices := []Invoice{}
	q := bson.M{"table_number": t.ID, "is_settled": false}
	err := db.DB.C("posinvoices").With(db.Session.Copy()).Find(q).All(&invoices)
	if err != nil {
		return err
	}
	if len(invoices) > 0 {
		t.Status = "Occupied"
		t.HasInvoice = true
	} else {
		t.Status = "Occupied"
		t.HasInvoice = false
	}
	updateQuery := bson.M{"updated_on": time.Now().UTC().Format("2006-01-02T15:04:05-0700"), "status": t.Status, "has_invoice": t.HasInvoice, "invoices_count": len(invoices)}

	err = db.DB.C("tables").With(db.Session.Copy()).Update(bson.M{"id": t.ID}, bson.M{"$set": updateQuery})
	if err != nil {
		return err
	}
	return nil
}
