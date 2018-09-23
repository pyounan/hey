package epsonxml

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"image"
	"os"
	"pos-proxy/config"
	"pos-proxy/printing"
	"strings"
	"time"

	"github.com/cloudinn/escpos/raster"

	"unicode/utf8"
)

func FolioHeader(folio *printing.FolioPrint) []printing.Text {

	folioHeader := []printing.Text{}

	headerText := ""
	if folio.Invoice.IsSettled {
		headerText = strings.ToUpper(printing.Translate("Tax Invoice")) + "\n\n"
		if folio.Invoice.PaidAmount < 0 {
			headerText += strings.ToUpper(printing.Translate("Return")) + "\n"
		}
	} else {
		headerText = strings.ToUpper(printing.Translate("proforma")) + "\n\n"
		if config.Config.IsFDMEnabled {
			headerText += strings.ToUpper(printing.Translate("This is not a")) + "\n"
			headerText += strings.ToUpper(printing.Translate("vaild tax invoice")) + "\n"
		}
	}
	headersSlice := []printing.Text{}
	var headers []string
	if folio.Store.InvoiceHeader != "" {
		headers = strings.Split(folio.Store.InvoiceHeader, "\n")
		for _, header := range headers {
			if utf8.RuneCountInString(header) > printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line") {
				headersSlice = append(
					headersSlice,
					printing.Text{Text: printing.Translate(header[0:printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line")])},
					printing.Text{Text: printing.Translate(header[printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line"):])},
				)
			} else {
				headersSlice = append(
					headersSlice,
					printing.Text{Text: printing.Translate(header) + "\n"},
				)
			}

		}
	}
	tableTakeout := ""
	if folio.Invoice.TableID != nil {
		tableTakeout = printing.Translate("Table") + ": " + *folio.Invoice.TableDetails
	} else {
		tableTakeout = printing.Translate("Takeout")
	}
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
	folioHeader = append(
		folioHeader,
		printing.Text{Align: "center"},
		printing.Text{DoubleHight: "true"},
		printing.Text{DoubleWidth: "true"},
		//logothing
		printing.Text{Text: headerText},
		printing.Text{Text: "\n"},
		printing.Text{Text: folio.Company.Name + "\n"},
		printing.Text{DoubleHight: "false"},
		printing.Text{DoubleWidth: "false"},
		printing.Text{Text: folio.Company.VATNumber + "\n"},
		printing.Text{Text: folio.Company.Address + "\n"},
		printing.Text{Text: folio.Company.PostalCode + "-" + folio.Company.City + "\n"},
	)
	folioHeader = append(folioHeader, headersSlice...)
	folioHeader = append(
		folioHeader,
		printing.Text{Text: folio.Store.Description + "\n"},
		printing.Text{Text: printing.Translate("Invoice number") + ": " + folio.Invoice.InvoiceNumber + "\n"},
		printing.Text{DoubleHight: "true"},
		printing.Text{DoubleWidth: "true"},
		printing.Text{Text: printing.Translate("Covers") + ":" + fmt.Sprintf("%d", folio.Invoice.Pax) + "\n"},
		printing.Text{Text: tableTakeout + "\n"},
		printing.Text{DoubleHight: "false"},
		printing.Text{DoubleWidth: "false"},
		printing.Text{Text: printing.Translate("Guest name") + ": " + guestName + "\n"},
		printing.Text{Text: "\n"},
	)

	return folioHeader
}
func FolioTableHeader(folio *printing.FolioPrint) []printing.Text {
	tableHeader := []printing.Text{}
	item := printing.Translate("Item")
	qty := printing.Translate("Qty")
	price := printing.Translate("Price")
	tableHeaderText := item + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "item_padding")-utf8.RuneCountInString(item)) +
		qty + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "qty_padding")-utf8.RuneCountInString(qty)) +
		price + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "price_padding")-utf8.RuneCountInString(price))
	if config.Config.IsFDMEnabled {
		tax := printing.Translate("Tax")
		tableHeaderText += tax + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "tax_padding")-utf8.RuneCountInString(tax))
	}
	tableHeader = append(
		tableHeader,
		printing.Text{Align: "left"},
		printing.Text{Reverse: "true"},
		printing.Text{UnderLine: "false"},
		printing.Text{Emphasized: "true"},
		printing.Text{Color: "color_1"},
		printing.Text{Text: tableHeaderText + "\n"},
		printing.Text{Reverse: "false"},
		printing.Text{UnderLine: "false"},
		printing.Text{Emphasized: "false"},
		printing.Text{Color: "color_1"},
	)

	return tableHeader
}
func FolioTableContent(folio *printing.FolioPrint) []printing.Text {
	tableContent := []printing.Text{}
	vatsToDisplay := map[string]bool{
		"A": false,
		"B": false,
		"C": false,
		"D": false,
	}
	folioItems := []printing.Text{}
	for _, item := range folio.Invoice.Items {
		price := fmt.Sprintf("%.2f", item.Price)
		desc := item.Description
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
		folioItems = append(folioItems, printing.Text{Text: row + "\n"})
	}
	subTotal := fmt.Sprintf("%.2f", folio.Invoice.Subtotal)
	subTotalVal := ""
	if folio.Invoice.HouseUse {
		subTotalVal = "0.00"
	} else {
		subTotalVal = subTotal
	}
	subTotalTrans := printing.Translate("Subtotal")
	subTotalText := subTotalTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
		utf8.RuneCountInString(subTotalVal)-utf8.RuneCountInString(subTotalTrans)) +
		subTotalVal
	totalDiscountText := ""
	if folio.TotalDiscounts > 0.0 {
		totalDiscount := fmt.Sprintf("%.2f", folio.TotalDiscounts)
		totalDiscountTrans := printing.Translate("Total discounts")
		totalDiscountText = totalDiscountTrans +
			printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
				utf8.RuneCountInString(totalDiscount)-utf8.RuneCountInString(totalDiscountTrans)) +
			totalDiscount
	}
	total := fmt.Sprintf("%.2f", folio.Invoice.Total)
	totalVal := "0.00"
	if folio.Invoice.HouseUse {
		totalVal = "0.00"
	} else {
		totalVal = total
	}
	totalTrans := printing.Translate("Total")
	totalText := totalTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "total_padding")-
		utf8.RuneCountInString(totalVal)-utf8.RuneCountInString(totalTrans)) + totalVal
	paymentText := ""
	if folio.Invoice.IsSettled == true && folio.Invoice.Postings != nil &&
		len(folio.Invoice.Postings) > 0 {
		paymentTrans := printing.Translate("Payment")
		paymentText = paymentTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
			utf8.RuneCountInString(paymentTrans)-utf8.RuneCountInString(total)) + total
	}
	postingSlice := []printing.Text{}
	for _, posting := range folio.Invoice.Postings {
		postingText := ""
		guestName := ""
		deptAmount := fmt.Sprintf("%.2f", posting.Amount)
		if posting.RoomNumber != nil && *posting.RoomNumber != 0 {

			if folio.Invoice.WalkinName != "" {
				guestName = folio.Invoice.WalkinName
			} else if folio.Invoice.ProfileDetails != "" {
				guestName = folio.Invoice.ProfileDetails
			} else if folio.Invoice.RoomDetails != nil || *folio.Invoice.RoomDetails != "" {
				guestName = *folio.Invoice.RoomDetails
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
		postingSlice = append(postingSlice, printing.Text{Text: postingText + "\n"})
		if len(posting.GatewayResponses) > 0 {
			for _, response := range posting.GatewayResponses {
				postingSlice = append(postingSlice, printing.Text{Text: response + "\n"})
			}
		}
	}
	received := "0.0"
	if folio.Invoice.Change != 0.0 {
		received = fmt.Sprintf("%.2f", folio.Invoice.Change+folio.Invoice.Total)
	} else {
		received = fmt.Sprintf("%.2f", folio.Invoice.Total)
	}
	receivedTrans := printing.Translate("Received")
	receivedText := receivedTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
		utf8.RuneCountInString(receivedTrans)-utf8.RuneCountInString(received)) +
		received

	change := "0.00"
	changeTrans := printing.Translate("Change")

	if folio.Invoice.Change != 0.0 {
		change = fmt.Sprintf("%.2f", folio.Invoice.Change)
	}
	changeText := changeTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
		utf8.RuneCountInString(change)-utf8.RuneCountInString(changeTrans)) + change

	tableContent = append(tableContent, folioItems...)
	tableContent = append(
		tableContent,
		printing.Text{Text: subTotalText + "\n"},
		printing.Text{Text: totalDiscountText + "\n"},
		printing.Text{DoubleHight: "true"},
		printing.Text{DoubleWidth: "true"},
		printing.Text{Text: totalText + "\n"},
		printing.Text{DoubleHight: "false"},
		printing.Text{DoubleWidth: "false"},
		printing.Text{Text: paymentText + "\n"},
	)
	tableContent = append(tableContent, postingSlice...)
	tableContent = append(
		tableContent,
		printing.Text{Text: "\n"},
		printing.Text{Text: receivedText + "\n"},
		printing.Text{Text: changeText + "\n"},
	)

	return tableContent
}

