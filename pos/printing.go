package pos

import (
	"fmt"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/pos/models"

	"gopkg.in/mgo.v2/bson"
)

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
	printer, err := getPrinterForTerminalIP(data.TerminalID)
	if err != nil {
		fmt.Printf("Printing Stopped, Could n't get Printer for terminal %v with error = %v\n",
			data.TerminalID, err)
		return
	}
	if printer.PrinterIP == nil {
		_, err := getPrinterSettings()
		fmt.Printf("Printing Stopped, Could n't get Printer for terminal %v with error = %v\n",
			data.TerminalID, err)
		return
	}

	if printerType == "Folio" {

	} else if printerType == "Kitchen" {

	}

}
func checkProxyPrintingEnabled() bool {
	return config.Config.ProxyPrintingEnabled
}
func getPrinterForTerminalIP(terminal int) (models.Printer, error) {
	printer := models.Printer{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("printers").With(session).Find(bson.M{"terminal": terminal}).One(printer)
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
	err := db.DB.C("storemenuitemconfig").With(session).Find(bson.M{}).All(&storeMenuItemConfigs)
	if err != nil {
		return nil, err
	}
	return storeMenuItemConfigs, nil
}

func getPrinterSettings() ([]models.PrinterSetting, error) {
	settings := []models.PrinterSetting{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("printersettings").With(session).Find(bson.M{}).All(&settings)
	if err != nil {
		return nil, err
	}
	return settings, nil
}
