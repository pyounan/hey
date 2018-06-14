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

//PrintRequest varibles for printing
type PrintRequest struct {
	PrinterType  string
	OrderedItems []models.EJEvent
	Items        []models.POSLineItem
	Invoice      models.Invoice
}

// MenuPrinter used to group item by menu id and hold printer too
type MenuPrinter struct {
	Printer printing.Printer
	Menu    int
}

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
func sendToPrint(req PrintRequest) {
	// printerType = folioPrinter

	var printer models.Printer
	var err error

	fmt.Printf("Printer Type %v\n", req.PrinterType)
	if req.PrinterType == kitchenPrinter {
		items := make(map[MenuPrinter][]models.EJEvent)
		// fmt.Printf("Events json:events %v\n", len(data.Invoice.Events))
		// fmt.Printf("Events json:grouped_lineitems %v\n", len(data.Invoice.GroupedLineItems))
		// fmt.Printf("Events json:posinvoicelineitem_set %v\n", len(data.Invoice.Items))
		// i := 0
		for _, e := range req.Invoice.GroupedLineItems {
			var menu models.StoreMenuItemConfig
			fmt.Println(e.Item)
			// fmt.Printf("Condition %v\n\n", (e.Quantity-e.SubmittedQuantity) > 0)
			// if (e.Quantity - e.SubmittedQuantity) > 0 {
			// 	continue
			// }
			menu, err = getMenuByItemID(*e.Item, req.Invoice.Store)
			// menu, err = getMenuByItemID(8, req.Invoice.Store)
			if err == nil {
				fmt.Printf("Menu id %v\n", menu.ID)
				fmt.Printf("Menu Menu %v\n", menu.Menu)
				printer, err = getPrinterByID(menu.AttachedAttributes.KitchenPrinter)
			} else {
				fmt.Printf("Printer Stopped with Menu Error %v\n", err)
				return
			}
			fmt.Printf("Printer for E %v \t printer %v\n\n", e.Description, printer)
			p := printing.MaptoPrinter(printer)
			menuPrinter := MenuPrinter{
				Printer: p,
				Menu:    menu.Menu,
			}
			if items[menuPrinter] != nil {
				fmt.Printf("Found Printer for E %v printer \t %v\n\n", e.Description, printer)
				events := items[menuPrinter]
				events = append(events, e)
				items[menuPrinter] = events
			} else {
				fmt.Printf("Not Found Found Printer for E %v  \t printer %v\n\n", e.Description, printer)
				items[menuPrinter] = []models.EJEvent{e}
			}
		}
		for group, event := range items {
			if group.Printer.PrinterIP != "" {
				fmt.Printf("Start printing on %v\n", group.Printer.PrinterIP)
				k := printing.KitchenPrint{}
				k.GropLineItems = event
				k.Printer = group.Printer
				if !k.Printer.IsUSB {
					k.Printer.PrinterIP = group.Printer.PrinterIP + ":9100"
				}
				k.Invoice = req.Invoice
				k.Timezone = config.Config.TimeZone
				// k.Timezone = "Africa/Cairo"
				k.Cashier, err = getCashierByNumber(req.Invoice.CashierNumber)
				if err != nil {
					fmt.Printf("Can't get casher for number %v,ERR %v\n", req.Invoice.CashierNumber, err)
					return
				}
				// defer func() {
				// 	if r := recover(); r != nil {
				// 		fmt.Printf("Recovered Kitchen Print %v\n", r)
				// 	}
				// }()
				fmt.Printf("Sent PrintKitchen %v\n", group.Printer.PrinterIP)
				for _, i := range k.GropLineItems {
					fmt.Println(i.Description)
				}
				err := printing.PrintKitchen(&k)
				if err != nil {
					fmt.Printf("Kitchen Printer err %v\n", err)
				}
			} else {
				log.Println("Printing stop no printer IP")
			}
		}

	}
	if req.PrinterType == folioPrinter {
		fmt.Printf("Items %v\n", len(req.OrderedItems))
		fmt.Printf("Printer Type %v\n", req.PrinterType)
		var printerIP string
		printer, err := getPrinterForTerminalIP(req.Invoice.TerminalID, "cashier")
		// printer, err := getPrinterForTerminalIP(2, "cashier")
		if err == nil {
			if printer.PrinterIP != nil {
				printerIP = *printer.PrinterIP
			}
		}
		fmt.Printf("Printer Folio %v\n", printer)
		if printerIP != "" {
			fmt.Printf("Start printing on %v\n", printerIP)
			f := printing.FolioPrint{}
			f.Printer = printer
			if !f.Printer.IsUSB {
				printerIP = printerIP + ":9100"
			}
			f.Items = req.OrderedItems
			f.Printer.PrinterIP = &printerIP
			f.Invoice = req.Invoice
			f.Timezone = config.Config.TimeZone
			// f.Timezone = "Africa/Cairo"
			f.Cashier, err = getCashierByNumber(req.Invoice.CashierNumber)
			if err != nil {
				fmt.Printf("Can't get casher for number %v,ERR %v\n", req.Invoice.CashierNumber, err)
				return
			}
			f.Terminal, err = getTerminalByID(req.Invoice.TerminalID)
			if err != nil {
				fmt.Printf("Can't get terminal for id %v,ERR %v\n", req.Invoice.TerminalID, err)
				return
			}
			f.Store, err = getStoreByID(req.Invoice.Store)
			if err != nil {
				fmt.Printf("Can't get store for number %v,ERR %v\n", req.Invoice.Store, err)
				return
			}
			f.Company, err = getCompany()
			if err != nil {
				fmt.Printf("Can't get Company, ERR %v\n", err)
				return
			}
			totalDiscount := 0.0
			for _, item := range req.Items {
				for _, d := range item.AppliedDiscounts {
					totalDiscount += d.Amount
				}
			}
			f.TotalDiscounts = totalDiscount
			// defer func() {
			// 	if r := recover(); r != nil {
			// 		fmt.Printf("Recovered Folio Print %v\n", r)
			// 	}
			// }()
			fmt.Printf("Send PrintFolio %v\n", printerIP)
			err := printing.PrintFolio(&f)
			if err != nil {
				fmt.Printf("Folio Printer error : %v\n", err)
			}
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
func getMenuByItemID(item int64, store int) (models.StoreMenuItemConfig, error) {
	storeMenuItemConfigs := models.StoreMenuItemConfig{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("storemenuitemconfig").With(session).Find(bson.M{"item": item, "store": store}).One(&storeMenuItemConfigs)
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
