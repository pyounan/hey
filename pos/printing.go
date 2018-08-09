package pos

import (
	"fmt"
	"log"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/income"
	"pos-proxy/pos/models"
	"pos-proxy/printing"
	"time"

	"gopkg.in/mgo.v2/bson"
)

const kitchenPrinter = "Kitchen"
const folioPrinter = "Folio"

//PrintRequest varibles for printing
type PrintRequest struct {
	PrinterType  string               `bson:"printer_type"`
	OrderedItems []models.EJEvent     `bson:"ordered_posinvoicelineitem_set"`
	Items        []models.POSLineItem `bson:"posinvoicelineitem_set"`
	Invoice      models.Invoice       `bson:"posinvoice"`
}

type QueuePrintRequest struct {
	ID             bson.ObjectId    `json:"id" bson:"_id,omitempty"`
	CreatedAt      time.Time        `bson:"created_at"`
	UpdatedAt      time.Time        `bson:"updated_at"`
	Status         string           `bson:"status"`
	Items          []models.EJEvent `bson:"items"`
	Invoice        models.Invoice   `bson:"invoice"`
	Terminal       models.Terminal  `bson:"termianl"`
	Store          models.Store     `bson:"store"`
	Cashier        income.Cashier   `bson:"cashier"`
	Company        income.Company   `bson:"company"`
	Printer        models.Printer   `bson:"printer"`
	TotalDiscounts float64          `bson:"total_discounts"`
	Timezone       string           `bson:"timezone"`
	GroupLineItems []models.EJEvent `bson:"group_lineitems"`
	//Kitchen or Folio
	PrintType string `bson:"print_type"`
}

//SetQueued mark print request as Queued to be print
func (q *QueuePrintRequest) SetQueued() {
	q.Status = "Queued"
}

//SetRetry mark print request as Retry
//mean that it tried to print once and failed
func (q *QueuePrintRequest) SetRetry() {
	q.Status = "Retry"
}

//SetPrinted mark print request as Printed
func (q *QueuePrintRequest) SetPrinted() {
	q.Status = "Printed"
}

