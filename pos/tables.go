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

// ListTables swagger:route GET /api/pos/tables/ tables getTableList
//
// List Tables
//
// returns a list of POS tables
//
// Responses:
//   200: []table
func ListTables(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}

	storeStr := r.URL.Query().Get("store")
	if storeStr != "" {
		storeID, _ := strconv.Atoi(storeStr)
		q["store_id"] = storeID
	}

	tables := []models.Table{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("tables").With(session).Find(q).All(&tables)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, tables)
}

// GetTable swagger:route GET /api/pos/tables/{id}/ tables getTable
//
// Get Table
//
// returns a details of a table by id
//
// Parameters:
// + name: id
//   in: path
//   required: true
//   schema:
//      type: integer
// Responses:
//   200: table
func GetTable(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	strID, _ := vars["id"]
	id, err := strconv.Atoi(strID)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	q["id"] = id

	table := models.Table{}
	session := db.Session.Copy()
	defer session.Close()
	err = db.DB.C("tables").With(session).Find(q).One(&table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, table)
}

// GetTableByNumber swagger:route GET /api/pos/tables/{number}/ tables getTableByNumber
//
// Get Table By Number
//
// returns a details of a table by table number
//
// Parameters:
// + name: number
//   in: path
//   required: true
//   schema:
//      type: integer
//
// Responses:
//   200: table
func GetTableByNumber(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	numStr, _ := vars["number"]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	q["number"] = num

	table := models.Table{}
	session := db.Session.Copy()
	defer session.Close()
	err = db.DB.C("tables").With(session).Find(q).One(&table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, table)
}

// UpdateTable swagger:route PUT /api/pos/tables/{id}/ tables updateTable
//
// Update Table
//
// updates table data by id
//
// Parameters:
// + name: id
//   in: path
//   required: true
//   schema:
//      type: integer
//
// + name: body
//   in: body
//   type: table
//   required: true
//
// Responses:
//   200: table
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

	session := db.Session.Copy()
	defer session.Close()
	err = db.DB.C("tables").With(session).Update(q, table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, table)

	helpers.ReturnSuccessMessage(w, table)
}

// GetTableLatestChangesResponse swagger:model getTableLatestChangesResponse
// defines the body for the response of GetTableLatestChanges request
type GetTableLatestChangesResponse struct {
	Invoices []models.Invoice `json:"posinvoices" bson:"posinvoices"`
	// flag that shows if the invoices on these table are currently locked or not
	LockedInvoices bool         `json:"lockedposinvoices" bson:"lockedposinvoices"`
	Table          models.Table `json:"table" bson:"table"`
	// ID of the terminal that is currently locking these invoices
	Terminal int `json:"terminal" bson:"terminal"`
}

// GetTableLatestChanges swagger:route POST /api/pos/tables/{id}/latestchanges/ tables listInvoicesOnTables
//
// List Unsettled Invoices on Table (getLatestChanges)
//
// returns a list of unsettled invoices (with full details) on a table at the moment.
//
// Parameters:
// + name: id
//   in: path
//   required: true
//   schema:
//      type: integer
// Responses:
//   200: getTableLatestChangesResponse
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

	session := db.Session.Copy()
	defer session.Close()
	table := models.Table{}
	err = db.DB.C("tables").With(session).Find(bson.M{"id": q["table_number"]}).One(&table)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	q["is_settled"] = false
	invoices := []models.Invoice{}
	err = db.DB.C("posinvoices").With(session).Find(q).All(&invoices)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	terminal := models.Terminal{}
	err = db.DB.C("terminals").With(session).Find(bson.M{"id": terminalID}).One(&terminal)
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

	resp := GetTableLatestChangesResponse{
		Invoices:       invoices,
		Table:          table,
		Terminal:       otherTerminal,
		LockedInvoices: invoicesLocked,
	}
	helpers.ReturnSuccessMessage(w, resp)
}
