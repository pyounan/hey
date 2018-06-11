package pos

import (
	"fmt"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/income"
	"pos-proxy/pos/models"

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
	printerType = folioPrinter
	fmt.Printf("Printer Type %v\n", printerType)
	// fmt.Printf("Events %v\n", data)
	if printerType == kitchenPrinter {
		// fmt.Printf("Printer Type %v\n", printerType)
		// fmt.Printf("Events : %v\n", data.Invoice.GroupedLineItems)
		for _, e := range data.Invoice.GroupedLineItems {
			// fmt.Printf("Events : %v\n", e)
			var printerIP string
			// fmt.Printf("Item ID %v\n", *e.Item)
			// meun, err := getMenuByItemID(*e.Item)
			meun, err := getMenuByItemID(8)
			// fmt.Printf("Menu %v\n", meun.AttachedAttributes)
			if err != nil {
				// fmt.Printf("Menu Error %v\n", err)
				smartPrinter, settingsError := getPrinterSettings()
				// fmt.Printf("Smart Printer 1 %v\n", &smartPrinter.IP)
				if settingsError != nil {
					fmt.Printf("Printing Stopped, Couldn't find smart printer IP %v, Error %v\n", e.Item, settingsError)
					fmt.Printf("Printing Stopped, Couldn't find printer for item %v, Error %v\n", e.Item, err)
					return
				} else {
					printerIP = *smartPrinter.IP
					fmt.Printf("Set Printer ip %v\n", printerIP)
				}
			} else {
				printer, err := getPrinterByID(meun.AttachedAttributes.KitchenPrinter)
				// fmt.Printf("get printer by id %v, found %v\n", meun.AttachedAttributes.KitchenPrinter, printer.PrinterID)
				if err != nil {
					smartPrinter, settingsError := getPrinterSettings()
					if settingsError != nil {
						fmt.Printf("Printing Stopped, Couldn't find smart printer IP %v, Error %v\n", e.Item, settingsError)
						fmt.Printf("Printing Stopped, Couldn't find printer for item %v, Error Printer Ip == 0\n", e.Item)
						return
					}
					printerIP = *smartPrinter.IP
					fmt.Printf("Set Printer ip %v\n", printerIP)

				} else {
					if printer.PrinterIP == nil {
						smartPrinter, settingsError := getPrinterSettings()
						if settingsError != nil {
							fmt.Printf("Printing Stopped, Couldn't find smart printer IP %v, Error %v\n", e.Item, settingsError)
							fmt.Printf("Printing Stopped, Couldn't find printer for item %v, Error Printer Ip == 0\n", e.Item)
							return
						}
						printerIP = *smartPrinter.IP
						fmt.Printf("Set Printer ip %v\n", printerIP)

					} else {
						printerIP = *printer.PrinterIP
						fmt.Printf("Set Printer ip %v\n", printerIP)
					}
				}
			}
			if printerIP != "" {
				fmt.Printf("Start printing on %v\n", printerIP)
			}
		}
	}
	if printerType == folioPrinter {
		fmt.Printf("Printer Type %v\n", printerType)
		var printerIP string
		// printer, err := getPrinterForTerminalIP(data.TerminalID, "cashier")
		printer, err := getPrinterForTerminalIP(1, "cashier")

		if err != nil {
			smartPrinter, settingsError := getPrinterSettings()
			if settingsError != nil {
				fmt.Printf("Printing Stopped, Couldn't find smart printer IP, Error %v\n", settingsError)
				fmt.Printf("Printing Stopped, Could n't get Printer for terminal %v with error = %v\n",
					data.TerminalID, err)
				return
			}
			printerIP = *smartPrinter.IP
			fmt.Printf("Set Printer ip %v\n", printerIP)
		}
		if printer.PrinterIP == nil {
			smartPrinter, smartError := getPrinterSettings()
			if smartError != nil {
				fmt.Println("Printing Stopped, Could n't get Smart Printer")
				fmt.Printf("Printing Stopped, Could n't get Printer for terminal %v with error printer IP = nil\n",
					data.TerminalID)
				return
			}
			if smartPrinter.IP == nil {
				fmt.Println("Printing Stopped, Could n't get Smart Printer IP == nil")
				fmt.Printf("Printing Stopped, Could n't get Printer for terminal %v with error printer IP = nil\n",
					data.TerminalID)
				return
			}
			printerIP = *printer.PrinterIP
			fmt.Printf("Set Printer ip %v\n", printer.PrinterIP)
		} else {
			printerIP = *printer.PrinterIP
			fmt.Printf("Set Printer ip %v\n", printer.PrinterIP)
		}
		// for _, item := range data.Invoice.Items {
		// 	getPrinterForTerminalIP(data.TerminalID, "cashier")
		// }

		fmt.Printf("Start printing on %v\n", printerIP)
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
func getCompany(number int) (income.Company, error) {
	company := income.Company{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("company").With(session).Find(nil).One(&company)
	if err != nil {
		return income.Company{}, err
	}
	return company, nil
}
