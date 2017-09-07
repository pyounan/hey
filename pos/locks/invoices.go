package locks

import (
	"fmt"
	"log"
	"errors"
	"pos-proxy/db"
	"pos-proxy/pos/models"

	"github.com/bsm/redis-lock"
)

func LockInvoices(invoices []models.Invoice, terminalID int) error {
	for _, i := range invoices {
		lockName := fmt.Sprintf("posinvoice_lock_%s_terminal_%d", i.InvoiceNumber, terminalID)
		l, err := lock.ObtainLock(db.Redis, lockName, nil)
		if err != nil {
			log.Println("failed to obtain invoice lock", lockName)
			return err
		} else if l == nil {
			log.Println(err)
			log.Println("failed to obtain invoice lock", lockName)
			return err
		}
		log.Println("obtain invoice lock", lockName)
		ok, err := l.Lock()
		if err != nil{
			log.Println("failed to renew invoice lock", lockName)
			return err
		} else if !ok {
			log.Println("failed to renew invoice lock", lockName)
			return errors.New("failed to renew lock")
		}
	}
	return nil
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
