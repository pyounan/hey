package models

import (
	"pos-proxy/db"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Table struct {
	ID int `json:"id" bson:"id"`
	Number     int       `json:"number" bson:"number"`
	StoreID    int       `json:"store" bson:"store"`
	Status     string    `json:"status" bson:"status"`
	HasInvoice bool      `json:"has_invoice" bson:"has_invoice"`
	UpdatedOn  time.Time `json:"updated_on" bson:"updated_on"`
}

func (t *Table) UpdateStatus() error {
	invoices := &[]Invoice{}
	q := bson.M{"table": t.Number, "is_settled": false}
	err := db.DB.C("posinvoices").Find(q).All(invoices)
	if err != nil {
		return err
	}
	if len(*invoices) > 0 {
		t.Status = "Occupied"
		t.HasInvoice = true
	} else {
		t.Status = "Vacant"
		t.HasInvoice = false
	}

	err = db.DB.C("tables").Update(bson.M{"number": t.Number}, bson.M{"$set": bson.M{"status": t.Status, "has_invoice": t.HasInvoice, "invoices_count": len(*invoices)}})
	if err != nil {
		return err
	}
	return nil
}
