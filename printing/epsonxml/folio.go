package epsonxml

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"image"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cloudinn/escpos/raster"

	"pos-proxy/config"
	"pos-proxy/printing"
)

func FolioHeader(folio *printing.FolioPrint) []printing.Image {
	lang := printing.SetLang(folio.Terminal.RCRS)
	folioHeader := []printing.Image{}

	// header txt
	headerText := ""
	if folio.Invoice.IsSettled {
		headerText = strings.ToUpper(printing.Translate("Tax Invoice", lang))
		if folio.Invoice.PaidAmount < 0 {
			headerText += strings.ToUpper(printing.Translate("Return", lang))
		}
	} else {
		headerText = strings.ToUpper(printing.Translate("proforma", lang))
		if config.Config.IsFDMEnabled {
			headerText += strings.ToUpper(printing.Translate("This is not a", lang))
			headerText += strings.ToUpper(printing.Translate("vaild tax invoice", lang))
		}
	}
	data, w, h, _ := p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-(len(headerText)+10))/2))+
		headerText, 40.0, true)
	imgXML := ImgToXML(data, w, h)
	folioHeader = append(folioHeader, *imgXML)

	// newline
	data, w, h, _ = p.TextToRaster(printing.Pad(len("newline")), 30.0, true)
	imgXML = ImgToXML(data, w, h)
	folioHeader = append(folioHeader, *imgXML)

	// company name
	data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-(len(folio.Company.Name)+8))/2))+
		folio.Company.Name, 38.0, true)
	imgXML = ImgToXML(data, w, h)
	folioHeader = append(folioHeader, *imgXML)

	// vat number
	data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-(len(folio.Company.VATNumber)))/2))+
		folio.Company.VATNumber, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	folioHeader = append(folioHeader, *imgXML)

	// address company
	data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-(len(folio.Company.Address)))/2))+
		folio.Company.Address, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	folioHeader = append(folioHeader, *imgXML)

	// company postcode
	data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-(len(folio.Company.PostalCode+"-"+folio.Company.City)))/2))+
		folio.Company.PostalCode+"-"+folio.Company.City, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	folioHeader = append(folioHeader, *imgXML)

	// Invoice header
	headersSlice := []printing.Image{}
	var headers []string

	if folio.Store.InvoiceHeader != "" {
		headers = strings.Split(folio.Store.InvoiceHeader, "\n")
		for _, header := range headers {
			if utf8.RuneCountInString(header) > printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line") {
				data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-(len(header[0:printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line")])))/2))+
					header[0:printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line")], 30.0, true)
				imgXML = ImgToXML(data, w, h)
				headersSlice = append(headersSlice, *imgXML)

				data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-(len(header[printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line"):])))/2))+
					header[printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line"):], 30.0, true)
				imgXML = ImgToXML(data, w, h)
				headersSlice = append(headersSlice, *imgXML)

			} else {
				data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-(len(header)))/2))+
					header, 30.0, true)
				imgXML = ImgToXML(data, w, h)
				headersSlice = append(headersSlice, *imgXML)
			}

		}
	}
	folioHeader = append(folioHeader, headersSlice...)

	// Description
	data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-(len(folio.Store.Description)))/2))+
		folio.Store.Description, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	folioHeader = append(folioHeader, *imgXML)

	// Invoice number
	data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-(len(printing.Translate("Invoice number", lang)+": "+folio.Invoice.InvoiceNumber)))/2))+
		printing.Translate("Invoice number", lang)+": "+folio.Invoice.InvoiceNumber, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	folioHeader = append(folioHeader, *imgXML)

	// Cover
	data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-(len(printing.Translate("Covers", lang)+": "+fmt.Sprintf("%d", folio.Invoice.Pax))+10))/2))+
		printing.Translate("Covers", lang)+": "+fmt.Sprintf("%d", folio.Invoice.Pax), 40.0, true)
	imgXML = ImgToXML(data, w, h)
	folioHeader = append(folioHeader, *imgXML)

	// Table Takout
	tableTakeout := ""
	if folio.Invoice.TableID != nil {
		tableTakeout = printing.Translate("Table", lang) + ": " + *folio.Invoice.TableDetails
	} else {
		tableTakeout = printing.Translate("Takeout", lang)
	}

	data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-(len(tableTakeout)+10))/2))+
		tableTakeout, 40.0, true)
	imgXML = ImgToXML(data, w, h)
	folioHeader = append(folioHeader, *imgXML)

	// Guest
	guestName := ""
	if folio.Invoice.WalkinName != "" {
		guestName = folio.Invoice.WalkinName
	} else if folio.Invoice.ProfileDetails != "" {
		guestName = folio.Invoice.ProfileDetails
	} else if folio.Invoice.RoomDetails != nil {
		guestName = *folio.Invoice.RoomDetails
	} else if folio.Invoice.PaymasterDetails != "" {
		guestName = folio.Invoice.PaymasterDetails
	}

	guest := printing.Translate("Guest name", lang) + ": " + printing.CheckLang(guestName)
	data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-
		(len(guest)/2))/2))+
		guest, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	folioHeader = append(folioHeader, *imgXML)

	return folioHeader
}

