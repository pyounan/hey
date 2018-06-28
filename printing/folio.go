package printing

import (
	"fmt"
	"pos-proxy/config"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cloudinn/escpos"
	"github.com/cloudinn/escpos/connection"
)

//PrintFolio to print folio recepit
func PrintFolio(folio *FolioPrint) error {

	printingParams := make(map[int]map[string]int)

	printingParams[80] = make(map[string]int)
	printingParams[80]["width"] = 800
	printingParams[80]["company_name_width"] = 2
	printingParams[80]["company_name_height"] = 2
	printingParams[80]["item_padding"] = 25
	printingParams[80]["qty_padding"] = 5
	printingParams[80]["price_padding"] = 8
	printingParams[80]["subtotal_padding"] = 40
	printingParams[80]["total_padding"] = 16
	printingParams[80]["fdm_rate_padding"] = 10
	printingParams[80]["fdm_taxable_padding"] = 10
	printingParams[80]["fdm_vat_padding"] = 10
	printingParams[80]["fdm_net_padding"] = 10
	printingParams[80]["tax_padding"] = 5
	printingParams[80]["char_per_line"] = 40

	printingParams[76] = make(map[string]int)
	printingParams[76]["width"] = 760
	printingParams[76]["company_name_width"] = 2
	printingParams[76]["company_name_height"] = 2
	printingParams[76]["item_padding"] = 25
	printingParams[76]["qty_padding"] = 5
	printingParams[76]["price_padding"] = 8
	printingParams[76]["subtotal_padding"] = 32
	printingParams[76]["total_padding"] = 15
	printingParams[76]["fdm_rate_padding"] = 8
	printingParams[76]["fdm_taxable_padding"] = 8
	printingParams[76]["fdm_vat_padding"] = 8
	printingParams[76]["fdm_net_padding"] = 8
	printingParams[76]["tax_padding"] = 5
	printingParams[76]["char_per_line"] = 32

	var p *escpos.Printer
	var err error
	if folio.Printer.IsUSB {
		p, err = connection.NewConnection("usb", *folio.Printer.PrinterIP)
		if err != nil {
			return err
		}
	} else {

		p, err = connection.NewConnection("network", *folio.Printer.PrinterIP)
		if err != nil {
			return err
		}
	}

	if folio.Store.Logo != "" {
		p.PrintImage(folio.Store.Logo)
	}
	p.SetImageHight(70)
	p.SetFontSizePoints(70)

	taxInvoiceRT := "Tax Invoice"
	proformaTR := "PRO FORMA"
	thisIsNotATR := "This is not a"
	validTaxInvoiceTR := "valid tax invoice"
	returnTR := "Return"

	if folio.Invoice.IsSettled {
		p.WriteString(Pad((16-utf8.RuneCountInString(taxInvoiceRT))/2) +
			CheckLang(strings.ToUpper(taxInvoiceRT)))
		if folio.Invoice.PaidAmount < 0 {
			p.WriteString(Pad((16-utf8.RuneCountInString(returnTR))/2) +
				CheckLang(strings.ToUpper(returnTR)))
		}
	} else {
		p.WriteString(Pad((16 - utf8.RuneCountInString(proformaTR)/2)) +
			CheckLang(strings.ToUpper(proformaTR)))
		if config.Config.IsFDMEnabled {
			p.WriteString(Pad((16-utf8.RuneCountInString(thisIsNotATR))/2) +
				CheckLang(strings.ToUpper(thisIsNotATR)))
			p.WriteString(Pad((16-utf8.RuneCountInString(validTaxInvoiceTR))/2) +
				CheckLang(strings.ToUpper(validTaxInvoiceTR)))
		}
	}
	p.Formfeed()
	p.WriteString(Pad((16-utf8.RuneCountInString(folio.Company.Name))/2) +
		CheckLang(folio.Company.Name))
	// p.SetFontSize(1, 1)
	p.SetImageHight(38)
	p.SetFontSizePoints(30)
	p.WriteString(Center(folio.Company.VATNumber,
		printingParams[folio.Printer.PaperWidth]["width"]) +
		CheckLang(folio.Company.VATNumber))
	p.WriteString(Center(folio.Company.Address,
		printingParams[folio.Printer.PaperWidth]["width"]) +
		CheckLang(folio.Company.Address))
	p.WriteString(Center(folio.Company.PostalCode+"-"+folio.Company.City,
		printingParams[folio.Printer.PaperWidth]["width"]) +
		CheckLang(folio.Company.PostalCode+"-"+folio.Company.City))
	var headers []string
	if folio.Store.InvoiceHeader != "" {
		headers = strings.Split(folio.Store.InvoiceHeader, "\n")
		for _, header := range headers {
			if utf8.RuneCountInString(header) >
				printingParams[folio.Printer.PaperWidth]["char_per_line"] {
				p.WriteString(Center(header[0:printingParams[folio.Printer.PaperWidth]["char_per_line"]],
					printingParams[folio.Printer.PaperWidth]["width"]) +
					CheckLang(header[0:printingParams[folio.Printer.PaperWidth]["char_per_line"]]))
				p.WriteString(Center(header[printingParams[folio.Printer.PaperWidth]["char_per_line"]:],
					printingParams[folio.Printer.PaperWidth]["width"]) +
					CheckLang(header[printingParams[folio.Printer.PaperWidth]["char_per_line"]:]))
			} else {

				p.WriteString(Center(header,
					printingParams[folio.Printer.PaperWidth]["width"]) +
					CheckLang(header))
			}
		}
	}
	p.WriteString(Center(folio.Store.Description,
		printingParams[folio.Printer.PaperWidth]["width"]) +
		CheckLang(folio.Store.Description))
	p.WriteString(Center("Invoice number: "+folio.Invoice.InvoiceNumber,
		printingParams[folio.Printer.PaperWidth]["width"]) +
		CheckLang("Invoice number: "+folio.Invoice.InvoiceNumber))
	// p.SetFontSize(2, 2)
	p.SetImageHight(70)
	p.SetFontSizePoints(70)
	p.WriteString(Pad((16-
		utf8.RuneCountInString("Covers: "+fmt.Sprintf("%d", folio.Invoice.Pax)))/2) +
		CheckLang("Covers: "+fmt.Sprintf("%d", folio.Invoice.Pax)))
	if folio.Invoice.TableID != nil {
		p.WriteString(Pad((16-
			utf8.RuneCountInString("Table: "+*folio.Invoice.TableDetails))/2) +
			CheckLang("Table: "+*folio.Invoice.TableDetails))
	} else {
		p.WriteString(Center("Takeout",
			printingParams[folio.Printer.PaperWidth]["width"]) +
			CheckLang("Takeout"))
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
	p.Formfeed()
	p.SetImageHight(38)
	p.SetFontSizePoints(30)
	if guestName != "" {
		p.WriteString(Center("Guest name: ",
			printingParams[folio.Printer.PaperWidth]["width"]) +
			CheckLang("Guest name: "))
		p.WriteString(Center(guestName,
			printingParams[folio.Printer.PaperWidth]["width"]) +
			CheckLang(guestName))
	}
	p.SetWhiteOnBlack(false)
	item := "Item"
	qty := "Qty"
	price := "Price"
	tableHeader := item + Pad(printingParams[folio.Printer.PaperWidth]["item_padding"]-
		utf8.RuneCountInString(item)) + qty +
		Pad(printingParams[folio.Printer.PaperWidth]["qty_padding"]-
			utf8.RuneCountInString(qty)) + price +
		Pad(printingParams[folio.Printer.PaperWidth]["price_padding"]-
			utf8.RuneCountInString(price))

	if config.Config.IsFDMEnabled {
		if printingParams[folio.Printer.PaperWidth]["width"] == 760 {
			p.SetFontSizePoints(22)
		} else {
			p.SetFontSizePoints(28)
		}
		tax := "Tax"
		p.WriteString(tableHeader + tax + Pad(printingParams[folio.Printer.PaperWidth]["tax_padding"]-
			utf8.RuneCountInString(tax)))
	} else {

		p.WriteString(tableHeader)
	}

	p.Formfeed()
	p.SetWhiteOnBlack(true)

	vatsToDisplay := map[string]bool{
		"A": false,
		"B": false,
		"C": false,
		"D": false,
	}

	for _, item := range folio.Invoice.Items {
		price := fmt.Sprintf("%.2f", item.Price)
		desc := CheckLang(item.Description)
		qty := CheckLang(fmt.Sprintf("%.2f", item.Quantity))
		text := desc + Pad(printingParams[folio.Printer.PaperWidth]["item_padding"]-
			utf8.RuneCountInString(desc)) + qty +
			Pad(printingParams[folio.Printer.PaperWidth]["qty_padding"]-
				utf8.RuneCountInString(qty)) + price +
			Pad(printingParams[folio.Printer.PaperWidth]["price_padding"]-
				utf8.RuneCountInString(price))

		if config.Config.IsFDMEnabled {
			if printingParams[folio.Printer.PaperWidth]["width"] == 760 {

				p.SetFontSizePoints(22)
			} else {
				p.SetFontSizePoints(28)
			}
			text += item.VAT + Pad(printingParams[folio.Printer.PaperWidth]["tax_padding"]-1)
			vatsToDisplay[item.VAT] = true
			p.WriteString(text)
		} else {
			p.SetFontSizePoints(30)
			p.WriteString(text)
		}
		p.Formfeed()
	}
	subTotal := fmt.Sprintf("%.2f", folio.Invoice.Subtotal)
	p.Formfeed()
	subTotalVal := ""
	if folio.Invoice.HouseUse {
		subTotalVal = "0.00"
	} else {
		subTotalVal = subTotal
	}
	p.SetFontSizePoints(30)
	subTotalTrans := "Subtotal"
	p.WriteString(CheckLang(subTotalTrans) +
		Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
			utf8.RuneCountInString(subTotalVal)-utf8.RuneCountInString(subTotalTrans)) +
		subTotalVal)

	if folio.TotalDiscounts > 0.0 {
		totalDiscount := fmt.Sprintf("%.2f", folio.TotalDiscounts)
		totalDiscountTrans := "Total discounts"
		p.WriteString(CheckLang(totalDiscountTrans) +
			Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
				utf8.RuneCountInString(totalDiscount)-utf8.RuneCountInString(totalDiscountTrans)) +
			totalDiscount)
	}

	total := fmt.Sprintf("%.2f", folio.Invoice.Total)
	p.SetImageHight(70)
	p.SetFontSizePoints(70)
	totalVal := "0.00"
	if folio.Invoice.HouseUse {
		totalVal = "0.00"
	} else {
		totalVal = total
	}
	totalTrans := "Total"
	p.WriteString(CheckLang(totalTrans) +
		Pad(printingParams[folio.Printer.PaperWidth]["total_padding"]-
			utf8.RuneCountInString(totalVal)-utf8.RuneCountInString(totalTrans)) +
		totalVal)
	p.SetImageHight(38)
	p.SetFontSizePoints(30)

	if folio.Invoice.IsSettled == true && folio.Invoice.Postings != nil &&
		len(folio.Invoice.Postings) > 0 {
		paymentTrans := "Payment"
		p.WriteString(CheckLang(paymentTrans) +
			Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
				utf8.RuneCountInString(paymentTrans)-utf8.RuneCountInString(total)) + total)
		for _, posting := range folio.Invoice.Postings {
			deptAmount := fmt.Sprintf("%.2f", posting.Amount)
			if posting.RoomNumber != nil || *posting.RoomNumber != 0 {
				if folio.Invoice.WalkinName != "" {
					guestName = folio.Invoice.WalkinName
				} else if folio.Invoice.ProfileDetails != "" {
					guestName = folio.Invoice.ProfileDetails
				} else if folio.Invoice.RoomDetails != nil || *folio.Invoice.RoomDetails != "" {
					guestName = *folio.Invoice.RoomDetails
				}
				p.WriteString(fmt.Sprintf("%d", posting.RoomNumber) +
					" " + CheckLang(guestName) +
					Pad(32-utf8.RuneCountInString(guestName)-
						utf8.RuneCountInString(fmt.Sprintf("%d", posting.RoomNumber))-
						utf8.RuneCountInString(deptAmount)) + deptAmount)
			} else {
				p.WriteString(CheckLang(posting.DepartmentDetails +
					Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
						utf8.RuneCountInString(posting.DepartmentDetails)-
						utf8.RuneCountInString(deptAmount)) + deptAmount))
			}
			if len(posting.GatewayResponses) > 0 {
				for _, response := range posting.GatewayResponses {
					p.WriteString(CheckLang(response))
				}
			}
		}
		p.Formfeed()
		received := "0.0"
		if folio.Invoice.Change != 0.0 {
			received = fmt.Sprintf("%.2f", folio.Invoice.Change+folio.Invoice.Total)
		} else {
			received = fmt.Sprintf("%.2f", folio.Invoice.Total)
		}
		receivedTrans := "Received"
		p.WriteString(CheckLang(receivedTrans) +
			Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
				utf8.RuneCountInString(receivedTrans)-utf8.RuneCountInString(received)) + received)
		change := "0.00"
		changeTrans := "Change"

		if folio.Invoice.Change != 0.0 {
			change = fmt.Sprintf("%.2f", folio.Invoice.Change)
		}
		p.WriteString(CheckLang(changeTrans) +
			Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
				utf8.RuneCountInString(change)-utf8.RuneCountInString(changeTrans)) + change)
	}
	p.Formfeed()

	if config.Config.IsFDMEnabled {
		taxableTrans := "Taxable"
		rateTrans := "Rate"
		vatTrans := "Vat"
		netTrans := "Net"
		for _, res := range folio.Invoice.FDMResponses {
			p.SetWhiteOnBlack(false)
			p.WriteString(rateTrans +
				Pad(printingParams[folio.Printer.PaperWidth]["fdm_rate_padding"]-
					utf8.RuneCountInString(rateTrans)) +
				Pad(printingParams[folio.Printer.PaperWidth]["fdm_taxable_padding"]-
					utf8.RuneCountInString(taxableTrans)) + taxableTrans +
				Pad(printingParams[folio.Printer.PaperWidth]["fdm_vat_padding"]-
					utf8.RuneCountInString(vatTrans)) + vatTrans +
				Pad(printingParams[folio.Printer.PaperWidth]["fdm_net_padding"]-
					utf8.RuneCountInString(netTrans)) + netTrans)
			p.SetWhiteOnBlack(true)
			for k, v := range vatsToDisplay {
				if v {
					taxableAmount := fmt.Sprintf("%.2f", res.VATSummary[k]["taxable_amount"])
					vatAmount := fmt.Sprintf("%.2f", res.VATSummary[k]["vat_amount"])
					netAmount := fmt.Sprintf("%.2f", res.VATSummary[k]["net_amount"])
					p.WriteString(k +
						Pad(printingParams[folio.Printer.PaperWidth]["fdm_rate_padding"]) +
						Pad(printingParams[folio.Printer.PaperWidth]["fdm_taxable_padding"]-
							utf8.RuneCountInString(taxableAmount)) + taxableAmount +
						Pad(printingParams[folio.Printer.PaperWidth]["fdm_vat_padding"]-
							utf8.RuneCountInString(vatAmount)) + vatAmount +
						Pad(printingParams[folio.Printer.PaperWidth]["fdm_net_padding"]-
							utf8.RuneCountInString(netAmount)) + netAmount)
				}
			}

			totalTaxableAmount := fmt.Sprintf("%.2f", res.VATSummary["Total"]["taxable_amount"])
			totalVatAmount := fmt.Sprintf("%.2f", res.VATSummary["Total"]["vat_amount"])
			totalNetAmount := fmt.Sprintf("%.2f", res.VATSummary["Total"]["net_amount"])
			totalTrans := "Total"
			p.WriteString(CheckLang(totalTrans) +
				Pad(printingParams[folio.Printer.PaperWidth]["fdm_rate_padding"]-
					utf8.RuneCountInString(totalTrans)) +
				Pad(printingParams[folio.Printer.PaperWidth]["fdm_taxable_padding"]-
					utf8.RuneCountInString(totalTaxableAmount)) + totalTaxableAmount +
				Pad(printingParams[folio.Printer.PaperWidth]["fdm_vat_padding"]-
					utf8.RuneCountInString(totalVatAmount)) + totalVatAmount +
				Pad(printingParams[folio.Printer.PaperWidth]["fdm_net_padding"]-
					utf8.RuneCountInString(totalNetAmount)) + totalNetAmount)
			p.Formfeed()
		}
	}

	loc, _ := time.LoadLocation(folio.Timezone)
	p.WriteString(CheckLang("Opened at: " + folio.Invoice.CreatedOn))
	if folio.Invoice.ClosedOn != nil {
		closedAt := folio.Invoice.ClosedOn.In(loc)
		closedAtStr := closedAt.Format("02 Jan 2006 15:04:05")
		p.SetFontSizePoints(28)
		p.WriteString(CheckLang("Closed on: " + closedAtStr))
		p.SetFontSizePoints(30)

	}
	p.WriteString(CheckLang("Created by: " + folio.Invoice.CashierDetails))
	p.WriteString(CheckLang("Printed by: " + fmt.Sprintf("%d", folio.Cashier.Number)))

	if config.Config.IsFDMEnabled {
		for _, res := range folio.Invoice.FDMResponses {
			p.WriteString(CheckLang("Ticket Number: " + res.TicketNumber))
			p.SetFontSizePoints(28)
			p.WriteString(CheckLang("Ticket Date: " + res.Date.Format("02 Jan 2006 15:04:05")))
			p.SetFontSizePoints(30)
			event := CheckLang("Event: ") + CheckLang(res.EventLabel)
			if utf8.RuneCountInString(event) > 32 {
				p.WriteString(event[0:32])
				p.WriteString(event[32:])
			} else {
				p.WriteString(event)
			}
			terminalIdentifier := CheckLang("Terminal Identifier: ") +
				CheckLang(folio.Terminal.RCRS) + "/" +
				CheckLang(folio.Terminal.Description)
			if utf8.RuneCountInString(terminalIdentifier) >
				printingParams[folio.Printer.PaperWidth]["char_per_line"] {
				p.WriteString(terminalIdentifier[0:32])
				p.WriteString(terminalIdentifier[32:])
			} else {
				p.WriteString(terminalIdentifier)
			}
			p.WriteString(CheckLang("Production Number: " + folio.Terminal.RCRS))
			p.WriteString(CheckLang("Software Version: " + res.SoftwareVersion))

			ticket := CheckLang("Ticket: ") +
				strings.Join(strings.Fields(res.TicketCounter), " ") +
				"/" + strings.Join(strings.Fields(res.TotalTicketCounter), " ") +
				" " + CheckLang(res.EventLabel)
			if utf8.RuneCountInString(ticket) > 32 {
				p.WriteString(ticket[0:32])
				p.WriteString(ticket[32:])
			} else {
				p.WriteString(ticket)
			}

			if utf8.RuneCountInString(res.PLUHash) > 32 {
				p.WriteString(CheckLang("Hash" + "s: " + res.PLUHash[0:25]))
				p.WriteString(Pad(7) + CheckLang(res.PLUHash[25:]))
			} else {
				p.WriteString(CheckLang("Hash" + ":" + res.PLUHash))
			}
			if folio.Invoice.IsSettled {
				p.WriteString(CheckLang("Ticket Sig" + ": " + res.Signature[0:20]))
				p.WriteString(Pad(13) + CheckLang(res.Signature[20:]))
			}
			p.SetFontSizePoints(28)
			p.WriteString(CheckLang("Control Data" + ": " + res.Date.String()[0:10] +
				" " + res.TimePeriod.Format("15:04:00")))
			p.SetFontSizePoints(30)
			p.WriteString(CheckLang("Control Module ID" + ": " + res.ProductionNumber))
			p.WriteString(CheckLang("VSC ID" + ": " + res.VSC))

		}
	}
	p.Formfeed()
	p.WriteString(CheckLang("Signature" + ":    " + "............."))
	p.Formfeed()
	if folio.Store.InvoiceFooter != "" {
		if utf8.RuneCountInString(folio.Store.InvoiceFooter) >
			printingParams[folio.Printer.PaperWidth]["char_per_line"] {
			p.WriteString(Center(folio.Store.InvoiceFooter[0:printingParams[folio.Printer.PaperWidth]["char_per_line"]],
				printingParams[folio.Printer.PaperWidth]["width"]) +
				CheckLang(folio.Store.InvoiceFooter[0:printingParams[folio.Printer.PaperWidth]["char_per_line"]]))
			p.WriteString(Center(folio.Store.InvoiceFooter[printingParams[folio.Printer.PaperWidth]["char_per_line"]:],
				printingParams[folio.Printer.PaperWidth]["width"]) +
				CheckLang(folio.Store.InvoiceFooter[printingParams[folio.Printer.PaperWidth]["char_per_line"]:]))
		} else {

			p.WriteString(Center(folio.Store.InvoiceFooter,
				printingParams[folio.Printer.PaperWidth]["width"]) + CheckLang(folio.Store.InvoiceFooter))
		}
	}
	p.Formfeed()
	p.Cut()
	return nil
}
