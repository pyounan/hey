package models

import (
	"fmt"
	"pos-proxy/db"

	"gopkg.in/mgo.v2/bson"
)

func advanceInvoiceNumber(terminalID int) (string, error) {
	invoiceNumber := ""
	terminal := make(map[string]interface{})
	err := db.DB.C("terminals").Find(bson.M{"id": terminalID}).One(&terminal)
	if err != nil {
		return "", err
	}
	id := (terminal["last_invoice_id"].(int)) + 1
	invoiceNumber = fmt.Sprintf("%d-%d", terminal["id"], id)
	return invoiceNumber, nil
}
