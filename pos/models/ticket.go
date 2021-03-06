package models

import (
	"gopkg.in/mgo.v2/bson"
)

type FDMTicket struct {
	ID            bson.ObjectId `json:"-" bson:"_id"`
	TicketNumber  string        `json:"ticket_number" bson:"ticket_number"`
	InvoiceNumber string        `json:"invoice_number" bson:"invoice_number"`
	CashierName   string        `json:"cashier_name" bson:"cashier_name"`
	CashierNumber string        `json:"cashier_number" bson:"cashier_number"`
	TerminalName  string        `json:"terminal_name" bson:"terminal_name"`
	TableNumber   string        `json:"table_number" bson:"table_number"`
	RCRS          string        `json:"rcrs" bson:"rcrs"`
	UserID        string        `json:"user_id" bson:"user_id"`
	TotalAmount   float64       `json:"total_amount" bson:"total_amount"`
	Items         []EJEvent     `json:"items" bson:"items"`
	PLUHash       string        `bson:"plu_hash"`
	VATs          []VAT         `bson:"vats"`
	ActionTime    string        `json:"action_time" bson:"action_time"`
	// only for payment, used in ej
	Postings     []Posting `bson:"postings"`
	ChangeAmount float64   `bson:"change_amount"`
	IsClosed     bool      `bson:"is_closed"`
}
