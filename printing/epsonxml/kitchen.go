package epsonxml

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cloudinn/escpos"

	"pos-proxy/printing"
)

var p *escpos.Printer

func KitchenHeader(kitchen *printing.KitchenPrint) []printing.Image {
	lang := printing.SetLang("")
	header := []printing.Image{}

	//Printer ID
	data, w, h, _ := p.TextToRaster(printing.CheckLang(printing.Translate("Printer ID", lang) + ": " + kitchen.Printer.PrinterID))
	imgXML := ImgToXML(data, w, h)
	header = append(header, *imgXML)

	//double Dashed Line
	doubleDashedLine := printing.AddLine("doubledashed", 55)

	data, w, h, _ = p.TextToRaster(doubleDashedLine)
	imgXML = ImgToXML(data, w, h)
	header = append(header, *imgXML)

	// Invoice Number
	data, w, h, _ = p.TextToRaster(printing.Pad((((kitchen.Printer.PaperWidth / 2) -
		len(printing.CheckLang(printing.Translate("Invoice number", lang)+": "+kitchen.Invoice.InvoiceNumber))) / 2)) +
		printing.CheckLang(printing.Translate("Invoice number", lang)+": "+kitchen.Invoice.InvoiceNumber))
	imgXML = ImgToXML(data, w, h)
	header = append(header, *imgXML)

	// Covers
	data, w, h, _ = p.TextToRaster(printing.Pad((((kitchen.Printer.PaperWidth / 2) -
		len(printing.Translate("Covers", lang)+": "+fmt.Sprintf("%d", kitchen.Invoice.Pax))) / 2)) +
		printing.Translate("Covers", lang) + ": " + fmt.Sprintf("%d", kitchen.Invoice.Pax))
	imgXML = ImgToXML(data, w, h)
	header = append(header, *imgXML)

	// Table
	tableTakeout := ""
	if kitchen.Invoice.TableID != nil {
		tableTakeout = printing.Translate("Table", lang) + ": " + *kitchen.Invoice.TableDetails
	} else {
		tableTakeout = printing.Translate("Takeout", lang)
	}

	data, w, h, _ = p.TextToRaster(printing.Pad((((kitchen.Printer.PaperWidth / 2) -
		len(tableTakeout)) / 2)) +
		tableTakeout)
	imgXML = ImgToXML(data, w, h)
	header = append(header, *imgXML)

	// Guest Name
	guestName := ""
	if kitchen.Invoice.WalkinName != "" {
		guestName = kitchen.Invoice.WalkinName
	} else if kitchen.Invoice.ProfileDetails != "" {
		guestName = kitchen.Invoice.ProfileDetails
	} else if kitchen.Invoice.RoomDetails != nil {
		guestName = *kitchen.Invoice.RoomDetails
	}

	guest := printing.Translate("Guest name", lang) + ": " + printing.CheckLang(guestName)
	data, w, h, _ = p.TextToRaster(printing.Pad((((kitchen.Printer.PaperWidth / 2) -
		(len(guest) / 2)) / 2)) +
		guest)
	imgXML = ImgToXML(data, w, h)
	header = append(header, *imgXML)

	// Cashier
	cashier := fmt.Sprintf("%d", kitchen.Cashier.Number) + "/" + printing.CheckLang(kitchen.Cashier.Name)
	data, w, h, _ = p.TextToRaster(printing.Pad((((kitchen.Printer.PaperWidth / 2) -
		(len(cashier) / 2)) / 2)) + cashier)
	imgXML = ImgToXML(data, w, h)
	header = append(header, *imgXML)

	// date
	loc, _ := time.LoadLocation(kitchen.Timezone)
	submittedOn := time.Now().In(loc)
	date := submittedOn.Format(time.RFC1123)

	data, w, h, _ = p.TextToRaster(printing.Pad((((kitchen.Printer.PaperWidth / 2) -
		len(date)) / 2)) + date)
	imgXML = ImgToXML(data, w, h)
	header = append(header, *imgXML)

	//double Dashed Line
	data, w, h, _ = p.TextToRaster(doubleDashedLine)
	imgXML = ImgToXML(data, w, h)
	header = append(header, *imgXML)

	return header
}

func OrderTableHeader(kitchen *printing.KitchenPrint) []printing.Image {
	lang := printing.SetLang("")
	tableHeader := []printing.Image{}

	//Header
	data, w, h, _ := p.TextToRaster(
		printing.Translate("Item", lang) + printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "item_kitchen")-
			utf8.RuneCountInString("Item")) +
			printing.Translate("Qty", lang) + printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "qty_kitchen")-utf8.RuneCountInString("Qty")) +
			printing.Translate("Unit", lang) + printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "unit")-utf8.RuneCountInString("Unit")))
	imgXML := ImgToXML(data, w, h)
	tableHeader = append(tableHeader, *imgXML)

	return tableHeader
}

