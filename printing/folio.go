package printing

import (
	"fmt"
	"pos-proxy/config"
	"strings"
	"time"

	"github.com/cloudinn/escpos"
	"github.com/cloudinn/escpos/connection"
)

//Pad insert padding between text
func Pad(size int) string {
	if size > 0 {
		return strings.Join(make([]string, size), " ")

	}
	return ""
}

//PrintFolio to print folio recepit
func PrintFolio(folio *FolioPrint) {

	printingParams := make(map[int]map[string]int)

	printingParams[80] = make(map[string]int)
	printingParams[80]["width"] = 800
	printingParams[80]["company_name_width"] = 2
	printingParams[80]["company_name_height"] = 2
	printingParams[80]["item_padding"] = 25
	printingParams[80]["qty_padding"] = 5
	printingParams[80]["price_padding"] = 8
	printingParams[80]["subtotal_padding"] = 40
	printingParams[80]["total_padding"] = 21
	printingParams[80]["fdm_rate_padding"] = 10
	printingParams[80]["fdm_taxable_padding"] = 10
	printingParams[80]["fdm_vat_padding"] = 10
	printingParams[80]["fdm_net_padding"] = 10
	printingParams[80]["tax_padding"] = 5

	printingParams[76] = make(map[string]int)
	printingParams[76]["width"] = 760
	printingParams[76]["company_name_width"] = 2
	printingParams[76]["company_name_height"] = 2
	printingParams[76]["item_padding"] = 19
	printingParams[76]["qty_padding"] = 5
	printingParams[76]["price_padding"] = 10
	printingParams[76]["subtotal_padding"] = 32
	printingParams[76]["total_padding"] = 17
	printingParams[76]["fdm_rate_padding"] = 9
	printingParams[76]["fdm_taxable_padding"] = 13
	printingParams[76]["fdm_vat_padding"] = 12
	printingParams[76]["fdm_net_padding"] = 8
	printingParams[76]["tax_padding"] = 5

	var p *escpos.Printer

	// if folio.Printer.IsUSB {
	// 	p = connection.NewConnection("usb", *folio.Printer.PrinterIP)
	// }

	// p = connection.NewConnection("usb", *folio.Printer.PrinterIP)
	p = connection.NewConnection("network", *folio.Printer.PrinterIP)

	p.SetAlign("center")
	p.SetFontSize(byte(printingParams[folio.Printer.PaperWidth]["company_name_width"]),
		byte(printingParams[folio.Printer.PaperWidth]["company_name_height"]))

	//Disabled by Design because it has a bug

	// if folio.Store.Logo !=nil ||  folio.Store.Logo != ""{
	// 	p.PrintImage(logoPath)
	// }

	p.SetFontSize(2, 2)

	taxInvoiceRT := "Tax Invoice"
	proformaTR := "PRO FORMA"
	thisIsNotATR := "This is not a"
	validTaxInvoiceTR := "valid tax invoice"
	returnTR := "Return"

	if folio.Invoice.IsSettled {
		p.WriteString(strings.ToUpper(taxInvoiceRT) + "\n\n")
		if folio.Invoice.PaidAmount < 0 {
			p.WriteString(strings.ToUpper(returnTR) + "\n")
		}
	} else {
		p.WriteString(strings.ToUpper(proformaTR) + "\n\n")
		if config.Config.IsFDMEnabled {
			p.WriteString(strings.ToUpper(thisIsNotATR) + "\n")
			p.WriteString(strings.ToUpper(validTaxInvoiceTR) + "\n")
		}
	}
	p.Formfeed()
	p.WriteString(folio.Company.Name + "\n")
	p.SetFontSize(1, 1)
	p.WriteString(folio.Company.VATNumber + "\n")
	p.WriteString(folio.Company.Address + "\n")
	p.WriteString(folio.Company.PostalCode + "-" + folio.Company.City + "\n")
	var headers []string
	if folio.Store.InvoiceHeader != "" {
		headers = strings.Split(folio.Store.InvoiceHeader, "\n")
		for _, header := range headers {
			p.WriteString(header + "\n")
		}
	}
	p.WriteString(folio.Store.Description + "\n")
	p.WriteString("Invoice number : " + folio.Invoice.InvoiceNumber + "\n")
	p.SetFontSize(2, 2)
	p.WriteString("Covers : " + fmt.Sprintf("%d", folio.Invoice.Pax) + "\n")
	if folio.Invoice.TableID != nil {
		p.WriteString("Table: " + *folio.Invoice.TableDetails + "\n")
	} else {
		p.WriteString("Takeout\n")
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
	p.SetFontSize(1, 1)
	if guestName != "" {
		p.WriteString("Guest name : " + guestName + "\n")
	}
	p.SetAlign("left")
	p.SetFont("A")
	p.SetReverse(1)
	p.SetEmphasize(1)
	item := "Item"
	qty := "Qty"
	price := "Price"

	p.WriteString(item + Pad(printingParams[folio.Printer.PaperWidth]["item_padding"]-len(item)) +
		" " + qty + Pad(printingParams[folio.Printer.PaperWidth]["qty_padding"]-len(qty)) + " " +
		price + Pad(printingParams[folio.Printer.PaperWidth]["price_padding"]-len(price)) + " ")

	if config.Config.IsFDMEnabled {
		tax := "Tax"
		p.WriteString(tax + Pad(printingParams[folio.Printer.PaperWidth]["tax_padding"]-len(tax)) + " ")
	}
	p.Formfeed()
	p.SetReverse(0)
	p.SetEmphasize(0)

	vatsToDisplay := make(map[string]bool)
	vatsToDisplay["A"] = false
	vatsToDisplay["B"] = false
	vatsToDisplay["C"] = false
	vatsToDisplay["D"] = false

	for _, item := range folio.Invoice.Items {
		price := fmt.Sprintf("%.2f", item.Price)
		desc := item.Description
		text := desc + Pad(printingParams[folio.Printer.PaperWidth]["item_padding"]-len(desc)) + " " +
			fmt.Sprintf("%.2f", item.Quantity) + Pad(printingParams[folio.Printer.PaperWidth]["qty_padding"]-len(fmt.Sprintf("%f", item.Quantity))) + " " +
			price + Pad(printingParams[folio.Printer.PaperWidth]["price_padding"]-len(string(price))) + " "

		if config.Config.IsFDMEnabled {
			text += item.VAT + Pad(printingParams[folio.Printer.PaperWidth]["tax_padding"]-1) + " "
			vatsToDisplay[item.VAT] = true
		}
		p.WriteString(text)
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
	subTotalTrans := "Subtotal"
	p.WriteString(subTotalTrans + Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
		len(subTotalVal)-len(subTotalTrans)) + " " + subTotalVal + "\n")

	if folio.TotalDiscounts > 0.0 {
		totalDiscount := fmt.Sprintf("%.2f", folio.TotalDiscounts)
		totalDiscountTrans := "Total discounts"
		p.WriteString(totalDiscountTrans + Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
			len(totalDiscount)-len(totalDiscountTrans)) + " " + totalDiscount + "\n")
	}

	total := fmt.Sprintf("%.2f", folio.Invoice.Total)
	p.SetFontSize(2, 2)
	totalVal := "0.00"
	if folio.Invoice.HouseUse {
		totalVal = "0.00"
	} else {
		totalVal = total
	}
	totalTrans := "Total"
	p.WriteString(totalTrans + Pad(printingParams[folio.Printer.PaperWidth]["total_padding"]-
		len(totalVal)-len(totalTrans)) + " " + totalVal + "\n")
	p.SetFontSize(1, 1)

	if folio.Invoice.IsSettled == true && folio.Invoice.Postings != nil && len(folio.Invoice.Postings) > 0 {
		paymentTrans := "Payment"
		p.WriteString(paymentTrans + Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
			len(paymentTrans)-len(total)) + " " + total + "\n")
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
				p.WriteString(fmt.Sprintf("%d", posting.RoomNumber) + " " + guestName +
					Pad(39-len(guestName)-len(fmt.Sprintf("%d", posting.RoomNumber))-len(deptAmount)) + " " + "\n")
			} else {
				p.WriteString(posting.DepartmentDetails +
					Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-
						len(posting.DepartmentDetails)-len(deptAmount)) + " " + deptAmount + "\n")
			}
			if len(posting.GatewayResponses) > 0 {
				for _, response := range posting.GatewayResponses {
					p.WriteString(response + "\n")
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
		p.WriteString(receivedTrans +
			Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-len(receivedTrans)-
				len(received)) + " " + "\n")
		change := "0.00"
		changeTrans := "Change"

		if folio.Invoice.Change != 0.0 {
			change = fmt.Sprintf("%.2f", folio.Invoice.Change)
		}
		p.WriteString(changeTrans +
			Pad(printingParams[folio.Printer.PaperWidth]["subtotal_padding"]-len(change)-
				len(changeTrans)) + " " + change + "\n")
	}
	p.Formfeed()

	// if config.Config.IsFDMEnabled {
	// 	taxableTrans := "Taxable"
	// 	rateTrans := "Rate"
	// 	vatTrans := "Vat"
	// 	netTrans := "Net"
	// 	for res := range folio.Invoice.FDMResponses {
	// 		p.SetReverse(1)
	// 		p.SetEmphasize(1)
	// 		p.SetFont("B")
	// 		p.WriteString(rateTrans +
	// 			Pad(printingParams[folio.Printer.PaperWidth]["fdm_rate_padding"]-
	// 				len(rateTrans)) + " " +
	// 			Pad(printingParams[folio.Printer.PaperWidth]["fdm_taxable_padding"]-
	// 				len(taxableTrans)) + " " +
	// 			Pad(printingParams[folio.Printer.PaperWidth]["fdm_vat_padding"]-
	// 				len(vatTrans)) + " " +
	// 			Pad(printingParams[folio.Printer.PaperWidth]["fdm_net_padding"]-
	// 				len(netTrans)) + " " + "\n")
	// 		p.SetReverse(0)
	// 		p.SetEmphasize(0)

	// 		//Vat amount section
	// 		for k, v := range vatsToDisplay {
	// 			if v {
	// 				taxableAmount := fmt.Sprintf("%.2f", res.VATSummary[k].txabelAmount)
	// 				vatAmount := fmt.Sprintf("%.2f", res.VATSummary[k].vatAmount)
	// 				netAmount := fmt.Sprintf("%.2f", res.VATSummary[k].netAmount)
	// 				p.WriteString(k +
	// 					Pad(printingParams[folio.Printer.PaperWidth]["fdm_rate_padding"]-1) +
	// 					" " + Pad(printingParams[folio.Printer.PaperWidth]["fdm_taxable_padding"]-
	// 					len(string(taxableAmount))) + " " + taxableAmount +
	// 					Pad(printingParams[folio.Printer.PaperWidth]["fdm_vat_padding"]-
	// 						len(string(vatAmount))) + " " + vatAmount +
	// 					Pad(printingParams[folio.Printer.PaperWidth]["fdm_net_padding"]-
	// 						len(string(netAmount))) + " " + netAmount + "\n")
	// 			}
	// 		}

	// 		totalTaxableAmount := fmt.Sprintf("%.2f", res.VATSummary.Total.taxableAmount)
	// 		totalVatAmount := fmt.Sprintf("%.2f", res.VATSummary.Total.vatAmount)
	// 		totalNetAmount := fmt.Sprintf("%.2f", res.VATSummary.Total.netAmount)
	// 		p.SetEmphasize(1)
	// 		totalTrans := "Total"
	// 		p.WriteString(totalTrans +
	// 			Pad(printingParams[folio.Printer.PaperWidth]["fdm_rate_padding"]-len(totalTrans)) +
	// 			" " + Pad(printingParams[folio.Printer.PaperWidth]["fdm_taxable_padding"]-
	// 			len(string(totalTaxableAmount))) + " " + totalTaxableAmount +
	// 			Pad(printingParams[folio.Printer.PaperWidth]["fdm_vat_padding"]-
	// 				len(string(totalVatAmount))) + " " + totalVatAmount +
	// 			Pad(printingParams[folio.Printer.PaperWidth]["fdm_net_padding"]-
	// 				len(string(totalNetAmount))) + " " + totalNetAmount + "\n")
	// 		p.Formfeed()
	// 		p.SetEmphasize(0)

	// 	}
	// }

	loc, _ := time.LoadLocation(folio.Timezone)
	p.WriteString("Opened at: " + folio.Invoice.CreatedOn + "\n")
	if folio.Invoice.ClosedOn != nil {
		closedAt := folio.Invoice.ClosedOn.In(loc)
		closedAtStr := closedAt.Format(time.RFC1123)
		p.WriteString("Closed on: " + closedAtStr + "\n")

	}
	p.WriteString("Created by : " + folio.Invoice.CashierDetails + "\n")
	p.WriteString("Printed by : " + fmt.Sprintf("%d", folio.Cashier.Number) + "\n")

	// if config.Config.IsFDMEnabled {
	// 	for res := range folio.Invoice.FDMResponses {
	// 		p.WriteString("Ticket Number: " + res.TicketNumber + "\n")
	// 		p.WriteString("Ticket Date: " + res.Date.String() + "\n")
	// 		p.WriteString("Event: " + res.EventLabel + "\n")
	// 		p.WriteString("Terminal Identifier: " +
	// 			folio.Terminal.RCRS + "/" + folio.Terminal.Description + "\n")
	// 		p.WriteString("Production Number: " + folio.Terminal.RCRS + "\n")
	// 		p.WriteString("Software Version: " + res.SoftwareVersion + "\n")
	// 		p.WriteString("Ticket: " + strings.Join(strings.Fields(res.TicketCounter), " ") +
	// 			"/" + strings.Join(strings.Fields(res.TotalTicketCounter), " ") + " " +
	// 			res.EventLabel[0:30] + "\n")

	// 		if len(res.PLUHash) > 32 {
	// 			p.WriteString("Hash" + "s: " + res.PLUHash[0:31] + "\n")
	// 			p.WriteString(Pad(7) + " " + res.PLUHash[31:] + "\n")
	// 		} else {
	// 			p.WriteString("Hash" + ":" + res.PLUHash + "\n")
	// 		}
	// 		if folio.Invoice.IsSettled {
	// 			p.WriteString("Ticket Sig" + ": " + res.Signature[0:25] + "\n")
	// 			p.WriteString(Pad(13) + " " + res.Signature[25:] + "\n")
	// 		}
	// 		p.WriteString("\n" + "Control Data" + ": " + res.Date.String()[0:10] +
	// 			" " + time.Parse("04:05:06", res.TimePeriod) + "\n")
	// 		p.WriteString("Control Module ID" + ": " + res.ProductionNumber + "\n")
	// 		p.WriteString("VSC ID" + ": " + res.VSC + "\n")

	// 	}
	// }
	p.Formfeed()
	p.SetFont("A")
	p.WriteString("Signature" + ":\t " + "............." + "\n")
	p.Formfeed()
	if folio.Store.InvoiceFooter != "" {
		p.SetAlign("center")
		p.WriteString(folio.Store.InvoiceFooter + "\n")
	}
	p.Formfeed()
	p.Cut()

}
