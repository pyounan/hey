package printing

import (
	"fmt"
	"pos-proxy/config"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/abadojack/whatlanggo"
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
	printingParams[80]["total_padding"] = 24
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
	printingParams[76]["total_padding"] = 19
	printingParams[76]["fdm_rate_padding"] = 8
	printingParams[76]["fdm_taxable_padding"] = 8
	printingParams[76]["fdm_vat_padding"] = 8
	printingParams[76]["fdm_net_padding"] = 8
	printingParams[76]["tax_padding"] = 5
	printingParams[76]["char_per_line"] = 32

	SetLang(folio.Terminal.RCRS)

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
	p.SetImageHight(52)
	p.SetFontSizePoints(50)

	taxInvoiceRT := Translate("Tax Invoice")
	proformaTR := Translate("PRO FORMA")
	thisIsNotATR := Translate("This is not a")
	validTaxInvoiceTR := Translate("valid tax invoice")
	returnTR := Translate("Return")

	if folio.Invoice.IsSettled {
		p.WriteString(Pad((25-utf8.RuneCountInString(taxInvoiceRT))/2) +
			CheckLang(strings.ToUpper(taxInvoiceRT)))
		if folio.Invoice.PaidAmount < 0 {
			p.WriteString(Pad((25-utf8.RuneCountInString(returnTR))/2) +
				CheckLang(strings.ToUpper(returnTR)))
		}
	} else {
		p.WriteString(Pad((25 - utf8.RuneCountInString(proformaTR)/2)) +
			CheckLang(strings.ToUpper(proformaTR)))
		if config.Config.IsFDMEnabled {
			p.WriteString(Pad((25-utf8.RuneCountInString(thisIsNotATR))/2) +
				CheckLang(strings.ToUpper(thisIsNotATR)))
			p.WriteString(Pad((25-utf8.RuneCountInString(validTaxInvoiceTR))/2) +
				CheckLang(strings.ToUpper(validTaxInvoiceTR)))
		}
	}
	p.Formfeed()
	p.WriteString(Pad((25-utf8.RuneCountInString(folio.Company.Name))/2) +
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
		CheckLang(Translate("Invoice number")+": "+folio.Invoice.InvoiceNumber))
	// p.SetFontSize(2, 2)
	p.SetImageHight(50)
	p.SetFontSizePoints(45)
	p.WriteString(Pad((25-
		utf8.RuneCountInString("Covers: "+fmt.Sprintf("%d", folio.Invoice.Pax)))/2) +
		CheckLang(Translate("Covers")+": "+fmt.Sprintf("%d", folio.Invoice.Pax)))
	if folio.Invoice.TableID != nil {
		p.WriteString(Pad((25-
			utf8.RuneCountInString("Table: "+*folio.Invoice.TableDetails))/2) +
			CheckLang(Translate("Table")+": "+*folio.Invoice.TableDetails))
	} else {
		p.WriteString(Pad(10) + CheckLang(Translate("Takeout")))
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
	guestNameTrans := Translate("Guest name")
	info := whatlanggo.Detect(guestNameTrans)
	if guestName != "" {
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(Center("Guest name: "+guestName,
				printingParams[folio.Printer.PaperWidth]["width"]) +
				CheckLang(guestName) + ": " + CheckLang(guestNameTrans))

		} else {
			p.WriteString(Center("Guest name: "+guestName,
				printingParams[folio.Printer.PaperWidth]["width"]) +
				CheckLang(guestNameTrans) + ": " + CheckLang(guestName))
		}
	}
	p.Formfeed()
	p.SetWhiteOnBlack(false)
	item := CheckLang(Translate("Item"))
	qty := CheckLang(Translate("Qty"))
	price := CheckLang(Translate("Price"))

	info = whatlanggo.Detect(item)
	tableHeader := ""
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		tableHeader = price + Pad(printingParams[folio.Printer.PaperWidth]["price_padding"]-
			utf8.RuneCountInString(price)) + qty +
			Pad(printingParams[folio.Printer.PaperWidth]["item_padding"]-
				utf8.RuneCountInString(qty)) + item +
			Pad(printingParams[folio.Printer.PaperWidth]["qty_padding"]-
				utf8.RuneCountInString(item))

	} else {

		tableHeader = item + Pad(printingParams[folio.Printer.PaperWidth]["item_padding"]-
			utf8.RuneCountInString(item)) + qty +
			Pad(printingParams[folio.Printer.PaperWidth]["qty_padding"]-
				utf8.RuneCountInString(qty)) + price +
			Pad(printingParams[folio.Printer.PaperWidth]["price_padding"]-
				utf8.RuneCountInString(price))
	}

	if config.Config.IsFDMEnabled {
		if printingParams[folio.Printer.PaperWidth]["width"] == 760 {
			p.SetFontSizePoints(22)
		} else {
			p.SetFontSizePoints(28)
		}
		tax := CheckLang(Translate("Tax"))
		p.WriteString(tableHeader + tax + Pad(printingParams[folio.Printer.PaperWidth]["tax_padding"]-
			utf8.RuneCountInString(tax)))
	} else {
		if printingParams[folio.Printer.PaperWidth]["width"] == 760 {
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
		desc := CheckLang(item.Description)
		qty := CheckLang(fmt.Sprintf("%.2f", item.Quantity))
		text := ""
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			text = price + Pad(printingParams[folio.Printer.PaperWidth]["price_padding"]-
				utf8.RuneCountInString(price)) + qty +
				Pad(printingParams[folio.Printer.PaperWidth]["item_padding"]-
					utf8.RuneCountInString(desc)) + desc +
				Pad(printingParams[folio.Printer.PaperWidth]["qty_padding"]-
					utf8.RuneCountInString(desc))

		} else {
			text = desc + Pad(printingParams[folio.Printer.PaperWidth]["item_padding"]-
				utf8.RuneCountInString(desc)) + qty +
				Pad(printingParams[folio.Printer.PaperWidth]["qty_padding"]-
					utf8.RuneCountInString(qty)) + price +
				Pad(printingParams[folio.Printer.PaperWidth]["price_padding"]-
					utf8.RuneCountInString(price))
		}
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
			if printingParams[folio.Printer.PaperWidth]["width"] == 760 {

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
	subTotalTrans := Translate("Subtotal")
	info = whatlanggo.Detect(subTotalTrans)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(subTotalVal +
			Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
				utf8.RuneCountInString(subTotalVal)-utf8.RuneCountInString(subTotalTrans)) +
			CheckLang(subTotalTrans))

	} else {
		p.WriteString(CheckLang(subTotalTrans) +
			Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
				utf8.RuneCountInString(subTotalVal)-utf8.RuneCountInString(subTotalTrans)) +
			subTotalVal)
	}
	if folio.TotalDiscounts > 0.0 {
		totalDiscount := fmt.Sprintf("%.2f", folio.TotalDiscounts)
		totalDiscountTrans := Translate("Total discounts")
		info = whatlanggo.Detect(totalDiscountTrans)
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(totalDiscount +
				Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
					utf8.RuneCountInString(totalDiscount)-utf8.RuneCountInString(totalDiscountTrans)) +
				CheckLang(totalDiscountTrans))

		} else {
			p.WriteString(CheckLang(totalDiscountTrans) +
				Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
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
	totalTrans := Translate("Total")
	info = whatlanggo.Detect(totalTrans)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(totalVal +
			Pad(printingParams[folio.Printer.PaperWidth]["total_padding"]-
				utf8.RuneCountInString(totalVal)-utf8.RuneCountInString(totalTrans)) +
			CheckLang(totalTrans))

	} else {
		p.WriteString(CheckLang(totalTrans) +
			Pad(printingParams[folio.Printer.PaperWidth]["total_padding"]-
				utf8.RuneCountInString(totalVal)-utf8.RuneCountInString(totalTrans)) +
			totalVal)
	}
	p.SetImageHight(38)
	p.SetFontSizePoints(30)

	if folio.Invoice.IsSettled == true && folio.Invoice.Postings != nil &&
		len(folio.Invoice.Postings) > 0 {
		paymentTrans := Translate("Payment")
		info = whatlanggo.Detect(paymentTrans)
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(total +
				Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
					utf8.RuneCountInString(paymentTrans)-utf8.RuneCountInString(total)) +
				CheckLang(paymentTrans))

		} else {
			p.WriteString(CheckLang(paymentTrans) +
				Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
					utf8.RuneCountInString(paymentTrans)-utf8.RuneCountInString(total)) + total)
		}
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
				if whatlanggo.Scripts[info.Script] == "Arabic" {
					p.WriteString(deptAmount +
						Pad(32-utf8.RuneCountInString(guestName)-
							utf8.RuneCountInString(fmt.Sprintf("%d", posting.RoomNumber))-
							utf8.RuneCountInString(deptAmount)) + fmt.Sprintf("%d", posting.RoomNumber) +
						" " + CheckLang(guestName))

				} else {
					if whatlanggo.Scripts[info.Script] == "Arabic" {
						p.WriteString(deptAmount +
							Pad(32-utf8.RuneCountInString(guestName)-
								utf8.RuneCountInString(fmt.Sprintf("%d", posting.RoomNumber))-
								utf8.RuneCountInString(deptAmount)) + fmt.Sprintf("%d", posting.RoomNumber) +
							" " + CheckLang(guestName))

					} else {
						p.WriteString(fmt.Sprintf("%d", posting.RoomNumber) +
							" " + CheckLang(guestName) +
							Pad(32-utf8.RuneCountInString(guestName)-
								utf8.RuneCountInString(fmt.Sprintf("%d", posting.RoomNumber))-
								utf8.RuneCountInString(deptAmount)) + deptAmount)
					}
				}
			} else {
				if whatlanggo.Scripts[info.Script] == "Arabic" {
					p.WriteString(deptAmount +
						Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
							utf8.RuneCountInString(posting.DepartmentDetails)-
							utf8.RuneCountInString(deptAmount)) +
						CheckLang(posting.DepartmentDetails))

				} else {
					p.WriteString(CheckLang(posting.DepartmentDetails +
						Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
							utf8.RuneCountInString(posting.DepartmentDetails)-
							utf8.RuneCountInString(deptAmount)) + deptAmount))
				}
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
		receivedTrans := Translate("Received")
		info = whatlanggo.Detect(receivedTrans)
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(received +
				Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
					utf8.RuneCountInString(receivedTrans)-utf8.RuneCountInString(received)) +
				CheckLang(receivedTrans))

		} else {
			p.WriteString(CheckLang(receivedTrans) +
				Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
					utf8.RuneCountInString(receivedTrans)-utf8.RuneCountInString(received)) +
				received)
		}
		change := "0.00"
		changeTrans := Translate("Change")

		if folio.Invoice.Change != 0.0 {
			change = fmt.Sprintf("%.2f", folio.Invoice.Change)
		}
		info = whatlanggo.Detect(changeTrans)
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(change +
				Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
					utf8.RuneCountInString(change)-utf8.RuneCountInString(changeTrans)) +
				CheckLang(changeTrans))

		} else {
			p.WriteString(CheckLang(changeTrans) +
				Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
					utf8.RuneCountInString(change)-utf8.RuneCountInString(changeTrans)) + change)
		}
	}
	p.Formfeed()

	if config.Config.IsFDMEnabled {
		taxableTrans := CheckLang(Translate("Taxable"))
		rateTrans := CheckLang(Translate("Rate"))
		vatTrans := CheckLang(Translate("Vat"))
		netTrans := CheckLang(Translate("Net"))
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
			totalTrans := Translate("Total")
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
	openedAt := Translate("Opened at")
	info = whatlanggo.Detect(openedAt)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
			utf8.RuneCountInString(folio.Invoice.CreatedOn)-
			utf8.RuneCountInString(openedAt)-2) +
			CheckLang(folio.Invoice.CreatedOn) + ": " + CheckLang(openedAt))

	} else {
		p.WriteString(CheckLang(openedAt) + ": " + CheckLang(folio.Invoice.CreatedOn))
	}
	if folio.Invoice.ClosedOn != nil {
		closedAt := folio.Invoice.ClosedOn.In(loc)
		closedAtStr := closedAt.Format("02 Jan 2006 15:04:05")
		p.SetFontSizePoints(28)
		cloasedOn := Translate("Closed on")
		info = whatlanggo.Detect(cloasedOn)
		if whatlanggo.Scripts[info.Script] == "Arabic" {
			p.WriteString(Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
				utf8.RuneCountInString(closedAtStr)-
				utf8.RuneCountInString(cloasedOn)+1) +
				CheckLang(closedAtStr) + ": " + CheckLang(cloasedOn))

		} else {
			p.WriteString(CheckLang(cloasedOn) + ": " + CheckLang(closedAtStr))
		}
		p.SetFontSizePoints(30)

	}
	createdBy := Translate("Created by")
	info = whatlanggo.Detect(createdBy)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
			utf8.RuneCountInString(folio.Invoice.CashierDetails)-
			utf8.RuneCountInString(createdBy)-2) +
			CheckLang(folio.Invoice.CashierDetails) + ": " + CheckLang(createdBy))

	} else {
		p.WriteString(CheckLang(createdBy) + ": " + CheckLang(folio.Invoice.CashierDetails))
	}
	printedBy := Translate("Printed by")
	info = whatlanggo.Detect(printedBy)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
			utf8.RuneCountInString(fmt.Sprintf("%d", folio.Cashier.Number))-
			utf8.RuneCountInString(printedBy)-2) +
			fmt.Sprintf("%d", folio.Cashier.Number) + ": " + CheckLang(printedBy))

	} else {
		p.WriteString(CheckLang(printedBy) + ": " + fmt.Sprintf("%d", folio.Cashier.Number))
	}
	if config.Config.IsFDMEnabled {
		for _, res := range folio.Invoice.FDMResponses {
			p.WriteString(CheckLang(Translate("Ticket Number") + ": " + res.TicketNumber))
			p.SetFontSizePoints(28)
			p.WriteString(CheckLang(Translate("Ticket Date") + ": " + res.Date.Format("02 Jan 2006 15:04:05")))
			p.SetFontSizePoints(30)
			event := CheckLang(Translate("Event")+": ") + CheckLang(res.EventLabel)
			if utf8.RuneCountInString(event) > 32 {
				p.WriteString(event[0:32])
				p.WriteString(event[32:])

			} else {
				p.WriteString(event)
			}
			terminalIdentifier := CheckLang(Translate("Terminal Identifier")+": ") +
				CheckLang(folio.Terminal.RCRS) + "/" +
				CheckLang(folio.Terminal.Description)
			if utf8.RuneCountInString(terminalIdentifier) >
				printingParams[folio.Printer.PaperWidth]["char_per_line"] {
				p.WriteString(terminalIdentifier[0:32])
				p.WriteString(terminalIdentifier[32:])
			} else {
				p.WriteString(terminalIdentifier)
			}
			p.WriteString(CheckLang(Translate("Production Number") + ": " + folio.Terminal.RCRS))
			p.WriteString(CheckLang(Translate("Software Version") + ": " + res.SoftwareVersion))

			ticket := CheckLang(Translate("Ticket")+": ") +
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
				p.WriteString(CheckLang(Translate("Hash") + "s: " + res.PLUHash[0:25]))
				p.WriteString(Pad(7) + CheckLang(res.PLUHash[25:]))
			} else {
				p.WriteString(CheckLang(Translate("Hash") + ":" + res.PLUHash))
			}
			if folio.Invoice.IsSettled {
				p.WriteString(CheckLang(Translate("Ticket Sig") + ": " + res.Signature[0:20]))
				p.WriteString(Pad(13) + CheckLang(res.Signature[20:]))
			}
			p.SetFontSizePoints(28)
			p.WriteString(CheckLang(Translate("Control Data") + ": " + res.Date.String()[0:10] +
				" " + res.TimePeriod.Format("15:04:00")))
			p.SetFontSizePoints(30)
			p.WriteString(CheckLang(Translate("Control Module ID") + ": " + res.ProductionNumber))
			p.WriteString(CheckLang(Translate("VSC ID") + ": " + res.VSC))

		}
	}
	p.Formfeed()
	signature := Translate("Signature")
	info = whatlanggo.Detect(signature)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		p.WriteString(Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
			utf8.RuneCountInString(signature)-18) +
			"............." + "    :" + CheckLang(signature))

	} else {
		p.WriteString(CheckLang(signature) + ":    " + ".............")
	}
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
