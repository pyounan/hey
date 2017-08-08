package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"

	"gopkg.in/mgo.v2/bson"
)

func ListPrinters(w http.ResponseWriter, r *http.Request) {
	printers := []map[string]interface{}{}
	err := db.DB.C("printers").Find(bson.M{}).All(&printers)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, printers)
}

func ListPrinterSettings(w http.ResponseWriter, r *http.Request) {
	settings := []map[string]interface{}{}
	err := db.DB.C("printersettings").Find(bson.M{}).All(&settings)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, settings)
}
