package epsonxml

import (
	"encoding/xml"
	"fmt"
	"pos-proxy/printing"
	"strings"
	"time"
	"unicode/utf8"
)

func KitchenHeader(kitchen *printing.KitchenPrint) []printing.Text {

	header := []printing.Text{}
	doubleDashedLine := printing.AddLine("doubledashed", 55)
	tableTakeout := ""
	if kitchen.Invoice.TableID != nil {
		tableTakeout = printing.Translate("Table") + ": " + *kitchen.Invoice.TableDetails
	} else {
		tableTakeout = printing.Translate("Takeout")
	}
	guestName := ""
	if kitchen.Invoice.WalkinName != "" {
		guestName = kitchen.Invoice.WalkinName
	} else if kitchen.Invoice.ProfileDetails != "" {
		guestName = kitchen.Invoice.ProfileDetails
	} else if kitchen.Invoice.RoomDetails != nil {
		guestName = *kitchen.Invoice.RoomDetails
	}
	loc, _ := time.LoadLocation(kitchen.Timezone)
	submittedOn := time.Now().In(loc)
	date := submittedOn.Format(time.RFC1123)
	header = append(
		header,
		printing.Text{Text: printing.Translate("Printer ID") + ": " + kitchen.Printer.PrinterID + "\n"},
		printing.Text{Font: "font_b"},
		printing.Text{Text: doubleDashedLine + "\n"},
		printing.Text{Align: "center"},
		printing.Text{Text: printing.Translate("Invoice number") + ": " + kitchen.Invoice.InvoiceNumber + "\n"},
		printing.Text{Text: printing.Translate("Covers") + ": " + fmt.Sprintf("%d", kitchen.Invoice.Pax) + "\n"},
		printing.Text{Text: tableTakeout + "\n"},
		printing.Text{Text: printing.Translate("Guest name") + ": " + guestName + "\n"},
		printing.Text{Text: fmt.Sprintf("%d", kitchen.Cashier.Number) + "/" + kitchen.Cashier.Name + "\n"},
		printing.Text{Text: date + "\n"},
		printing.Text{Linespc: "30"},
		printing.Text{Text: doubleDashedLine + "\n"},
	)
	return header
}

func OrderTableHeader(kitchen *printing.KitchenPrint) []printing.Text {
	tableHeader := []printing.Text{}
	tableHeader = append(
		tableHeader,
		printing.Text{Align: "left"},
		printing.Text{Reverse: "true"},
		printing.Text{UnderLine: "false"},
		printing.Text{Emphasized: "true"},
		printing.Text{Color: "color_1"},
		printing.Text{Font: "font_a"},
		printing.Text{Text: printing.Translate("Item") + printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "item_kitchen")-
			utf8.RuneCountInString("Item")) +
			printing.Translate("Qty") + printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "qty_kitchen")-utf8.RuneCountInString("Qty")) +
			printing.Translate("Unit") + printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "unit")-utf8.RuneCountInString("Unit")) + "\n"},
	)
	return tableHeader
}
func OrderTableContent(kitchen *printing.KitchenPrint) []printing.Text {
	tableContent := []printing.Text{}
	doubleDashedLine := printing.AddLine("doubledashed",
		printing.PrintingParams(kitchen.Printer.PaperWidth, "char_per_line"))
	dashedLine := printing.AddLine("dash",
		printing.PrintingParams(kitchen.Printer.PaperWidth, "char_per_line"))
	tableContent = append(
		tableContent,
		printing.Text{Reverse: "false"},
		printing.Text{UnderLine: "false"},
		printing.Text{Emphasized: "false"},
		printing.Text{Color: "color_1"},
		printing.Text{Text: doubleDashedLine + "\n"},
	)
	kitchecnItems := []printing.Text{}
	for _, item := range kitchen.GropLineItems {
		row := ""
		desc := item.Description
		qty := fmt.Sprintf("%.2f", item.Quantity)
		baseUnit := item.BaseUnit
		row = printing.CheckLang(desc) +
			printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "item_kitchen")-
				utf8.RuneCountInString(desc)) + qty +
			printing.Pad(printing.PrintingParams(kitchen.Printer.PaperWidth, "qty_kitchen")-
				utf8.RuneCountInString(qty)) + baseUnit + "\n"
		for _, condiment := range item.CondimentLineItems {
			row += condiment.Description + "\n"
		}
		if item.CondimentsComment != "" {
			row += item.CondimentsComment + "\n"
		}
		if item.LastChildInCourse {
			row += dashedLine + "\n"
		}
		kitchecnItems = append(kitchecnItems, printing.Text{Text: row})
	}
	tableContent = append(tableContent, kitchecnItems...)

	return tableContent
}
func KitchenFooter(kitchen *printing.KitchenPrint) []printing.Text {
	fotter := []printing.Text{}
	doubleDashedLine := printing.AddLine("doubledashed",
		printing.PrintingParams(kitchen.Printer.PaperWidth, "char_per_line"))
	fotter = append(
		fotter,
		printing.Text{Text: doubleDashedLine + "\n"},
		printing.Text{Font: "font_b"},
		printing.Text{DoubleWidth: "true"},
		printing.Text{DoubleHight: "true"},
		printing.Text{Align: "center"},
		printing.Text{Text: strings.ToUpper(printing.Translate("This is not a")) + "\n"},
		printing.Text{Text: strings.ToUpper(printing.Translate("valid tax invoice"))},
	)
	return fotter

}
func (e Epsonxml) PrintKitchen(kitchen *printing.KitchenPrint) error {
	xmlReq := printing.New()
	xmlReq.XMLns = "http://www.epson-pos.com/schemas/2011/03/epos-print"
	kitchenHeader := KitchenHeader(kitchen)
	tableHeader := OrderTableHeader(kitchen)
	tableContent := OrderTableContent(kitchen)
	kitchenFooter := KitchenFooter(kitchen)
	xmlReq.Text = append(xmlReq.Text, kitchenHeader...)
	xmlReq.Text = append(xmlReq.Text, tableHeader...)
	xmlReq.Text = append(xmlReq.Text, tableContent...)
	xmlReq.Text = append(xmlReq.Text, kitchenFooter...)
	xmlReq.Cut.Type = "feed"
	reqBody, err := xml.Marshal(xmlReq)
	if err != nil {
		return err
	}
	api := "http://" + *kitchen.Printer.PrinterIP + "/cgi-bin/epos/service.cgi?devid=local_printer"
	printing.Send(api, reqBody)
	// log.Println(string(reqBody))
	return nil
}
