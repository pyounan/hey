package fdm

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Ticket struct {
	ID            bson.ObjectId `json:"-" bson:"_id"`
	TicketNumber  string        `json:"ticket_number" bson:"ticket_number"`
	InvoiceNumber string        `json:"invoice_number" bson:"invoice_number"`
	RCRS          string        `json:"rcrs" bson:"rcrs"`
	UserID        string        `json:"user_id" bson:"user_id"`
	TotalAmount   float64       `json:"total_amount" bson:"total_amount"`
	Items         []POSLineItem `json:"items,omitempty" bson:"items"`
	IsSubmitted   bool          `json:"is_submitted,omitempty" bson:"is_sumbitted"`
	IsPaid        bool          `json:is_paid,omitempty" bson:"is_paid"`
	PLUHash       string        `bson:"plu_hash"`
	VATs          []VAT         `bson:"vats"`
	CreatedAt     time.Time     `bson:"created_at"`
}
