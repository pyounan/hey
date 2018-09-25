package esc

import (
	"fmt"
	"pos-proxy/printing"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/abadojack/whatlanggo"
	"github.com/cloudinn/escpos"
	"github.com/cloudinn/escpos/connection"
)

func KitchenHeader(kitchen *printing.KitchenPrint, p *escpos.Printer) error {
	lang := printing.SetLang("")

	p.WriteString(printing.CheckLang("Printer ID: " + kitchen.Printer.PrinterID))

	p.WriteString(strings.Repeat("=", printing.PrintingParams(kitchen.Printer.PaperWidth, "char_per_line")))

	p.WriteString(printing.Center("Invoice Number"+": "+kitchen.Invoice.InvoiceNumber,
		printing.PrintingParams(kitchen.Printer.PaperWidth, "width")) +
		printing.CheckLang(printing.Translate("Invoice number", lang)+": "+kitchen.Invoice.InvoiceNumber))

	p.WriteString(printing.Center("Covers"+": "+fmt.Sprintf("%d", kitchen.Invoice.Pax),
		printing.PrintingParams(kitchen.Printer.PaperWidth, "width")) +
		printing.CheckLang(printing.Translate("Covers", lang)+": "+fmt.Sprintf("%d", kitchen.Invoice.Pax)))
	if kitchen.Invoice.TableID != nil {
		p.WriteString(printing.Center("Table"+": "+*kitchen.Invoice.TableDetails,
			printing.PrintingParams(kitchen.Printer.PaperWidth, "width")) +
			printing.CheckLang(printing.Translate("Table", lang)+": "+*kitchen.Invoice.TableDetails))
	} else {
		p.WriteString(printing.Center("Takeout",
			printing.PrintingParams(kitchen.Printer.PaperWidth, "width")) +
			printing.CheckLang(printing.Translate("Takeout", lang)))
	}
	guestName := ""
	if kitchen.Invoice.WalkinName != "" {
		guestName = kitchen.Invoice.WalkinName
	} else if kitchen.Invoice.ProfileDetails != "" {
		guestName = kitchen.Invoice.ProfileDetails
	} else if kitchen.Invoice.RoomDetails != nil {
		guestName = *kitchen.Invoice.RoomDetails
	}
	guestNameTrans := printing.Translate("Guest name", lang)
	info := whatlanggo.Detect(guestNameTrans)
	if guestName != "" {
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(printing.Center(guestNameTrans+": "+guestName,
				printing.PrintingParams(kitchen.Printer.PaperWidth, "width")) +
				printing.CheckLang(guestName) + ": " + printing.CheckLang(guestNameTrans))

		} else {
			p.WriteString(printing.Center(guestNameTrans+": "+guestName,
				printing.PrintingParams(kitchen.Printer.PaperWidth, "width")) +
				printing.CheckLang(guestNameTrans) + ": " + printing.CheckLang(guestName))
		}
	}
	p.WriteString(printing.Center(fmt.Sprintf("%d", kitchen.Cashier.Number)+"/"+
		kitchen.Cashier.Name, printing.PrintingParams(kitchen.Printer.PaperWidth, "width")) +
		printing.CheckLang(fmt.Sprintf("%d", kitchen.Cashier.Number)+"/"+kitchen.Cashier.Name))

	loc, _ := time.LoadLocation(kitchen.Timezone)
	submittedOn := time.Now().In(loc)
	date := submittedOn.Format(time.RFC1123)
	p.WriteString(printing.Center(date, printing.PrintingParams(kitchen.Printer.PaperWidth, "width")) +
		date)

	p.WriteString(strings.Repeat("=",
		printing.PrintingParams(kitchen.Printer.PaperWidth, "char_per_line")))
	return nil
}

