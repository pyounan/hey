package printing

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/abadojack/whatlanggo"
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
	printingParams[76]["item_width"] = 25
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

	p.WriteString(Center("Invoice Number"+": "+kitchen.Invoice.InvoiceNumber,
		printingParams[kitchen.Printer.PaperWidth]["width"]) +
		CheckLang(Translate("Invoice number")+": "+kitchen.Invoice.InvoiceNumber))

	p.WriteString(Center("Covers"+": "+fmt.Sprintf("%d", kitchen.Invoice.Pax),
		printingParams[kitchen.Printer.PaperWidth]["width"]) +
		CheckLang(Translate("Covers")+": "+fmt.Sprintf("%d", kitchen.Invoice.Pax)))
	if kitchen.Invoice.TableID != nil {
		p.WriteString(Center("Table"+": "+*kitchen.Invoice.TableDetails,
			printingParams[kitchen.Printer.PaperWidth]["width"]) +
			CheckLang(Translate("Table")+": "+*kitchen.Invoice.TableDetails))
	} else {
		p.WriteString(Center("Takeout",
			printingParams[kitchen.Printer.PaperWidth]["width"]) +
			CheckLang(Translate("Takeout")))
	}
	guestName := ""
	if kitchen.Invoice.WalkinName != "" {
		guestName = kitchen.Invoice.WalkinName
	} else if kitchen.Invoice.ProfileDetails != "" {
		guestName = kitchen.Invoice.ProfileDetails
	} else if kitchen.Invoice.RoomDetails != nil {
		guestName = *kitchen.Invoice.RoomDetails
	}
	guestNameTrans := Translate("Guest name")
	info := whatlanggo.Detect(guestNameTrans)
	if guestName != "" {
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(Center(guestNameTrans+": "+guestName,
				printingParams[kitchen.Printer.PaperWidth]["width"]) +
				CheckLang(guestName) + ": " + CheckLang(guestNameTrans))

		} else {
			p.WriteString(Center(guestNameTrans+": "+guestName,
				printingParams[kitchen.Printer.PaperWidth]["width"]) +
				CheckLang(guestNameTrans) + ": " + CheckLang(guestName))
		}
	}
	p.WriteString(Center(fmt.Sprintf("%d", kitchen.Cashier.Number)+"/"+
		kitchen.Cashier.Name, printingParams[kitchen.Printer.PaperWidth]["width"]) +
		CheckLang(fmt.Sprintf("%d", kitchen.Cashier.Number)+"/"+kitchen.Cashier.Name))

	loc, _ := time.LoadLocation(kitchen.Timezone)
	submittedOn := time.Now().In(loc)
	date := submittedOn.Format(time.RFC1123)
	p.WriteString(Center(date, printingParams[kitchen.Printer.PaperWidth]["width"]) +
		date)

	p.WriteString(strings.Repeat("=",
		printingParams[kitchen.Printer.PaperWidth]["char_per_line"]))

	item := CheckLang(Translate("Item"))
	qty := CheckLang(Translate("Qty"))
	storeUnit := CheckLang(Translate("Unit"))
	p.SetWhiteOnBlack(false)
	if printingParams[kitchen.Printer.PaperWidth]["width"] == 760 {

		p.SetFontSizePoints(28)
	}
	info = whatlanggo.Detect(item)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(storeUnit + Pad(printingParams[kitchen.Printer.PaperWidth]["qty"]-
			utf8.RuneCountInString(storeUnit)) + qty +
			Pad(printingParams[kitchen.Printer.PaperWidth]["item_width"]-
				utf8.RuneCountInString(qty)) + item)

	} else {
		p.WriteString(item + Pad(printingParams[kitchen.Printer.PaperWidth]["item_width"]-
			utf8.RuneCountInString(item)) + qty +
			Pad(printingParams[kitchen.Printer.PaperWidth]["qty"]-
				utf8.RuneCountInString(qty)) + storeUnit)
	}
	p.SetWhiteOnBlack(true)
	p.SetFontSizePoints(30)
	p.WriteString(strings.Repeat("=",
		printingParams[kitchen.Printer.PaperWidth]["char_per_line"]))

	for _, item := range kitchen.GropLineItems {
		desc := CheckLang(item.Description)
		qty := CheckLang(fmt.Sprintf("%.2f", item.Quantity))
		baseUnit := CheckLang(item.BaseUnit)
		if printingParams[kitchen.Printer.PaperWidth]["width"] == 760 {

			p.SetFontSizePoints(28)
		}
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(baseUnit +
				Pad(printingParams[kitchen.Printer.PaperWidth]["qty"]-
					utf8.RuneCountInString(baseUnit)) + qty +
				Pad(printingParams[kitchen.Printer.PaperWidth]["item_width"]-
					utf8.RuneCountInString(desc)) + desc)

		} else {
			p.WriteString(desc +
				Pad(printingParams[kitchen.Printer.PaperWidth]["item_width"]-
					utf8.RuneCountInString(desc)) + qty +
				Pad(printingParams[kitchen.Printer.PaperWidth]["qty"]-
					utf8.RuneCountInString(qty)) + baseUnit)
		}

		for _, condiment := range item.CondimentLineItems {
			p.WriteString(CheckLang(condiment.Description))
		}
		if item.CondimentsComment != "" {
			if whatlanggo.Scripts[info.Script] == "Arabic" {
				p.WriteString(Pad(37-utf8.RuneCountInString(item.CondimentsComment)) +
					CheckLang(item.CondimentsComment))
			} else {
				p.WriteString(CheckLang(item.CondimentsComment))
			}
		}
		p.SetFontSizePoints(30)
		if item.LastChildInCourse {
			p.WriteString(strings.Repeat("-", printingParams[kitchen.Printer.PaperWidth]["char_per_line"]))
		}
	}
	p.WriteString(strings.Repeat("=",
		printingParams[kitchen.Printer.PaperWidth]["char_per_line"]))
	p.SetFontSizePoints(40)
	p.SetImageHight(42)
	p.WriteString(Pad((30-len("This is not a"))/2) + strings.ToUpper(CheckLang(Translate("This is not a"))))
	p.WriteString(Pad((30-len("valid tax invoice"))/2) +
		strings.ToUpper(CheckLang(Translate("valid tax invoice"))))
	p.Formfeed()
	p.Cut()
	return nil

}
