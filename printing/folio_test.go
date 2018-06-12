package printing

import (
	"pos-proxy/config"
	"pos-proxy/income"
	"pos-proxy/pos/models"
	"testing"
	"time"
)

func TestPrintFolio(t *testing.T) {
	config.Config.IsFDMEnabled = true
	invoice := models.Invoice{
		InvoiceNumber: "200-1153",
		Pax:           1,
		// Items: models.POSLineItem[]{
		// 	ID : 2,
		// 	Quantity:2,
		// 	Price:50
		// 	},
		CreatedOn:  "June 7,2018 2:27pm",
		IsSettled:  true,
		WalkinName: "Merit",
		Subtotal:   50,
		// TableID:      7,
		// TableDetails: "AER",
		Total: 50,
		// FDMResponses: [{
		// ProductionNumber:"production_number",
		// VSC :"vsc",
		// Date :time.Now(),
		// TimePeriod :time.Now(),
		// EventLabel:"event_label",
		// TicketCounter:"ticket_counter",
		// TotalTicketCounter :"total_ticket_counter",
		// Signature :"signature",
		// TicketNumber :"ticket_number",
		// TicketActionTime:"ticket_datetime",
		// SoftwareVersion:"software_version",
		// PLUHash :"plu_hash",
		// 	}],
		// Room:        90,
		// RoomDetails: "WER",
		HouseUse:       false,
		Change:         8.5,
		CashierDetails: "Sameh",

		// ClosedOn: time.Now(),
	}
	var id int64 = 1
	invoice.ID = &id
	invoice.TableID = &id
	invoice.Room = &id
	tableDetails := "QQQQ"
	invoice.TableDetails = &tableDetails
	roomDetails := "AAA"
	invoice.RoomDetails = &roomDetails
	now := time.Now()
	invoice.CreatedOn = "June7,20182:27 PM"
	invoice.UpdatedOn = now
	invoice.ClosedOn = &now
	item := models.POSLineItem{
		ID:          2,
		Quantity:    2,
		Price:       50,
		Description: "Cake",
	}
	invoice.Items = append(invoice.Items, item)
	fdmResponse := models.FDMResponse{
		ProductionNumber:   "production_number",
		VSC:                "vsc",
		Date:               time.Now(),
		TimePeriod:         time.Now(),
		EventLabel:         "event_label123456789654789321485498784626587845",
		TicketCounter:      "ticket_counter",
		TotalTicketCounter: "total_ticket_counter",
		Signature:          "signature123456789123456789123456789",
		TicketNumber:       "ticket_number",
		TicketActionTime:   "ticket_datetime",
		SoftwareVersion:    "software_version",
		PLUHash:            "plu_hash123456789987654321123456789",
	}
	invoice.FDMResponses = append(invoice.FDMResponses, fdmResponse)
	store := models.Store{
		ID:            9,
		Code:          "code",
		Description:   "description",
		InvoiceFooter: "invoice_footer",
		InvoiceHeader: "invoice_header",
	}
	cashier := income.Cashier{
		ID:          15,
		Name:        "Xcashier",
		Number:      20,
		EmployeeID:  "employee_id",
		FDMLanguage: "fdm_language,omitempty",
	}
	termnial := models.Terminal{
		ID:               88,
		Description:      "description",
		Number:           78,
		RCRS:             "rcrs_number",
		Store:            9,
		StoreDescription: "store_description",
	}
	company := income.Company{
		Name:       "Company name",
		Address:    "Address company x ",
		PostalCode: "11121",
		VATNumber:  "vat_number",
	}
	printer := models.Printer{
		ID:          6,
		PrinterType: "Epson",
		// PrinterIP:   "192.168.1.220:9100",
		PaperWidth: 80,
		IsDefault:  true,
		TerminalID: 88,
		IsUSB:      false,
	}
	ip := "192.168.1.220:9100"
	// ip := "/dev/usb/lp0"
	printer.PrinterIP = &ip
	folioPrint := FolioPrint{
		Invoice:        invoice,
		Terminal:       termnial,
		Store:          store,
		Cashier:        cashier,
		Company:        company,
		Printer:        printer,
		TotalDiscounts: 10.9,
		Timezone:       "Africa/Cairo",
	}

	err := PrintFolio(&folioPrint)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
