package income

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

func GetCurrency(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["id"]
	q["id"] = id

	c := make(map[string]interface{})
	err := db.DB.C("currencies").Find(q).One(&c)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, c)
}

func ListCurrencies(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	for key, val := range queryParams {
		if key == "department" {
			query[key] = true
		} else {
			query[key] = val
		}
	}
	currencies := []map[string]interface{}{}
	err := db.DB.C("currencies").Find(query).All(&currencies)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, currencies)
}
