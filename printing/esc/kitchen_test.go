package esc

import (
	"pos-proxy/income"
	"pos-proxy/pos/models"
	"pos-proxy/printing"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

func TestPrintKitchen(t *testing.T) {

	invoice := models.Invoice{
		InvoiceNumber: "200-1153",
		Pax:           1,
		CreatedOn:     "June 7,2018 2:27pm",
		WalkinName:    "ميريت ايهاب",
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
	item := models.EJEvent{
		Quantity:          2,
		Price:             50,
		Description:       "تشيز كيك بلو بيرى",
		CondimentsComment: "جيد",
		LastChildInCourse: true,
		BaseUnit:          "الكل",
	}
	item2 := models.EJEvent{
		Quantity:          2,
		Price:             50,
		Description:       "chocolate cake",
		CondimentsComment: "A1",
		LastChildInCourse: true,
		BaseUnit:          "each",
	}
	items := []models.EJEvent{}
	items = append(items, item)
	items = append(items, item2)

	cashier := income.Cashier{
		ID:          15,
		Name:        "كريم أحمد",
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
		IsUSB:       false,
		PrinterIP:   aws.String("192.168.1.220"),
		// PrinterIP: aws.String("/dev/usb/lp0"),
	}
	kitchenPrint := printing.KitchenPrint{
		Invoice:       invoice,
		Cashier:       cashier,
		Printer:       printer,
		GropLineItems: items,
		Timezone:      "Africa/Cairo",
	}
	esc := Esc{}
	err := esc.PrintKitchen(&kitchenPrint)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

}
