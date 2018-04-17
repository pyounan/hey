package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/models"

	"gopkg.in/mgo.v2/bson"
)

// ListPrinters swagger:route GET /api/pos/printer/ printers getPrinterList
//
// List Printers
//
// returns a list of printers
//
// Responses:
//   200: []printer
func ListPrinters(w http.ResponseWriter, r *http.Request) {
	printers := []models.Printer{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("printers").With(session).Find(bson.M{}).All(&printers)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, printers)
}

// ListPrinterSettings swagger:route GET /api/pos/printersettings/ printers listPrinterSettings
//
// List Printer Settings
//
// returns a list of printer settings
//
// Responses:
//   200: []printerSetting
func ListPrinterSettings(w http.ResponseWriter, r *http.Request) {
	settings := []models.PrinterSetting{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("printersettings").With(session).Find(bson.M{}).All(&settings)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, settings)
}
