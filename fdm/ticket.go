package fdm

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Ticket struct {
	ID            bson.ObjectId `json:"-" bson:"_id"`
	TicketNumber  string        `json:"ticket_number" bson:"ticket_number"`
	InvoiceNumber string        `json:"invoice_number" bson:"invoice_number"`
	CashierName   string        `json:"cashier_name" bson:"cashier_name"`
	CashierNumber string        `json:"cashier_number" bson:"cashier_number"`
	TerminalName  string        `json:"terminal_name" bson:"terminal_name"`
	TableNumber   string        `json"table_number" bson:"table_number"`
	RCRS          string        `json:"rcrs" bson:"rcrs"`
	UserID        string        `json:"user_id" bson:"user_id"`
	TotalAmount   float64       `json:"total_amount" bson:"total_amount"`
	Items         []POSLineItem `json:"items" bson:"items"`
	PLUHash       string        `bson:"plu_hash"`
	VATs          []VAT         `bson:"vats"`
	CreatedAt     time.Time     `bson:"created_at"`
}