func FolioTableHeader(folio *printing.FolioPrint) []printing.Image {
	lang := printing.SetLang(folio.Terminal.RCRS)
	tableHeader := []printing.Image{}

	// newline
	data, w, h, _ := p.TextToRaster(printing.Pad(len("newline")), 30.0, true)
	imgXML := ImgToXML(data, w, h)
	tableHeader = append(tableHeader, *imgXML)

	// Header
	item := printing.Translate("Item", lang)
	qty := printing.Translate("Qty", lang)
	price := printing.Translate("Price", lang)

	tableHeaderText := item + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "item_padding")-utf8.RuneCountInString(item)) +
		qty + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "qty_padding")-utf8.RuneCountInString(qty)) +
		price + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "price_padding")-utf8.RuneCountInString(price))

	if config.Config.IsFDMEnabled {
		tax := printing.Translate("Tax", lang)
		tableHeaderText += tax + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "tax_padding")-utf8.RuneCountInString(tax))
	}

	data, w, h, _ = p.TextToRaster(tableHeaderText, 30.0, false)
	imgXML = ImgToXML(data, w, h)
	tableHeader = append(tableHeader, *imgXML)

	return tableHeader
}

func FolioTableContent(folio *printing.FolioPrint) []printing.Image {
	lang := printing.SetLang(folio.Terminal.RCRS)
	tableContent := []printing.Image{}

	vatsToDisplay := map[string]bool{
		"A": false,
		"B": false,
		"C": false,
		"D": false,
	}

	// items
	for _, item := range folio.Invoice.Items {
		price := fmt.Sprintf("%.2f", item.Price)
		desc := printing.CheckLang(item.Description)
		qty := fmt.Sprintf("%.2f", item.Quantity)
		row := ""
		row = desc + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "item_padding")-
			utf8.RuneCountInString(desc)) + qty +
			printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "qty_padding")-
				utf8.RuneCountInString(qty)) + price +
			printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "price_padding")-
				utf8.RuneCountInString(price))
		if config.Config.IsFDMEnabled {
			row += item.VAT + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "tax_padding")-1)
			vatsToDisplay[item.VAT] = true
		}

		data, w, h, _ := p.TextToRaster(row, 30.0, true)
		imgXML := ImgToXML(data, w, h)

		tableContent = append(tableContent, *imgXML)
	}

	// newline
	data, w, h, _ := p.TextToRaster(printing.Pad(len("newline")), 30.0, true)
	imgXML := ImgToXML(data, w, h)
	tableContent = append(tableContent, *imgXML)

	// subtotal
	subTotal := fmt.Sprintf("%.2f", folio.Invoice.Subtotal)
	subTotalVal := ""
	if folio.Invoice.HouseUse {
		subTotalVal = "0.00"
	} else {
		subTotalVal = subTotal
	}
	subTotalTrans := printing.Translate("Subtotal", lang)
	subTotalText := subTotalTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
		utf8.RuneCountInString(subTotalVal)-utf8.RuneCountInString(subTotalTrans)) +
		subTotalVal

	data, w, h, _ = p.TextToRaster(subTotalText, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	tableContent = append(tableContent, *imgXML)

	// total discount
	totalDiscountText := ""
	if folio.TotalDiscounts > 0.0 {
		totalDiscount := fmt.Sprintf("%.2f", folio.TotalDiscounts)
		totalDiscountTrans := printing.Translate("Total discounts", lang)
		totalDiscountText = totalDiscountTrans +
			printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
				utf8.RuneCountInString(totalDiscount)-utf8.RuneCountInString(totalDiscountTrans)) +
			totalDiscount
	}
	data, w, h, _ = p.TextToRaster(totalDiscountText, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	tableContent = append(tableContent, *imgXML)

	// total
	total := fmt.Sprintf("%.2f", folio.Invoice.Total)
	totalVal := "0.00"
	if folio.Invoice.HouseUse {
		totalVal = "0.00"
	} else {
		totalVal = total
	}
	totalTrans := printing.Translate("Total", lang)
	totalText := totalTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "total_padding")-
		utf8.RuneCountInString(totalVal)-utf8.RuneCountInString(totalTrans)) + totalVal
	data, w, h, _ = p.TextToRaster(totalText, 40.0, true)
	imgXML = ImgToXML(data, w, h)
	tableContent = append(tableContent, *imgXML)

	// payment
	paymentText := ""
	if folio.Invoice.IsSettled == true && folio.Invoice.Postings != nil &&
		len(folio.Invoice.Postings) > 0 {
		paymentTrans := printing.Translate("Payment", lang)
		paymentText = paymentTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
			utf8.RuneCountInString(paymentTrans)-utf8.RuneCountInString(total)) + total
	}
	data, w, h, _ = p.TextToRaster(paymentText, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	tableContent = append(tableContent, *imgXML)

	//guest
	for _, posting := range folio.Invoice.Postings {
		postingText := ""
		guestName := ""
		deptAmount := fmt.Sprintf("%.2f", posting.Amount)
		if posting.RoomNumber != nil && *posting.RoomNumber != 0 {

			if folio.Invoice.WalkinName != "" {
				guestName = printing.CheckLang(folio.Invoice.WalkinName)
			} else if folio.Invoice.ProfileDetails != "" {
				guestName = printing.CheckLang(folio.Invoice.ProfileDetails)
			} else if folio.Invoice.RoomDetails != nil || *folio.Invoice.RoomDetails != "" {
				guestName = printing.CheckLang(*folio.Invoice.RoomDetails)
			}
			postingText = fmt.Sprintf("%d", posting.RoomNumber) +
				" " + guestName + printing.Pad(32-utf8.RuneCountInString(guestName)-
				utf8.RuneCountInString(fmt.Sprintf("%d", posting.RoomNumber))-
				utf8.RuneCountInString(deptAmount)) + deptAmount

		} else {

			postingText = posting.DepartmentDetails +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
					utf8.RuneCountInString(posting.DepartmentDetails)-
					utf8.RuneCountInString(deptAmount)) + deptAmount

		}
		data, w, h, _ = p.TextToRaster(postingText, 30.0, true)
		imgXML = ImgToXML(data, w, h)
		tableContent = append(tableContent, *imgXML)

		if len(posting.GatewayResponses) > 0 {
			for _, response := range posting.GatewayResponses {
				data, w, h, _ = p.TextToRaster(response, 30.0, true)
				imgXML = ImgToXML(data, w, h)
				tableContent = append(tableContent, *imgXML)
			}
		}
	}

	// received
	received := "0.0"
	if folio.Invoice.Change != 0.0 {
		received = fmt.Sprintf("%.2f", folio.Invoice.Change+folio.Invoice.Total)
	} else {
		received = fmt.Sprintf("%.2f", folio.Invoice.Total)
	}
	receivedTrans := printing.Translate("Received", lang)
	receivedText := receivedTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
		utf8.RuneCountInString(receivedTrans)-utf8.RuneCountInString(received)) +
		received
	data, w, h, _ = p.TextToRaster(receivedText, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	tableContent = append(tableContent, *imgXML)

	// cahnge
	change := "0.00"
	changeTrans := printing.Translate("Change", lang)

	if folio.Invoice.Change != 0.0 {
		change = fmt.Sprintf("%.2f", folio.Invoice.Change)
	}
	changeText := changeTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
		utf8.RuneCountInString(change)-utf8.RuneCountInString(changeTrans)) + change
	data, w, h, _ = p.TextToRaster(changeText, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	tableContent = append(tableContent, *imgXML)

	return tableContent
}

