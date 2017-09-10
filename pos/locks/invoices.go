package locks

import (
	"errors"
	"fmt"
	"log"
	"pos-proxy/db"
	"pos-proxy/pos/models"

	"github.com/bsm/redis-lock"
	"github.com/go-redis/redis"
)

func LockInvoices(invoices []models.Invoice, terminalID int) (int64, error) {
	lockName := fmt.Sprintf("posinvoices_lock")
	var otherTerminal int64 = 0
	client := db.Redis
	l, err := lock.ObtainLock(client, lockName, nil)
	if err != nil {
		log.Println("failed to obtain invoice lock", lockName)
		return otherTerminal, err
	} else if l == nil {
		log.Println(err)
		log.Println("failed to obtain invoice lock", lockName)
		return otherTerminal, err
	}
	log.Println("obtain invoice lock", lockName)
	ok, err := l.Lock()
	defer l.Unlock()
	if err != nil {
		log.Println("failed to renew invoice lock", lockName)
		return otherTerminal, err
	} else if !ok {
		log.Println("failed to renew invoice lock", lockName)
		return otherTerminal, errors.New("failed to renew lock")
	}
	keys := []string{}
	for _, i := range invoices {
		keys = append(keys, "invoice_"+i.InvoiceNumber)
	}
	err = client.Watch(func(tx *redis.Tx) error {

		for _, key := range keys {
			_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
				n, err := pipe.Get(key).Int64()
				if err != nil && err != redis.Nil {
					return err
				} else if n != int64(terminalID) {
					otherTerminal = n
					return errors.New("Invoice key already exists!!")
				}
				pipe.Set(key, terminalID, 0)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}, keys...)

	return otherTerminal, err
}

func UnlockInvoices(invoices []models.Invoice) error {
	for _, i := range invoices {
		l, err := lock.ObtainLock(db.Redis, fmt.Sprintf("posinvoice_%f", i.InvoiceNumber), nil)
		if err != nil {
			return err
		}
		l.Unlock()
	}
	return nil
}
