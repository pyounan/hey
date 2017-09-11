package pos

import (
	"encoding/json"
	"log"
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/locks"
	"pos-proxy/pos/models"
	"pos-proxy/syncer"
	"strconv"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

func ListTables(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}

	storeStr := r.URL.Query().Get("store")
	if storeStr != "" {
		storeID, _ := strconv.Atoi(storeStr)
		q["store_id"] = storeID
	}

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
	q["table_number"] = val
	terminalID, err := strconv.Atoi(r.URL.Query().Get("terminal_id"))
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	table := models.Table{}
	err = db.DB.C("tables").Find(bson.M{"id": q["table_number"]}).One(&table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	q["is_settled"] = false
	invoices := []models.Invoice{}
	err = db.DB.C("posinvoices").Find(q).All(&invoices)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	terminal := models.Terminal{}
	err = db.DB.C("terminals").Find(bson.M{"id": terminalID}).One(&terminal)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	invoicesLocked := false
	otherTerminal, err := locks.LockInvoices(invoices, terminal.ID)
	if err != nil {
		log.Println(err.Error())
		invoicesLocked = true
	}

	resp := bson.M{}
	resp["posinvoices"] = invoices
	resp["table"] = table
	resp["terminal"] = otherTerminal
	resp["lockedposinvoices"] = invoicesLocked
	helpers.ReturnSuccessMessage(w, resp)
}