func FDMSection(folio *printing.FolioPrint) []printing.Image {
	lang := printing.SetLang(folio.Terminal.RCRS)
	fdm := []printing.Image{}

	vatsToDisplay := map[string]bool{
		"A": false,
		"B": false,
		"C": false,
		"D": false,
	}

	if config.Config.IsFDMEnabled {
		taxableTrans := printing.Translate("Taxable", lang)
		rateTrans := printing.Translate("Rate", lang)
		vatTrans := printing.Translate("Vat", lang)
		netTrans := printing.Translate("Net", lang)

		for _, res := range folio.Invoice.FDMResponses {
			// table header
			fdmTableHeader := rateTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_rate_padding")-
				utf8.RuneCountInString(rateTrans)) +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_taxable_padding")-
					utf8.RuneCountInString(taxableTrans)) + taxableTrans +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_vat_padding")-
					utf8.RuneCountInString(vatTrans)) + vatTrans +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_net_padding")-
					utf8.RuneCountInString(netTrans)) + netTrans

			data, w, h, _ := p.TextToRaster(fdmTableHeader, 30.0, false)
			imgXML := ImgToXML(data, w, h)
			fdm = append(fdm, *imgXML)

			for k, v := range vatsToDisplay {
				if v {
					taxableAmount := fmt.Sprintf("%.2f", res.VATSummary[k]["taxable_amount"])
					vatAmount := fmt.Sprintf("%.2f", res.VATSummary[k]["vat_amount"])
					netAmount := fmt.Sprintf("%.2f", res.VATSummary[k]["net_amount"])
					fdmTableContent := k +
						printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_rate_padding")) +
						printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_taxable_padding")-
							utf8.RuneCountInString(taxableAmount)) + taxableAmount +
						printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_vat_padding")-
							utf8.RuneCountInString(vatAmount)) + vatAmount +
						printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_net_padding")-
							utf8.RuneCountInString(netAmount)) + netAmount

					data, w, h, _ := p.TextToRaster(fdmTableContent, 30.0, true)
					imgXML := ImgToXML(data, w, h)
					fdm = append(fdm, *imgXML)
				}
			}

			totalTaxableAmount := fmt.Sprintf("%.2f", res.VATSummary["Total"]["taxable_amount"])
			totalVatAmount := fmt.Sprintf("%.2f", res.VATSummary["Total"]["vat_amount"])
			totalNetAmount := fmt.Sprintf("%.2f", res.VATSummary["Total"]["net_amount"])
			totalTrans := printing.Translate("Total", lang)
			totalText := totalTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_rate_padding")-
				utf8.RuneCountInString(totalTrans)) +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_taxable_padding")-
					utf8.RuneCountInString(totalTaxableAmount)) + totalTaxableAmount +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_vat_padding")-
					utf8.RuneCountInString(totalVatAmount)) + totalVatAmount +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_net_padding")-
					utf8.RuneCountInString(totalNetAmount)) + totalNetAmount

			data, w, h, _ = p.TextToRaster(totalText, 30.0, true)
			imgXML = ImgToXML(data, w, h)
			fdm = append(fdm, *imgXML)
		}
	}

	// newline
	data, w, h, _ := p.TextToRaster(printing.Pad(len("newline")), 30.0, true)
	imgXML := ImgToXML(data, w, h)
	fdm = append(fdm, *imgXML)

	return fdm
}

