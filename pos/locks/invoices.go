package locks

import (
	"errors"
	"fmt"
	"log"
	"pos-proxy/db"
	"pos-proxy/pos/models"
	"strconv"

	"github.com/bsm/redis-lock"
	"github.com/go-redis/redis"
)

// LockInvoices creates keys in Redis for given invoices and lock
// them to the passed terminal id.
func LockInvoices(invoices []models.Invoice, terminalID int) (int, error) {
	lockName := fmt.Sprintf("posinvoices_lock")
	var otherTerminal int
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
	if err != nil {
		log.Println("failed to renew invoice lock", lockName)
		return otherTerminal, err
	} else if !ok {
		log.Println("failed to renew invoice lock", lockName)
		return otherTerminal, errors.New("failed to renew lock")
	}
	defer l.Unlock()

	keys := []string{}
	for _, i := range invoices {
		keys = append(keys, "invoice_"+i.InvoiceNumber)
	}
	err = client.Watch(func(tx *redis.Tx) error {

		for _, key := range keys {
			_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
				log.Println("PIPELINE KEY", key)
				val, err := tx.Get(key).Result()
				if err != nil && err != redis.Nil {
					return err
				}
				log.Println("REDIS VALUE", val)
				if err == redis.Nil {
					log.Println("INVOICE IS NOT LOCKED")
					pipe.Set(key, terminalID, 0)
					return nil
				}
				n, err := strconv.Atoi(val)
				if err != nil {
					log.Println("SOME ERROR", err, n, int64(terminalID))
					return err
				} else if n != terminalID {
					otherTerminal = n
					log.Println("PIPELINE TERMINAL IS LOCKED")
					return errors.New("invoice key already exists")
				}
				return nil
			})
			if err != nil && err != redis.Nil {
				log.Println("PIPELINE ERROR", err)
				return err
			}
		}
		return nil
	}, keys...)

	log.Println("WATCH ERR", err)
	if err != nil && err != redis.Nil {
		return otherTerminal, err
	}
	return otherTerminal, nil
}

// UnlockInvoices deletes keys of given invoices from Redis,
// and make the invoices available again to be picked up by
// other terminals.
func UnlockInvoices(invoices []models.Invoice) {
	client := db.Redis
	for _, i := range invoices {
		client.Del("invoice_" + i.InvoiceNumber)
	}
}
