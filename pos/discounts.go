package pos

import (
	"encoding/json"
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

// GetFixedDiscount returns an object of a FixedDiscount based on ID
func GetFixedDiscount(w http.ResponseWriter, r *http.Request) {
	idStr, _ := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	d := make(map[string]interface{})
	err := db.DB.C("fixeddiscounts").Find(bson.M{"id": id}).One(&d)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, d)
}

// CreateFixedDiscount creates a new object of a FixedDiscount
func CreateFixedDiscount(w http.ResponseWriter, r *http.Request) {
	d := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

	err = db.DB.C("fixeddiscounts").Insert(d)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, d)

	helpers.ReturnSuccessMessage(w, d)
}

// UpdateFixedDiscount updated a FixedDiscount object based on ID
func UpdateFixedDiscount(w http.ResponseWriter, r *http.Request) {
	idStr, _ := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	d := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()

	if _, ok := d["_id"]; ok {
		delete(d, "_id")
	}

	err = db.DB.C("fixeddiscounts").Update(bson.M{"id": id}, d)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, d)

	helpers.ReturnSuccessMessage(w, d)
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