func OrderTable(kitchen *printing.KitchenPrint, p *escpos.Printer) error {
	lang := printing.SetLang("")
	item := printing.CheckLang(printing.Translate("Item", lang))
	qty := printing.CheckLang(printing.Translate("Qty", lang))
	storeUnit := printing.CheckLang(printing.Translate("Unit", lang))
	p.SetWhiteOnBlack(false)
	if printing.PrintingParams(kitchen.Printer.PaperWidth, "width") == 760 {

		p.SetFontSizePoints(28)
	}
	info := whatlanggo.Detect(item)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(storeUnit + printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "qty_padding")-
			utf8.RuneCountInString(storeUnit)) + qty +
			printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "item_padding")-
				utf8.RuneCountInString(qty)) + item)

	} else {
		p.WriteString(item + printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "item_padding")-
			utf8.RuneCountInString(item)) + qty +
			printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "qty_padding")-
				utf8.RuneCountInString(qty)) + storeUnit)
	}
	p.SetWhiteOnBlack(true)
	p.SetFontSizePoints(30)
	p.WriteString(strings.Repeat("=",
		printing.PrintingParams(kitchen.Printer.PaperWidth, "char_per_line")))

	for _, item := range kitchen.GropLineItems {
		desc := printing.CheckLang(item.Description)
		qty := printing.CheckLang(fmt.Sprintf("%.2f", item.Quantity))
		baseUnit := printing.CheckLang(item.BaseUnit)
		if printing.PrintingParams(kitchen.Printer.PaperWidth, "width") == 760 {

			p.SetFontSizePoints(28)
		}
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(baseUnit +
				printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "qty_padding")-
					utf8.RuneCountInString(baseUnit)) + qty +
				printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "item_padding")-
					utf8.RuneCountInString(desc)) + desc)

		} else {
			p.WriteString(desc +
				printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "item_padding")-
					utf8.RuneCountInString(desc)) + qty +
				printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "qty_padding")-
					utf8.RuneCountInString(qty)) + baseUnit)
		}

		for _, condiment := range item.CondimentLineItems {
			p.WriteString(printing.CheckLang(condiment.Description))
		}
		if item.CondimentsComment != "" {
			if whatlanggo.Scripts[info.Script] == "Arabic" {
				p.WriteString(printing.Pad(37-utf8.RuneCountInString(item.CondimentsComment)) +
					printing.CheckLang(item.CondimentsComment))
			} else {
				p.WriteString(printing.CheckLang(item.CondimentsComment))
			}
		}
		p.SetFontSizePoints(30)
		if item.LastChildInCourse {
			p.WriteString(strings.Repeat("-", printing.PrintingParams(kitchen.Printer.PaperWidth, "char_per_line")))
		}
	}
	p.WriteString(strings.Repeat("=",
		printing.PrintingParams(kitchen.Printer.PaperWidth, "char_per_line")))

	return nil

}

func KitchenFooter(kitchen *printing.KitchenPrint, p *escpos.Printer) error {
	lang := printing.SetLang("")
	p.SetFontSizePoints(40)
	p.SetImageHight(42)
	p.WriteString(printing.Pad((30-len("This is not a"))/2) + strings.ToUpper(printing.CheckLang(printing.Translate("This is not a", lang))))
	p.WriteString(printing.Pad((30-len("valid tax invoice"))/2) +
		strings.ToUpper(printing.CheckLang(printing.Translate("valid tax invoice", lang))))
	p.Formfeed()

	return nil
}

//PrintKitchen to print kitchen recepit
func (e Esc) PrintKitchen(kitchen *printing.KitchenPrint) error {

	printing.SetLang("")
	var p *escpos.Printer
	var err error
	if kitchen.Printer.IsUSB {
		p, err = connection.NewConnection("usb", *kitchen.Printer.PrinterIP+":9100")
		if err != nil {
			return err
		}
	} else {
		p, err = connection.NewConnection("network", *kitchen.Printer.PrinterIP+":9100")
		if err != nil {
			return err
		}
	}
	KitchenHeader(kitchen, p)
	OrderTable(kitchen, p)
	KitchenFooter(kitchen, p)
	p.Cut()
	return nil

}
