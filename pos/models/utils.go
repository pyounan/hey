package models

import (
	"fmt"
	"pos-proxy/db"
	"time"

	lock "github.com/bsm/redis-lock"
	"github.com/go-redis/redis"

	"gopkg.in/mgo.v2/bson"
)

// AdvanceInvoiceNumber increases the terminal.last_invoice_id by one and
// returns a new invoice number.
func AdvanceInvoiceNumber(terminalID int) (string, error) {
	// Connect to Redis
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	defer client.Close()
	lockOpts := &lock.LockOptions{
		WaitTimeout: 3 * time.Second,
	}
	lock, err := lock.ObtainLock(client, fmt.Sprintf("terminal_%d_invoice_number", terminalID), lockOpts)
	if err != nil {
		return "", err
	} else if lock == nil {
		return "", fmt.Errorf("couldn't obtain terminal lock")
	}
	defer lock.Unlock()

	invoiceNumber := ""
	terminal := Terminal{}
	err = db.DB.C("terminals").Find(bson.M{"id": terminalID}).One(&terminal)
	if err != nil {
		return "", err
	}

	id := terminal.LastInvoiceID + 1
	invoiceNumber = fmt.Sprintf("%d-%d", terminal.ID, id)
	err = db.DB.C("terminals").Update(bson.M{"id": terminal.ID},
		bson.M{"$set": bson.M{"last_invoice_id": terminal.LastInvoiceID + 1}})
	if err != nil {
		return "", err
	}
	return invoiceNumber, nil
}

// SummarizeVAT calculates the total net_amount and vat_amount of each
// VAT rate
func SummarizeVAT(items *[]POSLineItem) map[string]VATSummary {
	summary := make(map[string]VATSummary)
	rates := []string{"A", "B", "C", "D", "Total"}
	for _, r := range rates {
		summary[r] = VATSummary{}
		summary[r]["net_amount"] = 0
		summary[r]["vat_amount"] = 0
		summary[r]["taxable_amount"] = 0
	}
	for _, item := range *items {
		summary[item.VAT]["net_amount"] += item.NetAmount
		summary[item.VAT]["vat_amount"] += item.NetAmount * item.VATPercentage / 100
		summary[item.VAT]["taxable_amount"] += item.Price

		summary["Total"]["net_amount"] += item.NetAmount
		summary["Total"]["vat_amount"] += item.NetAmount * item.VATPercentage / 100
		summary["Total"]["taxable_amount"] += item.Price
	}

	return summary
}
