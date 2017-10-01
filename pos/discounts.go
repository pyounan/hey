package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/syncer"
	"strconv"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

// ListFixedDiscounts returns a list of fixed discounts for a store
func ListFixedDiscounts(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	urlQuery := r.URL.Query()
	if _, ok := urlQuery["id"]; ok {
		id, _ := strconv.Atoi(urlQuery["id"][0])
		query["id"] = id
	}
	if _, ok := urlQuery["store"]; ok {
		id, _ := strconv.Atoi(urlQuery["store"][0])
		query["stores"] = id
	}
	if _, ok := urlQuery["poscashier_id"]; ok {
		id, _ := strconv.Atoi(urlQuery["poscashier_id"][0])
		query["cashiers"] = id
	}

	fixedDiscounts := []map[string]interface{}{}
	err := db.DB.C("fixeddiscounts").Find(query).Sort("id").All(&fixedDiscounts)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, fixedDiscounts)
}

// DeleteFixedDiscount deletes a fixed discount by id from mongodb
// then proxy the request to the backend
func DeleteFixedDiscount(w http.ResponseWriter, r *http.Request) {
	idStr, _ := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	err := db.DB.C("fixeddiscounts").Remove(bson.M{"id": id})
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, nil)
	helpers.ReturnSuccessMessage(w, true)
}
