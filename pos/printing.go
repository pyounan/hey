package pos

import (
	"fmt"
	"log"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/income"
	"pos-proxy/pos/models"
	"pos-proxy/printing"

	"gopkg.in/mgo.v2/bson"
)

const kitchenPrinter = "Kitchen"
const folioPrinter = "Folio"

//sendToPrint
//IF printer is Kitchen
//get invoice.Events
//Group items by storeMenuItemConfig ID
//Loop on each Group
//Get Printer object
//If printer id == null then chage it with smartprinter ip
//Send Item to printKitchenOrder
//IF printer is Folio
//For Invoice.Items
//Get terminal Cashier Printer
//If printer id == null then chage it with smartprinter ip
//printFolio
func sendToPrint(printerType string, data models.InvoicePOSTRequest) {
	// printerType = folioPrinter
	var printer models.Printer
	var err error
	// fmt.Printf("Printer Type %v\n", printerType)
	if printerType == kitchenPrinter {
		for _, e := range data.Invoice.GroupedLineItems {
			var printerIP string
			var menu models.StoreMenuItemConfig
			fmt.Println(e.Item)
			menu, err = getMenuByItemID(*e.Item)
			// menu, err = getMenuByItemID(8)
			if err == nil {
				printer, err = getPrinterByID(menu.AttachedAttributes.KitchenPrinter)
				if err == nil {
					if printer.PrinterIP != nil {
						printerIP = *printer.PrinterIP
					}
				}
			}
			// smartPrinter, smartErr := getPrinterSettings()
			// if smartErr != nil {
			// 	if printerIP == "" {
			// 		fmt.Printf("Printing Stopped, Couldn't find smart printer IP %v, Error %v\n", e.Item, smartErr)
			// 		fmt.Printf("Printing Stopped, Couldn't find printer for item %v, Error %v\n", e.Item, err)
			// 		return
			// 	}
			// } else {
			// 	if printerIP == "" {
			// 		printerIP = *smartPrinter.IP
			// 		fmt.Printf("Set Printer ip %v\n", printerIP)
			// 	}
			// }
			if printerIP != "" {
				fmt.Printf("Start printing on %v\n", printerIP)
				k := printing.KitchenPrint{}
				k.Printer = printer
				printerIP = printerIP + ":9100"
				k.Printer.PrinterIP = &printerIP
				k.Invoice = data.Invoice
				k.Timezone = config.Config.TimeZone
				// k.Timezone = "Africa/Cairo"
				k.Cashier, err = getCashierByNumber(data.CashierNumber)
				if err != nil {
					fmt.Printf("Can't get casher for number %v,ERR %v\n", data.CashierNumber, err)
					return
				}
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("Recovered Kitchen Print %v\n", r)
					}
				}()
				fmt.Printf("Sent PrintKitchen %v\n", printerIP)
				printing.PrintKitchen(&k)
			} else {
				log.Println("Printing stop no printer IP")
			}
		}
	}
	if printerType == folioPrinter {
		var printerIP string
		printer, err := getPrinterForTerminalIP(data.TerminalID, "cashier")
		// printer, err := getPrinterForTerminalIP(1, "cashier")
		if err == nil {
			if printer.PrinterIP != nil {
				printerIP = *printer.PrinterIP
			}
		}
		// smartPrinter, smartErr := getPrinterSettings()
		// if smartErr != nil {
		// 	if printerIP == "" {
		// 		fmt.Println("Printing Stopped, Couldn't find smart printer IP")
		// 		fmt.Printf("Printing Stopped, Couldn't find printer for terminal %v, Error %v\n", data.TerminalID, err)
		// 		return
		// 	}
		// } else {
		// 	if printerIP == "" {
		// 		printerIP = *smartPrinter.IP
		// 		fmt.Printf("Set Printer ip %v\n", printerIP)
		// 	}
		// }
		if printerIP != "" {
			fmt.Printf("Start printing on %v\n", printerIP)
			f := printing.FolioPrint{}
			f.Printer = printer
			printerIP = printerIP + ":9100"
			f.Printer.PrinterIP = &printerIP
			f.Invoice = data.Invoice
			f.Timezone = config.Config.TimeZone
			// f.Timezone = "Africa/Cairo"
			f.Cashier, err = getCashierByNumber(data.CashierNumber)
			if err != nil {
				fmt.Printf("Can't get casher for number %v,ERR %v\n", data.CashierNumber, err)
				return
			}
			f.Terminal, err = getTerminalByID(data.TerminalID)
			if err != nil {
				fmt.Printf("Can't get terminal for id %v,ERR %v\n", data.TerminalID, err)
				return
			}
			f.Store, err = getStoreByID(data.Invoice.Store)
			if err != nil {
				fmt.Printf("Can't get store for number %v,ERR %v\n", data.CashierNumber, err)
				return
			}
			f.Company, err = getCompany()
			if err != nil {
				fmt.Printf("Can't get store for number %v,ERR %v\n", data.CashierNumber, err)
				return
			}
			totalDiscount := 0.0
			for _, item := range data.Invoice.Items {
				for _, d := range item.AppliedDiscounts {
					totalDiscount += d.Amount
				}
			}
			f.TotalDiscounts = totalDiscount
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Recovered Folio Print %v\n", r)
				}
			}()
			fmt.Printf("Send PrintFolio %v\n", printerIP)
			printing.PrintFolio(&f)
		} else {
			log.Println("Printing stop no printer IP")
		}

	}
}
func checkProxyPrintingEnabled() bool {
	return config.Config.ProxyPrintingEnabled
}
func getPrinterForTerminalIP(terminal int, printerType string) (models.Printer, error) {
	printer := models.Printer{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("printers").With(session).Find(bson.M{"terminal": terminal, "printer_type": printerType}).One(&printer)
	// err := db.DB.C("printers").With(session).Find(bson.M{}).All(&printer)
	if err != nil {
		return models.Printer{}, err
	}
	return printer, nil
}
func getPrinterByID(id int) (models.Printer, error) {
	printer := models.Printer{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("printers").With(session).Find(bson.M{"id": id}).One(&printer)
	// err := db.DB.C("printers").With(session).Find(bson.M{}).All(&printer)
	if err != nil {
		return models.Printer{}, err
	}
	return printer, nil
}

func getstoreMenuItemConfigs() ([]models.StoreMenuItemConfig, error) {
	storeMenuItemConfigs := []models.StoreMenuItemConfig{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("storemenuitemconfig").With(session).Find(nil).All(&storeMenuItemConfigs)
	if err != nil {
		return nil, err
	}
	return storeMenuItemConfigs, nil
}
func getMenuByItemID(item int64) (models.StoreMenuItemConfig, error) {
	storeMenuItemConfigs := models.StoreMenuItemConfig{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("storemenuitemconfig").With(session).Find(bson.M{"item": item}).One(&storeMenuItemConfigs)
	if err != nil {
		return models.StoreMenuItemConfig{}, err
	}
	return storeMenuItemConfigs, nil
}

func getMenuByMenuID(menu *int64) (models.StoreMenuItemConfig, error) {
	storeMenuItemConfigs := models.StoreMenuItemConfig{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("storemenuitemconfig").With(session).Find(bson.M{"menu": menu}).One(&storeMenuItemConfigs)
	if err != nil {
		return models.StoreMenuItemConfig{}, err
	}
	return storeMenuItemConfigs, nil
}
func getPrinterSettings() (models.PrinterSetting, error) {
	settings := models.PrinterSetting{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("printersettings").With(session).Find(nil).One(&settings)
	if err != nil {
		return models.PrinterSetting{}, err
	}
	return settings, nil
}

func getTerminalByID(id int) (models.Terminal, error) {
	terminal := models.Terminal{}
	session := db.Session.Copy()
	defer session.Close()
	q := bson.M{"id": id}
	err := db.DB.C("terminals").With(session).Find(q).One(&terminal)
	if err != nil {
		return models.Terminal{}, err
	}
	return terminal, nil
}
func getStoreByID(id int) (models.Store, error) {
	store := models.Store{}
	session := db.Session.Copy()
	defer session.Close()
	q := bson.M{"id": id}
	err := db.DB.C("stores").With(session).Find(q).One(&store)
	if err != nil {
		return models.Store{}, err
	}
	return store, nil
}
func getCashierByNumber(number int) (income.Cashier, error) {
	cashier := income.Cashier{}
	session := db.Session.Copy()
	defer session.Close()
	q := bson.M{"number": number}
	err := db.DB.C("cashiers").With(session).Find(q).One(&cashier)
	if err != nil {
		return income.Cashier{}, err
	}
	return cashier, nil
}
func getCompany() (income.Company, error) {
	company := income.Company{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("company").With(session).Find(nil).One(&company)
	if err != nil {
		return income.Company{}, err
	}
	return company, nil
}
