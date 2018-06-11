package printing

import (
	"pos-proxy/income"
	"pos-proxy/pos/models"
)

//FolioPrint defines the objects and requried data for folio printing
type FolioPrint struct {
	Invoice        models.Invoice
	Terminal       models.Terminal
	Store          models.Store
	Cashier        income.Cashier
	Company        income.Company
	Printer        models.Printer
	TotalDiscounts float32
}

//KitchenPrint defines the objects and requried data for kitchen printing
type KitchenPrint struct {
	Invoice models.Invoice
	Printer models.Printer
	Cashier income.Cashier
}
