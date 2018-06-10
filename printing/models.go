package printing

import (
	"pos-proxy/income"
	"pos-proxy/pos/models"
)

type FolioPrint struct {
	Invoice        models.Invoice
	Terminal       models.Terminal
	Store          models.Store
	Cashier        income.Cashier
	Company        income.Company
	Printer        models.Printer
	TotalDiscounts float32
}
