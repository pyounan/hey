package locks

import (
	"fmt"
	"pos-proxy/db"
	"pos-proxy/pos/models"

	"github.com/bsm/redis-lock"
)

func LockInvoices(invoices []models.Invoice, terminalID string) error {
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
