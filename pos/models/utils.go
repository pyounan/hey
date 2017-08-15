package models

import (
	"fmt"
	"pos-proxy/db"
	"time"

	lock "github.com/bsm/redis-lock"
	"github.com/go-redis/redis"

	"gopkg.in/mgo.v2/bson"
)

func advanceInvoiceNumber(terminalID int) (string, error) {
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
