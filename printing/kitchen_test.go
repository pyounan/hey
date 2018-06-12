package printing

import (
	"pos-proxy/income"
	"pos-proxy/pos/models"
	"testing"
	"time"
)

func TestPrintKitchen(t *testing.T) {

	invoice := models.Invoice{
		InvoiceNumber: "200-1153",
		Pax:           1,
		CreatedOn:     "June 7,2018 2:27pm",
		WalkinName:    "Merit",
	}
	var id int64 = 1
	invoice.ID = &id
	invoice.TableID = &id
	invoice.Room = &id
	tableDetails := "T #6"
	invoice.TableDetails = &tableDetails
	roomDetails := "AAA"
	invoice.RoomDetails = &roomDetails
	now := time.Now()
	invoice.CreatedOn = "June7,20182:27 PM"
	invoice.UpdatedOn = now
	invoice.ClosedOn = &now
	item := models.POSLineItem{
		ID:                2,
		Quantity:          2,
		Price:             50,
		Description:       "Cake",
		CondimentsComment: "Good",
		LastChildInCourse: true,
		BaseUnit:          "each",
	}
	invoice.Items = append(invoice.Items, item)

	cashier := income.Cashier{
		ID:          15,
		Name:        "Xcashier",
		Number:      20,
		EmployeeID:  "employee_id",
		FDMLanguage: "fdm_language,omitempty",
	}

	printer := models.Printer{
		ID:          6,
		PrinterType: "Epson",
		PrinterID:   "123456",
		PaperWidth:  80,
		IsDefault:   true,
		TerminalID:  88,
		IsUSB:       true,
	}
	ip := "192.168.1.114:9100"
	// ip := "/dev/usb/lp0"
	printer.PrinterIP = &ip
	kitchenPrint := KitchenPrint{
		Invoice:  invoice,
		Cashier:  cashier,
		Printer:  printer,
		Timezone: "Africa/Cairo",
	}
	err := PrintKitchen(&kitchenPrint)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

}
