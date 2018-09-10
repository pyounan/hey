package epsonxml

import (
	"pos-proxy/config"
	"pos-proxy/income"
	"pos-proxy/pos/models"
	"pos-proxy/printing"
	"testing"
	"time"
)

func TestPrintFolio(t *testing.T) {
	config.Config.IsFDMEnabled = false
	invoice := models.Invoice{
		InvoiceNumber:  "200-1153",
		Pax:            1,
		CreatedOn:      "June 7,2018 2:27pm",
		IsSettled:      true,
		WalkinName:     "ميريت ايهاب",
		Subtotal:       50,
		Total:          1120.50,
		HouseUse:       false,
		Change:         8.5,
		CashierDetails: "Sameh",
	}
	var id int64 = 1
	invoice.ID = &id
	// invoice.TableID = &id
	invoice.Room = &id
	// tableDetails := "T #26"
	// invoice.TableDetails = &tableDetails
	roomDetails := "AAA"
	invoice.RoomDetails = &roomDetails
	now := time.Now()
	invoice.CreatedOn = "June7,20182:27 PM"
	invoice.UpdatedOn = now
	invoice.ClosedOn = &now
	item := models.POSLineItem{
		ID:          2,
		Quantity:    2,
		Price:       999.50,
		Description: "فوتوتشيني ألفريدو",
	}
	invoice.Items = append(invoice.Items, item)
	fdmResponse := models.FDMResponse{
		ProductionNumber:   "2000",
		VSC:                "vsc",
		Date:               time.Now(),
		TimePeriod:         time.Now(),
		EventLabel:         "event_label123456789654789321485498784626587845",
		TicketCounter:      "ticket_counter",
		TotalTicketCounter: "10000",
		Signature:          "signature123456789123456789123456789",
		TicketNumber:       "ticket_number",
		TicketActionTime:   "ticket_datetime",
		SoftwareVersion:    "software_version",
		PLUHash:            "plu_hash123456789987654321123456789",
	}

	postings := models.Posting{
		Amount:            70.50,
		AuditDate:         "2-05-2018",
		CashierDetails:    "Cashier X",
		Cashier:           52,
		CashierID:         80,
		Comments:          "comment",
		Department:        8,
		DepartmentDetails: "Dep details",
		PostingType:       "Posting Type",
	}
	var roomnum int64 = 84
	postings.RoomNumber = &roomnum

	invoice.Postings = append(invoice.Postings, postings)
	invoice.FDMResponses = append(invoice.FDMResponses, fdmResponse)
	store := models.Store{
		ID:            9,
		Code:          "code",
		Description:   "description",
		InvoiceFooter: "invoice_footer",
		InvoiceHeader: "invoice_header",
		Logo:          "https://image.ibb.co/jzbjbU/images.png",
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
		Name:       "Company X",
		Address:    "Address company x ",
		PostalCode: "11121",
		City:       "Cairo",
		VATNumber:  "vat_number",
	}
	printer := models.Printer{
		ID:          6,
		PrinterType: "Epson",
		PaperWidth:  80,
		IsDefault:   true,
		TerminalID:  88,
		IsUSB:       false,
	}
	ip := "192.168.1.220:9100"
	// ip := "/dev/usb/lp0"
	printer.PrinterIP = &ip
	folioPrint := printing.FolioPrint{
		Invoice:        invoice,
		Terminal:       termnial,
		Store:          store,
		Cashier:        cashier,
		Company:        company,
		Printer:        printer,
		TotalDiscounts: 10.9,
		Timezone:       "Africa/Cairo",
	}

	epsonxml := Epsonxml{}
	err := epsonxml.PrintFolio(&folioPrint)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
