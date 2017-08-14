package models

import (
	"fmt"
	"pos-proxy/config"
	"pos-proxy/db"
	"time"

	lock "github.com/bsm/redis-lock"
	"github.com/go-redis/redis"
	"gopkg.in/mgo.v2/bson"
)

type InvoicePOSTRequest struct {
	ActionTime    string  `json:"action_time"`
	Invoice       Invoice `json:"posinvoice"`
	RCRS          string  `json:"rcrs"`
	TerminalID    int     `json:"terminal_id"`
	TerminalNumber    int     `json:"terminal_number"`
	TerminalName  string  `json:"terminal_description"`
	CashierID     int  `json:"cashier_id"`
	CashierName   string  `json:"cashier_name"`
	CashierNumber int  `json:"cashier_number"`
	// only used for payment
	Payments     []Payment `json:"postings"`
	ChangeAmount float64   `json:"change_amount"`
	IsClosed     bool      `json:"is_closed,omitempty"`
}

func (req *InvoicePOSTRequest) Submit() error {

	if req.Invoice.InvoiceNumber == "" {
		// Connect to Redis
		client := redis.NewClient(&redis.Options{
			Network: "tcp",
			Addr:    "127.0.0.1:6379",
		})
		defer client.Close()

		lockOpts := &lock.LockOptions{
			WaitTimeout: 3 * time.Second,
		}
		lock, err := lock.ObtainLock(client, fmt.Sprintf("terminal_%d_invoice_number", req.TerminalID), lockOpts)
		if err != nil {
			return err
		} else if lock == nil {
			return fmt.Errorf("Couldn't obtain terminal lock.")
		}
		defer lock.Unlock()

		// create a new invoice with a new invoice number
		invoiceNumber, err := advanceInvoiceNumber(req.TerminalID)
		if err != nil {
			return err
		}
		req.Invoice.InvoiceNumber = invoiceNumber
	}

	q := bson.M{"$set": req.Invoice}
	_, err := db.DB.C("posinvoices").Upsert(bson.M{"invoice_number": req.Invoice.InvoiceNumber}, q)
	if err != nil {
		return err
	}

	if config.Config.IsFDMEnabled == false {
		// log to ej if fdm is not enabled
		go func() {
			// ej.log()
		}()
	}
	return nil
}
