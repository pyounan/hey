package esc

import (
	"fmt"
	"pos-proxy/config"
	"pos-proxy/printing"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/abadojack/whatlanggo"
	"github.com/cloudinn/escpos"
	"github.com/cloudinn/escpos/connection"
)

func FolioHeader(folio *printing.FolioPrint, p *escpos.Printer) error {

	lang := printing.SetLang(folio.Terminal.RCRS)
	if folio.Store.Logo != "" {
		imageFile, err := printing.GetImage(folio.Store.Logo)
		if err != nil {
			return err
		}
		p.PrintImage(imageFile)
	}
	p.SetImageHight(52)
	p.SetFontSizePoints(50)

	taxInvoiceRT := printing.Translate("Tax Invoice", lang)
	proformaTR := printing.Translate("PRO FORMA", lang)
	thisIsNotATR := printing.Translate("This is not a", lang)
	validTaxInvoiceTR := printing.Translate("valid tax invoice", lang)
	returnTR := printing.Translate("Return", lang)

	if folio.Invoice.IsSettled {
		p.WriteString(printing.Pad((25-utf8.RuneCountInString(taxInvoiceRT))/2) +
			printing.CheckLang(strings.ToUpper(taxInvoiceRT)))
		if folio.Invoice.PaidAmount < 0 {
			p.WriteString(printing.Pad((25-utf8.RuneCountInString(returnTR))/2) +
				printing.CheckLang(strings.ToUpper(returnTR)))
		}
	} else {
		p.WriteString(printing.Pad((25-utf8.RuneCountInString(proformaTR))/2) +
			printing.CheckLang(strings.ToUpper(proformaTR)))
		if config.Config.IsFDMEnabled {
			p.WriteString(printing.Pad((25-utf8.RuneCountInString(thisIsNotATR))/2) +
				printing.CheckLang(strings.ToUpper(thisIsNotATR)))
			p.WriteString(printing.Pad((25-utf8.RuneCountInString(validTaxInvoiceTR))/2) +
				printing.CheckLang(strings.ToUpper(validTaxInvoiceTR)))
		}
	}
	p.Formfeed()
	p.WriteString(printing.Pad((25-utf8.RuneCountInString(folio.Company.Name))/2) +
		printing.CheckLang(folio.Company.Name))
	p.SetImageHight(38)
	p.SetFontSizePoints(30)
	p.WriteString(printing.Center(folio.Company.VATNumber,
		printing.PrintingParams(folio.Printer.PaperWidth, "width")) +
		printing.CheckLang(folio.Company.VATNumber))
	p.WriteString(printing.Center(folio.Company.Address,
		printing.PrintingParams(folio.Printer.PaperWidth, "width")) +
		printing.CheckLang(folio.Company.Address))
	p.WriteString(printing.Center(folio.Company.PostalCode+"-"+folio.Company.City,
		printing.PrintingParams(folio.Printer.PaperWidth, "width")) +
		printing.CheckLang(folio.Company.PostalCode+"-"+folio.Company.City))
	var headers []string
	if folio.Store.InvoiceHeader != "" {
		headers = strings.Split(folio.Store.InvoiceHeader, "\n")
		for _, header := range headers {
			if utf8.RuneCountInString(header) >
				printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line") {
				p.WriteString(printing.Center(header[0:printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line")],
					printing.PrintingParams(folio.Printer.PaperWidth, "width")) +
					printing.CheckLang(header[0:printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line")]))
				p.WriteString(printing.Center(header[printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line"):],
					printing.PrintingParams(folio.Printer.PaperWidth, "width")) +
					printing.CheckLang(header[printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line"):]))
			} else {

				p.WriteString(printing.Center(header,
					printing.PrintingParams(folio.Printer.PaperWidth, "width")) +
					printing.CheckLang(header))
			}
		}
	}
	p.WriteString(printing.Center(folio.Store.Description,
		printing.PrintingParams(folio.Printer.PaperWidth, "width")) +
		printing.CheckLang(folio.Store.Description))
	p.WriteString(printing.Center("Invoice number: "+folio.Invoice.InvoiceNumber,
		printing.PrintingParams(folio.Printer.PaperWidth, "width")) +
		printing.CheckLang(printing.Translate("Invoice number", lang)+": "+folio.Invoice.InvoiceNumber))
	// p.SetFontSize(2, 2)
	p.SetImageHight(50)
	p.SetFontSizePoints(45)
	p.WriteString(printing.Pad((25-
		utf8.RuneCountInString("Covers: "+fmt.Sprintf("%d", folio.Invoice.Pax)))/2) +
		printing.CheckLang(printing.Translate("Covers", lang)+": "+fmt.Sprintf("%d", folio.Invoice.Pax)))
	if folio.Invoice.TableID != nil {
		p.WriteString(printing.Pad((25-
			utf8.RuneCountInString("Table: "+*folio.Invoice.TableDetails))/2) +
			printing.CheckLang(printing.Translate("Table", lang)+": "+*folio.Invoice.TableDetails))
	} else {
		p.WriteString(printing.Pad((25-utf8.RuneCountInString("Takeout"))/2) + printing.CheckLang(printing.Translate("Takeout", lang)))
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
	p.SetImageHight(38)
	p.SetFontSizePoints(30)
	guestNameTrans := printing.Translate("Guest name", lang)
	info := whatlanggo.Detect(guestNameTrans)
	if guestName != "" {
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(printing.Center("Guest name: "+guestName,
				printing.PrintingParams(folio.Printer.PaperWidth, "width")) +
				printing.CheckLang(guestName) + ": " + printing.CheckLang(guestNameTrans))

		} else {
			p.WriteString(printing.Center("Guest name: "+guestName,
				printing.PrintingParams(folio.Printer.PaperWidth, "width")) +
				printing.CheckLang(guestNameTrans) + ": " + printing.CheckLang(guestName))
		}
	}
	p.Formfeed()

	return nil
}
func FolioTable(folio *printing.FolioPrint, p *escpos.Printer) error {
	lang := printing.SetLang(folio.Terminal.RCRS)

	p.SetWhiteOnBlack(false)
	item := printing.CheckLang(printing.Translate("Item", lang))
	qty := printing.CheckLang(printing.Translate("Qty", lang))
	price := printing.CheckLang(printing.Translate("Price", lang))

	info := whatlanggo.Detect(item)
	tableHeader := ""
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		tableHeader = price + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "price_padding")-
			utf8.RuneCountInString(price)) + qty +
			printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "item_padding")-
				utf8.RuneCountInString(qty)) + item +
			printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "qty_padding")-
				utf8.RuneCountInString(item))

	} else {

		tableHeader = item + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "item_padding")-
			utf8.RuneCountInString(item)) + qty +
			printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "qty_padding")-
				utf8.RuneCountInString(qty)) + price +
			printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "price_padding")-
				utf8.RuneCountInString(price))
	}

	if config.Config.IsFDMEnabled {
		if printing.PrintingParams(folio.Printer.PaperWidth, "width") == 760 {
			p.SetFontSizePoints(22)
		} else {
			p.SetFontSizePoints(28)
		}
		tax := printing.CheckLang(printing.Translate("Tax", lang))
		p.WriteString(tableHeader + tax + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "tax_padding")-
			utf8.RuneCountInString(tax)))
	} else {
		if printing.PrintingParams(folio.Printer.PaperWidth, "width") == 760 {
			p.SetFontSizePoints(26)
		} else {
			p.SetFontSizePoints(28)
		}

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
		desc := printing.CheckLang(item.Description)
		qty := printing.CheckLang(fmt.Sprintf("%.2f", item.Quantity))
		text := ""
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			text = price + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "price_padding")-
				utf8.RuneCountInString(price)) + qty +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "item_padding")-
					utf8.RuneCountInString(desc)) + desc +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "qty_padding")-
					utf8.RuneCountInString(desc))

		} else {
			text = desc + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "item_padding")-
				utf8.RuneCountInString(desc)) + qty +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "qty_padding")-
					utf8.RuneCountInString(qty)) + price +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "price_padding")-
					utf8.RuneCountInString(price))
		}
		if config.Config.IsFDMEnabled {
			if printing.PrintingParams(folio.Printer.PaperWidth, "width") == 760 {

				p.SetFontSizePoints(22)
			} else {
				p.SetFontSizePoints(28)
			}
			text += item.VAT + printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "tax_padding")-1)
			vatsToDisplay[item.VAT] = true
			p.WriteString(text)
		} else {
			if printing.PrintingParams(folio.Printer.PaperWidth, "width") == 760 {

				p.SetFontSizePoints(26)
			} else {
				p.SetFontSizePoints(28)
			}
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
	subTotalTrans := printing.Translate("Subtotal", lang)
	info = whatlanggo.Detect(subTotalTrans)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(subTotalVal +
			printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
				utf8.RuneCountInString(subTotalVal)-utf8.RuneCountInString(subTotalTrans)) +
			printing.CheckLang(subTotalTrans))

	} else {
		p.WriteString(printing.CheckLang(subTotalTrans) +
			printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
				utf8.RuneCountInString(subTotalVal)-utf8.RuneCountInString(subTotalTrans)) +
			subTotalVal)
	}
	if folio.TotalDiscounts > 0.0 {
		totalDiscount := fmt.Sprintf("%.2f", folio.TotalDiscounts)
		totalDiscountTrans := printing.Translate("Total discounts", lang)
		info = whatlanggo.Detect(totalDiscountTrans)
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(totalDiscount +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
					utf8.RuneCountInString(totalDiscount)-utf8.RuneCountInString(totalDiscountTrans)) +
				printing.CheckLang(totalDiscountTrans))

		} else {
			p.WriteString(printing.CheckLang(totalDiscountTrans) +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
					utf8.RuneCountInString(totalDiscount)-utf8.RuneCountInString(totalDiscountTrans)) +
				totalDiscount)
		}
	}

	total := fmt.Sprintf("%.2f", folio.Invoice.Total)
	p.SetImageHight(50)
	p.SetFontSizePoints(50)
	totalVal := "0.00"
	if folio.Invoice.HouseUse {
		totalVal = "0.00"
	} else {
		totalVal = total
	}
	totalTrans := printing.Translate("Total", lang)
	info = whatlanggo.Detect(totalTrans)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(totalVal +
			printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "total_padding")-
				utf8.RuneCountInString(totalVal)-utf8.RuneCountInString(totalTrans)) +
			printing.CheckLang(totalTrans))

	} else {
		p.WriteString(printing.CheckLang(totalTrans) +
			printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "total_padding")-
				utf8.RuneCountInString(totalVal)-utf8.RuneCountInString(totalTrans)) +
			totalVal)
	}
	p.SetImageHight(38)
	p.SetFontSizePoints(30)
	guestName := ""
	if folio.Invoice.IsSettled == true && folio.Invoice.Postings != nil &&
		len(folio.Invoice.Postings) > 0 {
		paymentTrans := printing.Translate("Payment", lang)
		info = whatlanggo.Detect(paymentTrans)
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(total +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
					utf8.RuneCountInString(paymentTrans)-utf8.RuneCountInString(total)) +
				printing.CheckLang(paymentTrans))

		} else {
			p.WriteString(printing.CheckLang(paymentTrans) +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
					utf8.RuneCountInString(paymentTrans)-utf8.RuneCountInString(total)) + total)
		}
		for _, posting := range folio.Invoice.Postings {
			deptAmount := fmt.Sprintf("%.2f", posting.Amount)
			if posting.RoomNumber != nil && *posting.RoomNumber != 0 {

				if folio.Invoice.WalkinName != "" {
					guestName = folio.Invoice.WalkinName
				} else if folio.Invoice.ProfileDetails != "" {
					guestName = folio.Invoice.ProfileDetails
				} else if folio.Invoice.RoomDetails != nil || *folio.Invoice.RoomDetails != "" {
					guestName = *folio.Invoice.RoomDetails
				}
				if whatlanggo.Scripts[info.Script] == "Arabic" {
					p.WriteString(deptAmount +
						printing.Pad(32-utf8.RuneCountInString(guestName)-
							utf8.RuneCountInString(fmt.Sprintf("%d", posting.RoomNumber))-
							utf8.RuneCountInString(deptAmount)) + fmt.Sprintf("%d", posting.RoomNumber) +
						" " + printing.CheckLang(guestName))

				} else {
					if whatlanggo.Scripts[info.Script] == "Arabic" {
						p.WriteString(deptAmount +
							printing.Pad(32-utf8.RuneCountInString(guestName)-
								utf8.RuneCountInString(fmt.Sprintf("%d", posting.RoomNumber))-
								utf8.RuneCountInString(deptAmount)) + fmt.Sprintf("%d", posting.RoomNumber) +
							" " + printing.CheckLang(guestName))

					} else {
						p.WriteString(fmt.Sprintf("%d", posting.RoomNumber) +
							" " + printing.CheckLang(guestName) +
							printing.Pad(32-utf8.RuneCountInString(guestName)-
								utf8.RuneCountInString(fmt.Sprintf("%d", posting.RoomNumber))-
								utf8.RuneCountInString(deptAmount)) + deptAmount)
					}
				}
			} else {
				if whatlanggo.Scripts[info.Script] == "Arabic" {
					p.WriteString(deptAmount +
						printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
							utf8.RuneCountInString(posting.DepartmentDetails)-
							utf8.RuneCountInString(deptAmount)) +
						printing.CheckLang(posting.DepartmentDetails))

				} else {
					p.WriteString(printing.CheckLang(posting.DepartmentDetails +
						printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
							utf8.RuneCountInString(posting.DepartmentDetails)-
							utf8.RuneCountInString(deptAmount)) + deptAmount))
				}
			}
			if len(posting.GatewayResponses) > 0 {
				for _, response := range posting.GatewayResponses {
					p.WriteString(printing.CheckLang(response))
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
		receivedTrans := printing.Translate("Received", lang)
		info = whatlanggo.Detect(receivedTrans)
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(received +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
					utf8.RuneCountInString(receivedTrans)-utf8.RuneCountInString(received)) +
				printing.CheckLang(receivedTrans))

		} else {
			p.WriteString(printing.CheckLang(receivedTrans) +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
					utf8.RuneCountInString(receivedTrans)-utf8.RuneCountInString(received)) +
				received)
		}
		change := "0.00"
		changeTrans := printing.Translate("Change", lang)

		if folio.Invoice.Change != 0.0 {
			change = fmt.Sprintf("%.2f", folio.Invoice.Change)
		}
		info = whatlanggo.Detect(changeTrans)
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(change +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
					utf8.RuneCountInString(change)-utf8.RuneCountInString(changeTrans)) +
				printing.CheckLang(changeTrans))

		} else {
			p.WriteString(printing.CheckLang(changeTrans) +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
					utf8.RuneCountInString(change)-utf8.RuneCountInString(changeTrans)) + change)
		}
	}
	p.Formfeed()
	return nil
}

