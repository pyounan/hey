package printing

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cloudinn/escpos"
	"github.com/cloudinn/escpos/connection"
)

//PrintKitchen to print kitchen recepit
func PrintKitchen(kitchen *KitchenPrint) error {

	printingParams := make(map[int]map[string]int)

	printingParams[80] = make(map[string]int)
	printingParams[80]["width"] = 800
	printingParams[80]["char_per_line"] = 40
	printingParams[80]["item_width"] = 30
	printingParams[80]["store_unit"] = 2
	printingParams[80]["qty"] = 5

	printingParams[76] = make(map[string]int)
	printingParams[76]["width"] = 760
	printingParams[76]["char_per_line"] = 32
	printingParams[76]["item_width"] = 24
	printingParams[76]["store_unit"] = 2
	printingParams[76]["qty"] = 5

	var p *escpos.Printer
	var err error
	if kitchen.Printer.IsUSB {
		p, err = connection.NewConnection("usb", kitchen.Printer.PrinterIP)
		if err != nil {
			for i := 0; i <= 2; i++ {
				time.Sleep(1 * time.Second)
				p, err = connection.NewConnection("usb", kitchen.Printer.PrinterIP)
				if err == nil {
					break
				}
			}
			return err
		}
	} else {
		p, err = connection.NewConnection("network", kitchen.Printer.PrinterIP)
		if err != nil {
			return err
		}
	}

	p.WriteString(CheckLang("Printer ID: " + kitchen.Printer.PrinterID))

	p.WriteString(strings.Repeat("=", printingParams[kitchen.Printer.PaperWidth]["char_per_line"]))

	// p.SetAlign("center")
	p.WriteString(Center("Invoice Number"+": "+kitchen.Invoice.InvoiceNumber) +
		CheckLang("Invoice Number"+": "+kitchen.Invoice.InvoiceNumber))

	p.WriteString(Center("Covers"+": "+fmt.Sprintf("%d", kitchen.Invoice.Pax)) +
		CheckLang("Covers"+": "+fmt.Sprintf("%d", kitchen.Invoice.Pax)))

	if kitchen.Invoice.TableID != nil {
		p.WriteString(Center("Table"+": "+*kitchen.Invoice.TableDetails) +
			CheckLang("Table"+": "+*kitchen.Invoice.TableDetails))
	} else {
		p.WriteString(Center("Takeout") + CheckLang("Takeout"))
	}
	guestName := ""
	if kitchen.Invoice.WalkinName != "" {
		guestName = kitchen.Invoice.WalkinName
	} else if kitchen.Invoice.ProfileDetails != "" {
		guestName = kitchen.Invoice.ProfileDetails
	} else if kitchen.Invoice.RoomDetails != nil {
		guestName = *kitchen.Invoice.RoomDetails
	}
	if guestName != "" {
		p.WriteString(Center("Guest name"+": ") + CheckLang("Guest name"+": "))
		p.WriteString(Center(guestName) + CheckLang(guestName))
	}
	p.WriteString(Center(fmt.Sprintf("%d", kitchen.Cashier.Number)+"/"+kitchen.Cashier.Name) +
		CheckLang(fmt.Sprintf("%d", kitchen.Cashier.Number)+"/"+kitchen.Cashier.Name))

	loc, _ := time.LoadLocation(kitchen.Timezone)
	submittedOn := time.Now().In(loc)
	date := submittedOn.Format(time.RFC1123)
	p.WriteString(Center(date) + date)

	p.WriteString(strings.Repeat("=", printingParams[kitchen.Printer.PaperWidth]["char_per_line"]))

	item := "Item"
	qty := "Qty"
	storeUnit := "Unit"
	p.SetWhiteOnBlack(false)

	p.WriteString(item + Pad(printingParams[kitchen.Printer.PaperWidth]["item_width"]-
		utf8.RuneCountInString(item)) + qty +
		Pad(printingParams[kitchen.Printer.PaperWidth]["qty"]-utf8.RuneCountInString(qty)) +
		storeUnit)

	p.SetWhiteOnBlack(true)
	p.WriteString(strings.Repeat("=", printingParams[kitchen.Printer.PaperWidth]["char_per_line"]))

	for _, item := range kitchen.GropLineItems {
		desc := CheckLang(item.Description)
		qty := CheckLang(fmt.Sprintf("%.2f", item.Quantity))
		baseUnit := CheckLang(item.BaseUnit)
		p.WriteString(desc +
			Pad(printingParams[kitchen.Printer.PaperWidth]["item_width"]-
				utf8.RuneCountInString(desc)) + qty +
			Pad(printingParams[kitchen.Printer.PaperWidth]["qty"]-
				utf8.RuneCountInString(qty)) + baseUnit)

		for _, condiment := range item.CondimentLineItems {
			p.WriteString(CheckLang(condiment.Description))
		}
		if item.CondimentsComment != "" {
			p.WriteString(CheckLang(item.CondimentsComment))
		}
		if item.LastChildInCourse {
			p.WriteString(strings.Repeat("-", printingParams[kitchen.Printer.PaperWidth]["char_per_line"]))
		}
	}
	p.WriteString(strings.Repeat("=", printingParams[kitchen.Printer.PaperWidth]["char_per_line"]))
	p.SetFontSizePoints(40)
	p.WriteString(Pad((30-len("This is not a"))/2) + strings.ToUpper("This is not a"))
	p.WriteString(Pad((30-len("valid tax invoice"))/2) + strings.ToUpper("valid tax invoice"))
	p.Formfeed()
	p.Cut()
	return nil

}
