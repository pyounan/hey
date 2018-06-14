package printing

import (
	"pos-proxy/income"
	"pos-proxy/pos/models"
)

//FolioPrint defines the objects and requried data for folio printing
type FolioPrint struct {
	Items          []models.EJEvent
	Invoice        models.Invoice
	Terminal       models.Terminal
	Store          models.Store
	Cashier        income.Cashier
	Company        income.Company
	Printer        models.Printer
	TotalDiscounts float64
	Timezone       string
}

//KitchenPrint defines the objects and requried data for kitchen printing
type KitchenPrint struct {
	GropLineItems []models.EJEvent
	Invoice       models.Invoice
	Printer       Printer
	Cashier       income.Cashier
	Timezone      string
}
type Printer struct {
	ID        int
	PrinterID string
	//type : cashier or kitchen
	PrinterType string
	PrinterIP   string
	PaperWidth  int
	IsDefault   bool
	TerminalID  int
	IsUSB       bool
}

func MaptoPrinter(p models.Printer) Printer {
	printer := Printer{}
	printer.ID = p.ID
	printer.IsDefault = p.IsDefault
	printer.PaperWidth = p.PaperWidth
	printer.PrinterIP = *p.PrinterIP
	printer.PrinterType = p.PrinterType
	printer.PrinterID = p.PrinterID
	printer.IsUSB = p.IsUSB
	printer.TerminalID = p.TerminalID
	return printer
}