func FolioFooter(folio *printing.FolioPrint) []printing.Image {
	lang := printing.SetLang(folio.Terminal.RCRS)
	footer := []printing.Image{}

	// open at
	openedAt := printing.Translate("Opened at", lang) + ": " + folio.Invoice.CreatedOn
	data, w, h, _ := p.TextToRaster(openedAt, 30.0, true)
	imgXML := ImgToXML(data, w, h)
	footer = append(footer, *imgXML)

	// close on
	loc, _ := time.LoadLocation(folio.Timezone)
	closedOnText := ""
	if folio.Invoice.ClosedOn != nil {
		closedAt := folio.Invoice.ClosedOn.In(loc)
		closedAtStr := closedAt.Format("02 Jan 2006 15:04:05")
		cloasedOn := printing.Translate("Closed on", lang)
		closedOnText = cloasedOn + ": " + closedAtStr
	}
	data, w, h, _ = p.TextToRaster(closedOnText, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	footer = append(footer, *imgXML)

	//
	createdBy := printing.Translate("Created by", lang) + ": " + folio.Invoice.CashierDetails
	data, w, h, _ = p.TextToRaster(createdBy, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	footer = append(footer, *imgXML)

	//
	printedBy := printing.Translate("Printed by", lang) + ": " + strconv.Itoa(folio.Cashier.Number)
	data, w, h, _ = p.TextToRaster(printedBy, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	footer = append(footer, *imgXML)

	// fdm
	if config.Config.IsFDMEnabled {
		for _, res := range folio.Invoice.FDMResponses {
			// ticket number
			ticketNumber := printing.Translate("Ticket Number", lang) + ": " + res.TicketNumber
			data, w, h, _ = p.TextToRaster(ticketNumber, 30.0, true)
			imgXML = ImgToXML(data, w, h)
			footer = append(footer, *imgXML)

			// ticket date
			ticketDate := printing.Translate("Ticket Date", lang) + ": " + res.Date.Format("02 Jan 2006 15:04:05")
			data, w, h, _ = p.TextToRaster(ticketDate, 30.0, true)
			imgXML = ImgToXML(data, w, h)
			footer = append(footer, *imgXML)

			// event
			event := printing.Translate("Event", lang) + ": " + res.EventLabel
			if utf8.RuneCountInString(event) > 32 {
				data, w, h, _ = p.TextToRaster(event[0:32], 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)

				data, w, h, _ = p.TextToRaster(event[32:], 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)

			} else {
				data, w, h, _ = p.TextToRaster(event, 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)
			}

			// terminal identifier
			terminalIdentifier := printing.Translate("Terminal Identifier", lang) + ": " +
				folio.Terminal.RCRS + "/" + folio.Terminal.Description
			if utf8.RuneCountInString(terminalIdentifier) >
				printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line") {
				data, w, h, _ = p.TextToRaster(terminalIdentifier[0:32], 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)

				data, w, h, _ = p.TextToRaster(terminalIdentifier[32:], 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)
			} else {
				data, w, h, _ = p.TextToRaster(terminalIdentifier, 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)
			}

			// production number
			productionNumber := printing.Translate("Production Number", lang) + ": " + folio.Terminal.RCRS
			data, w, h, _ = p.TextToRaster(productionNumber, 30.0, true)
			imgXML = ImgToXML(data, w, h)
			footer = append(footer, *imgXML)

			// software version
			SWVersion := printing.Translate("Software Version", lang) + ": " + res.SoftwareVersion
			data, w, h, _ = p.TextToRaster(SWVersion, 30.0, true)
			imgXML = ImgToXML(data, w, h)
			footer = append(footer, *imgXML)

			// ticket
			ticket := printing.Translate("Ticket", lang) + ": " +
				strings.Join(strings.Fields(res.TicketCounter), " ") +
				"/" + strings.Join(strings.Fields(res.TotalTicketCounter), " ") +
				" " + res.EventLabel
			if utf8.RuneCountInString(ticket) > 32 {
				data, w, h, _ = p.TextToRaster(ticket[0:32], 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)

				data, w, h, _ = p.TextToRaster(ticket[32:], 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)
			} else {
				data, w, h, _ = p.TextToRaster(ticket, 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)
			}

			// hashs
			if utf8.RuneCountInString(res.PLUHash) > 32 {
				data, w, h, _ = p.TextToRaster(printing.Translate("Hash", lang)+"s: "+res.PLUHash[0:25], 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)

				data, w, h, _ = p.TextToRaster(printing.Pad(7)+res.PLUHash[25:], 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)
			} else {
				data, w, h, _ = p.TextToRaster(printing.Translate("Hash", lang)+": "+res.PLUHash, 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)
			}

			// ticket sig
			if folio.Invoice.IsSettled {
				data, w, h, _ = p.TextToRaster(printing.Translate("Ticket Sig", lang)+": "+res.Signature[0:20], 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)

				data, w, h, _ = p.TextToRaster(printing.Pad(13)+res.Signature[20:], 30.0, true)
				imgXML = ImgToXML(data, w, h)
				footer = append(footer, *imgXML)
			}

			// control data
			controlData := printing.Translate("Control Data", lang) + ": " + res.Date.String()[0:10] +
				" " + res.TimePeriod.Format("15:04:00")
			data, w, h, _ = p.TextToRaster(controlData, 30.0, true)
			imgXML = ImgToXML(data, w, h)
			footer = append(footer, *imgXML)

			// control module id
			controlModule := printing.Translate("Control Module ID", lang) + ": " + res.ProductionNumber
			data, w, h, _ = p.TextToRaster(controlModule, 30.0, true)
			imgXML = ImgToXML(data, w, h)
			footer = append(footer, *imgXML)

			// VSC ID
			vsc := printing.Translate("VSC ID", lang) + ": " + res.VSC
			data, w, h, _ = p.TextToRaster(vsc, 30.0, true)
			imgXML = ImgToXML(data, w, h)
			footer = append(footer, *imgXML)
		}

	}

	// signature
	signature := printing.Translate("Signature", lang)
	sigText := signature + ":    " + "............."
	data, w, h, _ = p.TextToRaster(sigText, 30.0, true)
	imgXML = ImgToXML(data, w, h)
	footer = append(footer, *imgXML)

	// newline
	data, w, h, _ = p.TextToRaster(printing.Pad(len("newline")), 30.0, true)
	imgXML = ImgToXML(data, w, h)
	footer = append(footer, *imgXML)

	// Invoice footer
	if folio.Store.InvoiceFooter != "" {
		if utf8.RuneCountInString(folio.Store.InvoiceFooter) >
			printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line") {
			data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-
				(len(folio.Store.InvoiceFooter[0:printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line")])))/2))+
				folio.Store.InvoiceFooter[0:printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line")],
				30.0, true)
			imgXML = ImgToXML(data, w, h)
			footer = append(footer, *imgXML)

			data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-
				(len(folio.Store.InvoiceFooter[printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line"):])))/2))+
				folio.Store.InvoiceFooter[printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line"):],
				30.0, true)
			imgXML = ImgToXML(data, w, h)
			footer = append(footer, *imgXML)

		} else {
			data, w, h, _ = p.TextToRaster(printing.Pad((((folio.Printer.PaperWidth/2)-
				(len(folio.Store.InvoiceFooter)))/2))+
				folio.Store.InvoiceFooter, 30.0, true)
			imgXML = ImgToXML(data, w, h)
			footer = append(footer, *imgXML)
		}
	}

	return footer
}

//PrintFolio Prints xml format of folio receipt
func (e Epsonxml) PrintFolio(folio *printing.FolioPrint) error {
	xmlReq := printing.New()
	xmlReq.XMLns = "http://schemas.xmlsoap.org/soap/envelope/"
	eposPrint := printing.EposPrint{}
	eposPrint.XMLns = "http://www.epson-pos.com/schemas/2011/03/epos-print"
	eposPrint.Align = &printing.Text{Align: "center"}

	// image logo
	if folio.Store.Logo != "" {
		imagePath, err := printing.GetImage(folio.Store.Logo)
		if err != nil {
			return err
		}

		imgFile, err := os.Open(imagePath)
		if err != nil {
			return err
		}
		defer imgFile.Close()

		img, _, err := image.Decode(imgFile)
		if err != nil {
			return err
		}

		rasterConv := &raster.Converter{
			MaxWidth:  512,
			Threshold: 0.5,
		}
		data, _, _ := rasterConv.ToRaster(img)
		logoImg := printing.Image{}
		logoImg.Image = base64.StdEncoding.EncodeToString(data)
		logoImg.Width, logoImg.Height, err = printing.GetImageDimension(imagePath)
		if err != nil {
			return err
		}
		logoImg.Color = "color_1"
		logoImg.Mode = "mono"

		eposPrint.Image = append(eposPrint.Image, logoImg)

		// newline
		data, w, h, _ := p.TextToRaster(printing.Pad(len("newline")), 30.0, true)
		imgXML := ImgToXML(data, w, h)
		eposPrint.Image = append(eposPrint.Image, *imgXML)
	}

	// header
	folioHeader := FolioHeader(folio)
	eposPrint.Image = append(eposPrint.Image, folioHeader...)

	// table header
	folioTableHeader := FolioTableHeader(folio)
	eposPrint.Image = append(eposPrint.Image, folioTableHeader...)

	// table content
	folioTableContent := FolioTableContent(folio)
	eposPrint.Image = append(eposPrint.Image, folioTableContent...)

	// fdm
	if config.Config.IsFDMEnabled == true {
		fdmSection := FDMSection(folio)
		eposPrint.Image = append(eposPrint.Image, fdmSection...)
	}

	// footer
	folioFooter := FolioFooter(folio)
	eposPrint.Image = append(eposPrint.Image, folioFooter...)

	eposPrint.Text = append(eposPrint.Text, printing.Text{Text: "\n\n"})
	eposPrint.Cut.Type = "feed"
	xmlReq.Body.EposPrint = eposPrint
	reqBody, err := xml.Marshal(xmlReq)
	if err != nil {
		return err
	}
	// append xml header
	reqBody = []byte(xml.Header + string(reqBody))
	api := "http://" + *folio.Printer.PrinterIP + "/cgi-bin/epos/service.cgi?devid=" +
		folio.Printer.PrinterID + "&timeout=6000"
	printing.Send(api, reqBody)
	return nil
}