func FDMSection(folio *printing.FolioPrint, p *escpos.Printer) error {

	lang := printing.SetLang(folio.Terminal.RCRS)

	vatsToDisplay := map[string]bool{
		"A": false,
		"B": false,
		"C": false,
		"D": false,
	}
	if config.Config.IsFDMEnabled {
		taxableTrans := printing.CheckLang(printing.Translate("Taxable", lang))
		rateTrans := printing.CheckLang(printing.Translate("Rate", lang))
		vatTrans := printing.CheckLang(printing.Translate("Vat", lang))
		netTrans := printing.CheckLang(printing.Translate("Net", lang))
		for _, res := range folio.Invoice.FDMResponses {
			p.SetWhiteOnBlack(false)
			p.WriteString(rateTrans +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_rate_padding")-
					utf8.RuneCountInString(rateTrans)) +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_taxable_padding")-
					utf8.RuneCountInString(taxableTrans)) + taxableTrans +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_vat_padding")-
					utf8.RuneCountInString(vatTrans)) + vatTrans +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_net_padding")-
					utf8.RuneCountInString(netTrans)) + netTrans)
			p.SetWhiteOnBlack(true)
			for k, v := range vatsToDisplay {
				if v {
					taxableAmount := fmt.Sprintf("%.2f", res.VATSummary[k]["taxable_amount"])
					vatAmount := fmt.Sprintf("%.2f", res.VATSummary[k]["vat_amount"])
					netAmount := fmt.Sprintf("%.2f", res.VATSummary[k]["net_amount"])
					p.WriteString(k +
						printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_rate_padding")) +
						printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_taxable_padding")-
							utf8.RuneCountInString(taxableAmount)) + taxableAmount +
						printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_vat_padding")-
							utf8.RuneCountInString(vatAmount)) + vatAmount +
						printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_net_padding")-
							utf8.RuneCountInString(netAmount)) + netAmount)
				}
			}

			totalTaxableAmount := fmt.Sprintf("%.2f", res.VATSummary["Total"]["taxable_amount"])
			totalVatAmount := fmt.Sprintf("%.2f", res.VATSummary["Total"]["vat_amount"])
			totalNetAmount := fmt.Sprintf("%.2f", res.VATSummary["Total"]["net_amount"])
			totalTrans := printing.Translate("Total", lang)
			p.WriteString(printing.CheckLang(totalTrans) +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_rate_padding")-
					utf8.RuneCountInString(totalTrans)) +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_taxable_padding")-
					utf8.RuneCountInString(totalTaxableAmount)) + totalTaxableAmount +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_vat_padding")-
					utf8.RuneCountInString(totalVatAmount)) + totalVatAmount +
				printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "fdm_net_padding")-
					utf8.RuneCountInString(totalNetAmount)) + totalNetAmount)
			p.Formfeed()
		}
	}
	return nil
}
func FolioFooter(folio *printing.FolioPrint, p *escpos.Printer) error {

	lang := printing.SetLang(folio.Terminal.RCRS)

	loc, _ := time.LoadLocation(folio.Timezone)
	openedAt := printing.Translate("Opened at", lang)
	info := whatlanggo.Detect(openedAt)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
			utf8.RuneCountInString(folio.Invoice.CreatedOn)-
			utf8.RuneCountInString(openedAt)-2) +
			printing.CheckLang(folio.Invoice.CreatedOn) + ": " + printing.CheckLang(openedAt))

	} else {
		p.WriteString(printing.CheckLang(openedAt) + ": " + printing.CheckLang(folio.Invoice.CreatedOn))
	}
	if folio.Invoice.ClosedOn != nil {
		closedAt := folio.Invoice.ClosedOn.In(loc)
		closedAtStr := closedAt.Format("02 Jan 2006 15:04:05")
		p.SetFontSizePoints(28)
		cloasedOn := printing.Translate("Closed on", lang)
		info = whatlanggo.Detect(cloasedOn)
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
				utf8.RuneCountInString(closedAtStr)-
				utf8.RuneCountInString(cloasedOn)+1) +
				printing.CheckLang(closedAtStr) + ": " + printing.CheckLang(cloasedOn))

		} else {
			p.WriteString(printing.CheckLang(cloasedOn) + ": " + printing.CheckLang(closedAtStr))
		}
		p.SetFontSizePoints(30)

	}
	createdBy := printing.Translate("Created by", lang)
	info = whatlanggo.Detect(createdBy)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
			utf8.RuneCountInString(folio.Invoice.CashierDetails)-
			utf8.RuneCountInString(createdBy)-2) +
			printing.CheckLang(folio.Invoice.CashierDetails) + ": " + printing.CheckLang(createdBy))

	} else {
		p.WriteString(printing.CheckLang(createdBy) + ": " + printing.CheckLang(folio.Invoice.CashierDetails))
	}
	printedBy := printing.Translate("Printed by", lang)
	info = whatlanggo.Detect(printedBy)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
			utf8.RuneCountInString(fmt.Sprintf("%d", folio.Cashier.Number))-
			utf8.RuneCountInString(printedBy)-2) +
			fmt.Sprintf("%d", folio.Cashier.Number) + ": " + printing.CheckLang(printedBy))

	} else {
		p.WriteString(printing.CheckLang(printedBy) + ": " + fmt.Sprintf("%d", folio.Cashier.Number))
	}
	if config.Config.IsFDMEnabled {
		for _, res := range folio.Invoice.FDMResponses {
			p.WriteString(printing.CheckLang(printing.Translate("Ticket Number", lang) + ": " + res.TicketNumber))
			p.SetFontSizePoints(28)
			p.WriteString(printing.CheckLang(printing.Translate("Ticket Date", lang) + ": " + res.Date.Format("02 Jan 2006 15:04:05")))
			p.SetFontSizePoints(30)
			event := printing.CheckLang(printing.Translate("Event", lang)) + ": " + printing.CheckLang(res.EventLabel)
			if utf8.RuneCountInString(event) > 32 {
				p.WriteString(event[0:32])
				p.WriteString(event[32:])

			} else {
				p.WriteString(event)
			}
			terminalIdentifier := printing.CheckLang(printing.Translate("Terminal Identifier", lang)+": ") +
				printing.CheckLang(folio.Terminal.RCRS) + "/" +
				printing.CheckLang(folio.Terminal.Description)
			if utf8.RuneCountInString(terminalIdentifier) >
				printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line") {
				p.WriteString(terminalIdentifier[0:32])
				p.WriteString(terminalIdentifier[32:])
			} else {
				p.WriteString(terminalIdentifier)
			}
			p.WriteString(printing.CheckLang(printing.Translate("Production Number", lang) + ": " + folio.Terminal.RCRS))
			p.WriteString(printing.CheckLang(printing.Translate("Software Version", lang) + ": " + res.SoftwareVersion))

			ticket := printing.CheckLang(printing.Translate("Ticket", lang)) + ": " +
				strings.Join(strings.Fields(res.TicketCounter), " ") +
				"/" + strings.Join(strings.Fields(res.TotalTicketCounter), " ") +
				" " + printing.CheckLang(res.EventLabel)
			if utf8.RuneCountInString(ticket) > 32 {
				p.WriteString(ticket[0:32])
				p.WriteString(ticket[32:])
			} else {
				p.WriteString(ticket)
			}

			if utf8.RuneCountInString(res.PLUHash) > 32 {
				p.WriteString(printing.CheckLang(printing.Translate("Hash", lang) + "s: " + res.PLUHash[0:25]))
				p.WriteString(printing.Pad(7) + printing.CheckLang(res.PLUHash[25:]))
			} else {
				p.WriteString(printing.CheckLang(printing.Translate("Hash", lang) + ":" + res.PLUHash))
			}
			if folio.Invoice.IsSettled {
				p.WriteString(printing.CheckLang(printing.Translate("Ticket Sig", lang) + ": " + res.Signature[0:20]))
				p.WriteString(printing.Pad(13) + printing.CheckLang(res.Signature[20:]))
			}
			p.SetFontSizePoints(28)
			p.WriteString(printing.CheckLang(printing.Translate("Control Data", lang) + ": " + res.Date.String()[0:10] +
				" " + res.TimePeriod.Format("15:04:00")))
			p.SetFontSizePoints(30)
			p.WriteString(printing.CheckLang(printing.Translate("Control Module ID", lang) + ": " + res.ProductionNumber))
			p.WriteString(printing.CheckLang(printing.Translate("VSC ID", lang) + ": " + res.VSC))

		}
	}
	p.Formfeed()
	signature := printing.Translate("Signature", lang)
	info = whatlanggo.Detect(signature)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(printing.Pad(printing.PrintingParams(folio.Printer.PaperWidth, "subtotal_padding")-
			utf8.RuneCountInString(signature)-18) +
			"............." + "    :" + printing.CheckLang(signature))

	} else {
		p.WriteString(printing.CheckLang(signature) + ":    " + ".............")
	}
	p.Formfeed()
	if folio.Store.InvoiceFooter != "" {
		if utf8.RuneCountInString(folio.Store.InvoiceFooter) >
			printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line") {
			p.WriteString(printing.Center(folio.Store.InvoiceFooter[0:printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line")],
				printing.PrintingParams(folio.Printer.PaperWidth, "width")) +
				printing.CheckLang(folio.Store.InvoiceFooter[0:printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line")]))
			p.WriteString(printing.Center(folio.Store.InvoiceFooter[printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line"):],
				printing.PrintingParams(folio.Printer.PaperWidth, "width")) +
				printing.CheckLang(folio.Store.InvoiceFooter[printing.PrintingParams(folio.Printer.PaperWidth, "char_per_line"):]))
		} else {

			p.WriteString(printing.Center(folio.Store.InvoiceFooter,
				printing.PrintingParams(folio.Printer.PaperWidth, "width")) + printing.CheckLang(folio.Store.InvoiceFooter))
		}
	}
	p.Formfeed()
	return nil
}

//PrintFolio to print folio recepit
func (e Esc) PrintFolio(folio *printing.FolioPrint) error {

	var p *escpos.Printer
	var err error
	if folio.Printer.IsUSB {
		p, err = connection.NewConnection("usb", *folio.Printer.PrinterIP+":9100")
		if err != nil {
			return err
		}
	} else {

		p, err = connection.NewConnection("network", *folio.Printer.PrinterIP+":9100")
		if err != nil {
			return err
		}
	}

	FolioHeader(folio, p)
	FolioTable(folio, p)
	FDMSection(folio, p)
	FolioFooter(folio, p)

	p.Cut()
	return nil
}