//sendToPrint
//IF printer is Kitchen
//get invoice.ItemsPerPrinter
//Loop on each Printer
//Get Printer object
//If printer id == null then chage it with smartprinter ip
//Send Items to printKitchenOrder
//IF printer is Folio
//For Invoice.Items
//Get terminal Cashier Printer
//If printer id == null then chage it with smartprinter ip
//printFolio
func sendToPrint(req PrintRequest) {

	var printer models.Printer
	var err error

	if req.PrinterType == kitchenPrinter {
		for printerID, events := range req.Invoice.ItemsPerPrinter {
			printer, err = getPrinterByID(printerID)
			if err != nil {
				fmt.Printf("Printer Stopped with Printer Error %v\n", err)
				continue
			} else {
				if printer.PrinterIP == nil {
					fmt.Printf("Printer Stopped with Printer Error IP == nil")
					continue
				}

				if *printer.PrinterIP != "" {
					queueReq := QueuePrintRequest{}
					queueReq.PrintType = kitchenPrinter
					queueReq.GroupLineItems = events
					queueReq.Printer = printer
					if !queueReq.Printer.IsUSB {
						printIP := *printer.PrinterIP + ":9100"
						queueReq.Printer.PrinterIP = &printIP
					}
					queueReq.Invoice = req.Invoice
					queueReq.Timezone = config.Config.TimeZone
					// k.Timezone = "Africa/Cairo"
					queueReq.Cashier, err = getCashierByNumber(req.Invoice.CashierNumber)
					if err != nil {
						fmt.Printf("Can't get casher for number %v,ERR %v\n", req.Invoice.CashierNumber, err)
						continue
					}
					queueReq.CreatedAt = time.Now()
					queueReq.UpdatedAt = time.Now()
					queueReq.SetQueued()
					err = QueuePrint(queueReq)
					if err != nil {
						fmt.Printf("Failed to store request %v\n", err)
					}
				} else {
					log.Println("Printing stop no printer IP")
				}

			}
		}

	}
	if req.PrinterType == folioPrinter {
		var printerIP string
		printer, err := getPrinterForTerminalIP(req.Invoice.TerminalID, "cashier")
		if err == nil {
			if printer.PrinterIP != nil {
				printerIP = *printer.PrinterIP
			}
		}
		if printerIP != "" {
			queueReq := QueuePrintRequest{}
			queueReq.PrintType = folioPrinter
			queueReq.Printer = printer
			if !queueReq.Printer.IsUSB {
				printerIP = printerIP + ":9100"
			}
			queueReq.Items = req.OrderedItems
			queueReq.Printer.PrinterIP = &printerIP
			queueReq.Invoice = req.Invoice
			queueReq.Timezone = config.Config.TimeZone
			queueReq.Cashier, err = getCashierByNumber(req.Invoice.CashierNumber)
			if err != nil {
				fmt.Printf("Can't get casher for number %v,ERR %v\n", req.Invoice.CashierNumber, err)
				return
			}
			queueReq.Terminal, err = getTerminalByID(req.Invoice.TerminalID)
			if err != nil {
				fmt.Printf("Can't get terminal for id %v,ERR %v\n", req.Invoice.TerminalID, err)
				return
			}
			queueReq.Store, err = getStoreByID(req.Invoice.Store)
			if err != nil {
				fmt.Printf("Can't get store for number %v,ERR %v\n", req.Invoice.Store, err)
				return
			}
			queueReq.Company, err = getCompany()
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
			queueReq.TotalDiscounts = totalDiscount
			queueReq.CreatedAt = time.Now()
			queueReq.UpdatedAt = time.Now()
			queueReq.SetQueued()
			err := QueuePrint(queueReq)
			if err != nil {
				fmt.Printf("Failed to store request %v\n", err)
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

//QueuePrint save new print request in printer queue
func QueuePrint(req QueuePrintRequest) error {
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("printerqueue").With(session).Insert(req)
	return err
}

//GetQueuePrint try to get the latest print request with status == Queued
//If no request availabe
//then try to get the latest request with status == retry
func GetQueuePrint() (QueuePrintRequest, error) {
	session := db.Session.Copy()
	defer session.Close()
	req := QueuePrintRequest{}
	req.SetQueued()
	q := bson.M{"status": req.Status}
	err := db.DB.C("printerqueue").With(session).Find(q).Limit(1).Sort("updated_at").One(&req)
	if err != nil {
		req.SetRetry()
		q := bson.M{"status": req.Status}
		err = db.DB.C("printerqueue").With(session).Find(q).Limit(1).Sort("updated_at").One(&req)
	}
	return req, err
}
func UpdateQueuePrint(req QueuePrintRequest) error {
	session := db.Session.Copy()
	defer session.Close()
	q := bson.M{"_id": req.ID}
	err := db.DB.C("printerqueue").With(session).Update(q, req)
	return err
}

//StartPrinter get the
func StartPrinter() {
	req, err := GetQueuePrint()
	if err != nil {
		time.Sleep(1 * time.Second)
		go StartPrinter()
		return
	}
	if req.PrintType == folioPrinter {
		folioPrint := printing.FolioPrint{}
		folioPrint.Items = req.Items
		folioPrint.Invoice = req.Invoice
		folioPrint.Terminal = req.Terminal
		folioPrint.Store = req.Store
		folioPrint.Cashier = req.Cashier
		folioPrint.Company = req.Company
		folioPrint.Printer = req.Printer
		folioPrint.TotalDiscounts = req.TotalDiscounts
		folioPrint.Timezone = req.Timezone
		err := printing.PrintFolio(&folioPrint)
		if err != nil {
			req.SetRetry()
			req.UpdatedAt = time.Now()
			//Update database
			err := UpdateQueuePrint(req)
			if err != nil {
				fmt.Printf("Failed to update request %+v\n", err)
			}
			time.Sleep(1 * time.Second)
			go StartPrinter()
			return
		}
		req.SetPrinted()
		req.UpdatedAt = time.Now()
		//Update database
		updateErr := UpdateQueuePrint(req)
		if updateErr != nil {
			fmt.Printf("Failed to update request %v\n", err)
		}
	} else if req.PrintType == kitchenPrinter {
		kitchenPrint := printing.KitchenPrint{}
		kitchenPrint.GropLineItems = req.GroupLineItems
		kitchenPrint.Printer = req.Printer
		kitchenPrint.Invoice = req.Invoice
		kitchenPrint.Cashier = req.Cashier
		kitchenPrint.Timezone = req.Timezone
		err := printing.PrintKitchen(&kitchenPrint)
		if err != nil {
			req.SetRetry()
			req.UpdatedAt = time.Now()
			//Update database
			updateErr := UpdateQueuePrint(req)
			if updateErr != nil {
				fmt.Printf("Failed to update request %v\n", err)
			}
			time.Sleep(1 * time.Second)
			go StartPrinter()
			return
		}
		req.SetPrinted()
		req.UpdatedAt = time.Now()
		//Update database
		updateErr := UpdateQueuePrint(req)
		if updateErr != nil {
			fmt.Printf("Failed to update request %v\n", err)
		}
	}
	time.Sleep(1 * time.Second)
	go StartPrinter()
	return
}