func FDMSection(folio *printing.FolioPrint) []printing.Text {
	fdm := []printing.Text{}
	vatsToDisplay := map[string]bool{
		"A": false,
		"B": false,
		"C": false,
		"D": false,
	}
	if config.Config.IsFDMEnabled {
		taxableTrans := printing.Translate("Taxable")
		rateTrans := printing.Translate("Rate")
		vatTrans := printing.Translate("Vat")
		netTrans := printing.Translate("Net")
		fdmTable := []printing.Text{}
		for _, res := range folio.Invoice.FDMResponses {
			fdmTableHeader := rateTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_rate_padding")-
				utf8.RuneCountInString(rateTrans)) +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_taxable_padding")-
					utf8.RuneCountInString(taxableTrans)) + taxableTrans +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_vat_padding")-
					utf8.RuneCountInString(vatTrans)) + vatTrans +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_net_padding")-
					utf8.RuneCountInString(netTrans)) + netTrans
			fdmTable = append(
				fdmTable,
				printing.Text{Reverse: "true"},
				printing.Text{UnderLine: "false"},
				printing.Text{Emphasized: "true"},
				printing.Text{Color: "color_1"},
				printing.Text{Text: fdmTableHeader},
				printing.Text{Reverse: "false"},
				printing.Text{UnderLine: "false"},
				printing.Text{Emphasized: "false"},
				printing.Text{Color: "color_1"},
				printing.Text{Text: "\n"})
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
					fdmTable = append(
						fdmTable,
						printing.Text{Text: fdmTableContent + "\n"},
					)
				}
			}

			totalTaxableAmount := fmt.Sprintf("%.2f", res.VATSummary["Total"]["taxable_amount"])
			totalVatAmount := fmt.Sprintf("%.2f", res.VATSummary["Total"]["vat_amount"])
			totalNetAmount := fmt.Sprintf("%.2f", res.VATSummary["Total"]["net_amount"])
			totalTrans := printing.Translate("Total")
			totalText := totalTrans + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_rate_padding")-
				utf8.RuneCountInString(totalTrans)) +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_taxable_padding")-
					utf8.RuneCountInString(totalTaxableAmount)) + totalTaxableAmount +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_vat_padding")-
					utf8.RuneCountInString(totalVatAmount)) + totalVatAmount +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_net_padding")-
					utf8.RuneCountInString(totalNetAmount)) + totalNetAmount
			fdmTable = append(fdmTable, printing.Text{Text: totalText + "\n"})
		}
		fdm = append(fdm, fdmTable...)
		fdm = append(fdm, printing.Text{Text: "\n"})
	}
	return fdm
}
func FolioFooter(folio *printing.FolioPrint) []printing.Text {
	footer := []printing.Text{}
	loc, _ := time.LoadLocation(folio.Timezone)
	openedAt := printing.Translate("Opened at")
	closedOnText := ""
	fdm := []printing.Text{}
	if folio.Invoice.ClosedOn != nil {
		closedAt := folio.Invoice.ClosedOn.In(loc)
		closedAtStr := closedAt.Format("02 Jan 2006 15:04:05")
		cloasedOn := printing.Translate("Closed on")
		closedOnText = cloasedOn + ": " + closedAtStr
	}
	createdBy := printing.Translate("Created by")
	printedBy := printing.Translate("Printed by")
	fdm = append(
		fdm,
		printing.Text{Text: openedAt + ": " + folio.Invoice.CreatedOn + "\n"},
		printing.Text{Text: closedOnText + "\n"},
		printing.Text{Text: createdBy + ": " + folio.Invoice.CashierDetails + "\n"},
		printing.Text{Text: printedBy + ": " + fmt.Sprintf("%d", folio.Cashier.Number) + "\n"},
	)
	if config.Config.IsFDMEnabled {
		for _, res := range folio.Invoice.FDMResponses {
			fdmResponse := []printing.Text{}
			ticketNumber := printing.Translate("Ticket Number") + ": " + res.TicketNumber + "\n"
			ticketDate := printing.Translate("Ticket Date") + ": " + res.Date.Format("02 Jan 2006 15:04:05") + "\n"
			event := printing.Translate("Event") + ": " + res.EventLabel + "\n"
			eventSlice := []printing.Text{}
			if utf8.RuneCountInString(event) > 32 {
				eventSlice = append(
					eventSlice,
					printing.Text{Text: event[0:32] + "\n"},
					printing.Text{Text: event[32:] + "\n"})

			} else {
				eventSlice = append(eventSlice, printing.Text{Text: event + "\n"})
			}

			terminalSlice := []printing.Text{}
			terminalIdentifier := printing.Translate("Terminal Identifier") + ": " +
				folio.Terminal.RCRS + "/" + folio.Terminal.Description
			if utf8.RuneCountInString(terminalIdentifier) >
				printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line") {
				terminalSlice = append(terminalSlice, printing.Text{Text: terminalIdentifier[0:32] + "\n"},
					printing.Text{Text: terminalIdentifier[32:] + "\n"})

			} else {
				terminalSlice = append(terminalSlice, printing.Text{Text: terminalIdentifier + "\n"})

			}
			productionNumber := printing.Translate("Production Number") + ": " + folio.Terminal.RCRS + "\n"
			SWVersion := printing.Translate("Software Version") + ": " + res.SoftwareVersion + "\n"

			ticket := printing.Translate("Ticket") + ": " +
				strings.Join(strings.Fields(res.TicketCounter), " ") +
				"/" + strings.Join(strings.Fields(res.TotalTicketCounter), " ") +
				" " + res.EventLabel
			ticketSlice := []printing.Text{}
			if utf8.RuneCountInString(ticket) > 32 {
				ticketSlice = append(ticketSlice,
					printing.Text{Text: ticket[0:32] + "\n"},
					printing.Text{Text: ticket[32:] + "\n"})
			} else {
				ticketSlice = append(ticketSlice, printing.Text{Text: ticket + "\n"})
			}

			hashSlice := []printing.Text{}
			if utf8.RuneCountInString(res.PLUHash) > 32 {
				hashSlice = append(
					hashSlice,
					printing.Text{Text: printing.Translate("Hash") + "s: " + res.PLUHash[0:25] + "\n"},
					printing.Text{Text: printing.Pad(7) + res.PLUHash[25:] + "\n"},
				)
			} else {
				hashSlice = append(hashSlice, printing.Text{Text: printing.Translate("Hash") + ": " + res.PLUHash + "\n"})
			}
			ticketSigSlice := []printing.Text{}
			if folio.Invoice.IsSettled {
				ticketSigSlice = append(
					ticketSigSlice,
					printing.Text{Text: printing.Translate("Ticket Sig") + ": " + res.Signature[0:20] + "\n"},
					printing.Text{Text: printing.Pad(13) + res.Signature[20:] + "\n"},
				)
			}
			controlData := printing.Translate("Control Data") + ": " + res.Date.String()[0:10] +
				" " + res.TimePeriod.Format("15:04:00") + "\n"
			controlModule := printing.Translate("Control Module ID") + ": " + res.ProductionNumber + "\n"
			vsc := printing.Translate("VSC ID") + ": " + res.VSC + "\n"

			fdmResponse = append(fdmResponse, printing.Text{Text: ticketNumber}, printing.Text{Text: ticketDate})
			fdmResponse = append(fdmResponse, eventSlice...)
			fdmResponse = append(fdmResponse, terminalSlice...)
			fdmResponse = append(fdmResponse, printing.Text{Text: productionNumber}, printing.Text{Text: SWVersion})
			fdmResponse = append(fdmResponse, ticketSlice...)
			fdmResponse = append(fdmResponse, hashSlice...)
			fdmResponse = append(fdmResponse, ticketSigSlice...)
			fdmResponse = append(
				fdmResponse,
				printing.Text{Text: controlData},
				printing.Text{Text: controlModule},
				printing.Text{Text: vsc},
			)
			fdm = append(fdm, fdmResponse...)
		}

	}
	signature := printing.Translate("Signature")
	sigText := signature + ":    " + "............." + "\n"
	footerSlice := []printing.Text{}
	if folio.Store.InvoiceFooter != "" {
		if utf8.RuneCountInString(folio.Store.InvoiceFooter) >
			printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line") {
			footerSlice = append(
				footerSlice,
				printing.Text{Text: folio.Store.InvoiceFooter[0:printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line")] + "\n"},
				printing.Text{Text: folio.Store.InvoiceFooter[printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line"):] + "\n"},
			)
		} else {
			footerSlice = append(
				footerSlice,
				printing.Text{Text: folio.Store.InvoiceFooter + "\n"},
			)
		}
	}

	footer = append(footer, fdm...)
	footer = append(footer, printing.Text{Text: sigText}, printing.Text{Align: "center"})
	footer = append(footer, footerSlice...)
	return footer

}

//PrintFolio Prints xml format of folio receipt
func (e Epsonxml) PrintFolio(folio *printing.FolioPrint) error {
	xmlReq := printing.New()
	xmlReq.XMLns = "http://schemas.xmlsoap.org/soap/envelope/"
	eposPrint := printing.EposPrint{}
	eposPrint.XMLns = "http://www.epson-pos.com/schemas/2011/03/epos-print"
	eposPrint.Align = &printing.Text{Align: "center"}
	if folio.Store.Logo != "" {
		imagePath, err := printing.GetImage(folio.Store.Logo)
		if err != nil {
			return err
		}
		imgFile, err := os.Open(imagePath)
		if err != nil {
			return err
		}
		img, _, err := image.Decode(imgFile)
		imgFile.Close()
		if err != nil {
			return err
		}
		rasterConv := &raster.Converter{
			MaxWidth:  512,
			Threshold: 0.5,
		}
		data, _, _ := rasterConv.ToRaster(img)
		eposPrint.Image = &printing.Image{}
		eposPrint.Image.Image = base64.StdEncoding.EncodeToString(data)
		eposPrint.Image.Width, eposPrint.Image.Height, err = printing.GetImageDimension(imagePath)
		if err != nil {
			return err
		}
		eposPrint.Image.Color = "color_1"
		eposPrint.Image.Mode = "mono"
	}
	eposPrint.Text = append(eposPrint.Text, printing.Text{Text: "\n\n"})
	folioHeader := FolioHeader(folio)
	folioTableHeader := FolioTableHeader(folio)
	folioTableContent := FolioTableContent(folio)
	fdmSection := []printing.Text{}
	if config.Config.IsFDMEnabled == true {
		fdmSection = FDMSection(folio)
	}
	folioFooter := FolioFooter(folio)
	eposPrint.Text = append(eposPrint.Text, folioHeader...)
	eposPrint.Text = append(eposPrint.Text, folioTableHeader...)
	eposPrint.Text = append(eposPrint.Text, folioTableContent...)
	eposPrint.Text = append(eposPrint.Text, fdmSection...)
	eposPrint.Text = append(eposPrint.Text, folioFooter...)
	eposPrint.Text = append(eposPrint.Text, printing.Text{Text: "\n\n"})
	eposPrint.Cut.Type = "feed"
	xmlReq.Body.EposPrint = eposPrint
	reqBody, err := xml.Marshal(xmlReq)
	if err != nil {
		return err
	}
	api := "http://" + *folio.Printer.PrinterIP + "/cgi-bin/epos/service.cgi?devid=" +
		folio.Printer.PrinterID + "&timeout=6000"
	printing.Send(api, reqBody)
	return nil
}
