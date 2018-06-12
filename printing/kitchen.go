package printing

import (
    "fmt"
    "strings"
    "time"

    "github.com/cloudinn/escpos"
    "github.com/cloudinn/escpos/connection"
)

//PrintKitchen to print kitchen recepit
func PrintKitchen(kitchen *KitchenPrint) {

    printingParams := make(map[int]map[string]int)

    printingParams[80] = make(map[string]int)
    printingParams[80]["width"] = 800
    printingParams[80]["char_per_line"] = 28
    printingParams[80]["item_width"] = 18
    printingParams[80]["store_unit"] = 5
    printingParams[80]["qty"] = 8

    printingParams[76] = make(map[string]int)
    printingParams[76]["width"] = 760
    printingParams[76]["char_per_line"] = 33
    printingParams[76]["item_width"] = 26
    printingParams[76]["store_unit"] = 5
    printingParams[76]["qty"] = 8

    var p *escpos.Printer

    if kitchen.Printer.IsUSB {
        p = connection.NewConnection("usb", *kitchen.Printer.PrinterIP)
    } else {

        p = connection.NewConnection("network", *kitchen.Printer.PrinterIP)
    }

    p.SetAlign("left")
    if kitchen.Printer.PaperWidth == 76 {
        p.SetFont("A")
        p.SetFontSize(1, 2)
    } else {
        p.SetFont("B")
        p.SetFontSize(2, 2)
    }
    p.WriteString("Printer ID: " + kitchen.Printer.PrinterID + "\n")
    p.WriteString(strings.Repeat("=", printingParams[kitchen.Printer.PaperWidth]["char_per_line"]) + "\n")
    p.SetAlign("center")
    p.WriteString("Invoice Number" + ": " + kitchen.Invoice.InvoiceNumber + "\n")
    p.WriteString("Covers" + ": " + fmt.Sprintf("%d", kitchen.Invoice.Pax) + "\n")
    if kitchen.Invoice.TableID != nil {
        p.WriteString("Table" + ": " + *kitchen.Invoice.TableDetails + "\n")
    } else {
        p.WriteString("Takeout" + "\n")
    }
    guestName := ""
    if kitchen.Invoice.WalkinName != "" {
        guestName = kitchen.Invoice.WalkinName
    } else if kitchen.Invoice.ProfileDetails != "" {
        guestName = kitchen.Invoice.ProfileDetails
    } else if kitchen.Invoice.RoomDetails != nil {
        guestName = *kitchen.Invoice.RoomDetails
    }
    if guestName != "" {
        p.WriteString("Guest name" + ": " + guestName + "\n")
    }
    p.WriteString(fmt.Sprintf("%d", kitchen.Cashier.Number) + "/" + kitchen.Cashier.Name + "\n")
    loc, _ := time.LoadLocation(kitchen.Timezone)
    submittedOn := time.Now().In(loc)
    date := submittedOn.Format(time.RFC1123)
    p.SetFont("A")
    p.SetFontSize(1, 1)
    p.WriteString(date + "\n")

    p.WriteString(strings.Repeat("=", printingParams[kitchen.Printer.PaperWidth]["char_per_line"]) + "\n")
    p.SetAlign("left")
    p.SetReverse(1)
    p.SetEmphasize(1)

    if kitchen.Printer.PaperWidth == 76 {
        p.SetFont("A")
        p.SetFontSize(1, 2)
    } else {
        p.SetFont("B")
        p.SetFontSize(2, 2)
    }
    item := "Item"
    qty := "Qty"
    storeUnit := "Unit"
    p.WriteString(item + Pad(printingParams[kitchen.Printer.PaperWidth]["item_width"]-
        len(item)) + qty + Pad(printingParams[kitchen.Printer.PaperWidth]["qty"]-len(qty)) +
        storeUnit + Pad(printingParams[kitchen.Printer.PaperWidth]["store_unit"]-
        len(storeUnit)) + "\n")
    p.SetReverse(0)
    p.SetEmphasize(0)
    if kitchen.Printer.PaperWidth == 76 {
        p.SetFont("A")
        p.SetFontSize(1, 2)
    } else {
        p.SetFont("B")
        p.SetFontSize(2, 2)
    }
    p.WriteString(strings.Repeat("=", printingParams[kitchen.Printer.PaperWidth]["char_per_line"]) + "\n")

    for _, item := range kitchen.Invoice.Items {
        desc := item.Description
        p.WriteString(desc + Pad(printingParams[kitchen.Printer.PaperWidth]["item_width"]-
            len(desc)) + fmt.Sprintf("%.2f", item.Quantity) +
            Pad(printingParams[kitchen.Printer.PaperWidth]["qty"]-
                len(fmt.Sprintf("%.2f", item.Quantity))) + item.BaseUnit + "\n")
        for _, condiment := range item.CondimentLineItems {
            p.WriteString(condiment.Description + "\n")
        }
        if item.CondimentsComment != "" {
            p.WriteString(item.CondimentsComment + "\n")
        }
        if item.LastChildInCourse {
            p.WriteString(strings.Repeat("-", printingParams[kitchen.Printer.PaperWidth]["char_per_line"]) + "\n")
        }
    }
    p.WriteString(strings.Repeat("=", printingParams[kitchen.Printer.PaperWidth]["char_per_line"]) + "\n")
    p.Formfeed()
    p.SetAlign("center")
    text := strings.ToUpper("This is not a") + "\n" + strings.ToUpper("valid tax invoice")
    p.WriteString(text + "\n")
    p.Formfeed()
    p.Cut()

}
