package pos

import (
	"encoding/json"
	"net/http"
	"pos-proxy/db"
	"pos-proxy/pos/models"
	"pos-proxy/helpers"
	"pos-proxy/syncer"
	"strconv"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

func ListTables(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}

	tables := []models.Table{}
	err := db.DB.C("tables").Find(q).All(&tables)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, tables)
}

func GetTable(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	strID, _ := vars["number"]
	id, err := strconv.Atoi(strID)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	q["number"] = id

	table := models.Table{}
	err = db.DB.C("tables").Find(q).One(&table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, table)
}

func UpdateTable(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["id"]
	val, err := strconv.Atoi(id)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	q["id"] = val

	table := models.Table{}
	err = json.NewDecoder(r.Body).Decode(&table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	err = db.DB.C("tables").Update(q, table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, table)

	helpers.ReturnSuccessMessage(w, table)
}

func GetTableLatestChanges(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["id"]
	val, err := strconv.Atoi(id)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	q["table"] = val
	terminalID, err := strconv.Atoi(r.URL.Query().Get("terminal_id"))
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	table := models.Table{}
	err = db.DB.C("tables").Find(bson.M{"id": q["table"]}).One(&table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	// Fix
	q["is_settled"] = false
	invoices := []models.Invoice{}
	err = db.DB.C("posinvoices").Find(q).All(&invoices)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	terminal := make(map[string]interface{})
	err = db.DB.C("terminals").Find(bson.M{"id": terminalID}).One(&terminal)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	resp := bson.M{}
	resp["posinvoices"] = invoices
	resp["table"] = table
	resp["terminal"] = terminal["number"]
	resp["lockedposinvoices"] = false
	helpers.ReturnSuccessMessage(w, resp)
}