func OrderTableContent(kitchen *printing.KitchenPrint) []printing.Image {
	tableContent := []printing.Image{}

	//double Dashed Line
	doubleDashedLine := printing.AddLine("doubledashed",
		printing.PrintingParams(kitchen.Printer.PaperWidth, "char_per_line"))

	data, w, h, _ := p.TextToRaster(doubleDashedLine)
	imgXML := ImgToXML(data, w, h)
	tableContent = append(tableContent, *imgXML)

	// dash line
	dashedLine := printing.AddLine("dash",
		printing.PrintingParams(kitchen.Printer.PaperWidth, "char_per_line"))

	// items
	kitchecnItems := []printing.Image{}
	for _, item := range kitchen.GropLineItems {
		// item
		row := ""
		desc := item.Description
		qty := fmt.Sprintf("%.2f", item.Quantity)
		baseUnit := item.BaseUnit
		row = printing.CheckLang(desc) +
			printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "item_kitchen")-
				utf8.RuneCountInString(desc)) + qty +
			printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "qty_kitchen")-
				utf8.RuneCountInString(qty)) + printing.CheckLang(baseUnit)
		for _, condiment := range item.CondimentLineItems {
			row += printing.CheckLang(condiment.Description)
		}
		data, w, h, _ := p.TextToRaster(row)
		imgXML := ImgToXML(data, w, h)
		kitchecnItems = append(kitchecnItems, *imgXML)

		if item.CondimentsComment != "" {
			row = printing.CheckLang(item.CondimentsComment)
			data, w, h, _ := p.TextToRaster(row)
			imgXML := ImgToXML(data, w, h)
			kitchecnItems = append(kitchecnItems, *imgXML)
		}

		// dash line
		if item.LastChildInCourse {
			row = dashedLine
		}
		data, w, h, _ = p.TextToRaster(row)
		imgXML = ImgToXML(data, w, h)
		kitchecnItems = append(kitchecnItems, *imgXML)
	}
	tableContent = append(tableContent, kitchecnItems...)

	return tableContent
}

func KitchenFooter(kitchen *printing.KitchenPrint) []printing.Image {
	lang := printing.SetLang("")
	footer := []printing.Image{}

	// double Dashed Line
	doubleDashedLine := printing.AddLine("doubledashed", printing.PrintingParams(kitchen.Printer.PaperWidth, "char_per_line"))

	data, w, h, _ := p.TextToRaster(doubleDashedLine)
	imgXML := ImgToXML(data, w, h)
	footer = append(footer, *imgXML)

	// note
	data, w, h, _ = p.TextToRaster(printing.Pad((((kitchen.Printer.PaperWidth / 2) - len("This is not a")) / 2)) +
		printing.CheckLang(strings.ToUpper(printing.Translate("This is not a", lang))))
	imgXML = ImgToXML(data, w, h)
	footer = append(footer, *imgXML)

	data, w, h, _ = p.TextToRaster(printing.Pad((((kitchen.Printer.PaperWidth / 2) - len("valid tax invoice")) / 2)) +
		printing.CheckLang(strings.ToUpper(printing.Translate("valid tax invoice", lang))))
	imgXML = ImgToXML(data, w, h)
	footer = append(footer, *imgXML)

	return footer
}

func (e Epsonxml) PrintKitchen(kitchen *printing.KitchenPrint) error {
	xmlReq := printing.New()
	xmlReq.XMLns = "http://schemas.xmlsoap.org/soap/envelope/"
	eposPrint := printing.EposPrint{}
	eposPrint.XMLns = "http://www.epson-pos.com/schemas/2011/03/epos-print"
	eposPrint.Layout = &printing.Layout{}
	eposPrint.Layout.Type = "receipt"
	eposPrint.Layout.Width = "800"

	// header
	kitchenHeader := KitchenHeader(kitchen)
	eposPrint.Image = append(eposPrint.Image, kitchenHeader...)

	// table header
	tableHeader := OrderTableHeader(kitchen)
	eposPrint.Image = append(eposPrint.Image, tableHeader...)

	// table content
	tableContent := OrderTableContent(kitchen)
	eposPrint.Image = append(eposPrint.Image, tableContent...)

	// footer
	kitchenFooter := KitchenFooter(kitchen)
	eposPrint.Image = append(eposPrint.Image, kitchenFooter...)

	eposPrint.Text = append(eposPrint.Text, printing.Text{Text: "\n\n"})
	eposPrint.Cut.Type = "feed"
	xmlReq.Body.EposPrint = eposPrint
	reqBody, err := xml.Marshal(xmlReq)
	if err != nil {
		return err
	}
	// append xml header
	reqBody = []byte(xml.Header + string(reqBody))
	api := "http://" + *kitchen.Printer.PrinterIP + "/cgi-bin/epos/service.cgi?devid=" +
		kitchen.Printer.PrinterID + "&timeout=6000"
	printing.Send(api, reqBody)
	_ = reqBody
	return nil
}
