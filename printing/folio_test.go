package printing

import (
	"log"
	"pos-proxy/config"
	"pos-proxy/income"
	"pos-proxy/pos/models"
	"testing"
	"time"
)

func TestPad(t *testing.T) {
	log.Println("test")
	log.Println(Pad(1), len(Pad(1)))
	log.Println(Pad(0), len(Pad(0)))
	log.Println(Pad(10), len(Pad(10)))
}
func TestPrintFolio(t *testing.T) {
	config.Config.IsFDMEnabled = true
	invoice := models.Invoice{
		InvoiceNumber: "200-1153",
		Pax:           1,
		// 	Quantity:2,
		// 	Price:50
		// 	},
		CreatedOn:  "June 7,2018 2:27pm",
		IsSettled:  true,
		WalkinName: "ميريت ايهاب",
		Subtotal:   50,
		// TableID:      7,
		// TableDetails: "AER",
		Total: 1120.50,
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
	tableDetails := "T #26"
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
		Amount:         70.50,
		AuditDate:      "2-05-2018",
		CashierDetails: "Cashier X",
		Cashier:        52,
		CashierID:      80,
		Comments:       "comment",
		// CurrencyID             *int64     `json:"currency_id" bson:"currency_id"`
		// Currency               int        `json:"currency" bson:"currency"`
		// CurrencyDetails        string     `json:"currency_details" bson:"currency_details"`
		Department:        8,
		DepartmentDetails: "Dep details",
		// ForeignAmount          float64    `json:"foreign_amount" bson:"foreign_amount"`
		// FrontendID             string     `json:"frontend_id" bson:"frontend_id"`
		// PosinvoiceID           int        `json:"posinvoice_id" bson:"posinvoice_id"`
		PostingType: "Posting Type",
		// Room                   *int64     `json:"room" bson:"room"`
		// RoomNumber:  nil,
		// RoomDetails: nil,
		// PosPostingInformations []Posting  `json:"pospostinginformations" bson:"pospostinginformations"`
		// PaymentLog             PaymentLog `json:"paymentlog" bson:"paymentlog"`
		// CCType                 *string    `json:"cc_type" bson:"cc_type"`
		// pospostinginformatios only
		// Sign             string   `json:"sign,omitempty" bson:"sign,omitempty"`
		// Type             string   `json:"type,omitempty" bson:"type,omitempty"`
		// Cancelled        bool     `json:"cancelled,omitempty" bson:"cancelled,omitempty"`
		// GatewayResponses: {"hi", "hola", "bye"},
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
		Logo:          "image.png",
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
		// PrinterIP:   "192.168.1.220:9100",
		PaperWidth: 80,
		IsDefault:  true,
		TerminalID: 88,
		IsUSB:      true,
	}
	// ip := "192.168.1.220:9100"
	ip := "/dev/usb/lp0"
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
