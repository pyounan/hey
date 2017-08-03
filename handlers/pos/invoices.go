package pos

import (
	"encoding/json"
	"net/http"
	"pos-proxy/db"
	"pos-proxy/handlers"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

func ListInvoices(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	for key, val := range r.URL.Query() {
		if key == "store" {
			num, err := strconv.Atoi(val[0])
			if err != nil {
				handlers.ReturnJSONError(w, err.Error())
				return
			}
			q[key] = num
		} else if key == "updated_on" {
			t, err := time.Parse("", val[0])
			if err != nil {
				handlers.ReturnJSONError(w, err.Error())
				return
			}
			q[key] = bson.M{"$gt": t}
		} else {
			q[key] = val
		}
	}
	var invoices []map[string]interface{}
	err := db.DB.C("posinvoices").Find(q).All(&invoices)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	handlers.ReturnJSONMessage(w, invoices)
}

func GetInvoice(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["invoice_nubmer"]
	q["invoice_number"] = id

	var invoice map[string]interface{}
	err := db.DB.C("posinvoices").Find(q).One(&invoice)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	handlers.ReturnJSONMessage(w, invoice)
}

type InvoicePOSTRequest struct {
	Invoice            map[string]interface{}   `json:"invoice"`
	LineItems          []map[string]interface{} `json:"lineitems"`
	Events             []map[string]interface{} `json:"events"`
	PostingInformation []map[string]interface{} `json:"posting_information,omitempty"`
	RCRS               string                   `json:"rcrs,omitempty"`
}

// CreateOrUpdateInvoice
func CreateOrUpdateInvoice(w http.ResponseWriter, r *http.Request) {
	var req InvoicePOSTRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	defer r.Body.Close()

	fdmResp, err := submitToFDM(req)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}

	q := bson.M{"$set": req.Invoice}
	_, err = db.DB.C("posinvoices").Upsert(bson.M{"invoice_number": req.Invoice["invoice_number"]}, q)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
}

func UpdateInvoice(w http.ResponseWriter, r *http.Request) {

}

func LockInvoice(w http.ResponseWriter, r *http.Request) {

}

func UnlockInvoice(w http.ResponseWriter, r *http.Request) {

}

func FolioInvoice(w http.ResponseWriter, r *http.Request) {

}

func PayInvoice(w http.ResponseWriter, r *http.Request) {

}

func RefundInvoice(w http.ResponseWriter, r *http.Request) {

}

func Houseuse(w http.ResponseWriter, r *http.Request) {

}
