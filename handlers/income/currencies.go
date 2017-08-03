package income

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/handlers"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

func GetCurrency(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["id"]
	q["id"] = id

	var c map[string]interface{}
	err := db.DB.C("currencies").Find(q).One(&c)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	handlers.ReturnJSONMessage(w, c)
}

func ListCurrencies(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	for key, val := range queryParams {
		query[key] = val
	}
	var currencies []map[string]interface{}
	err := db.DB.C("currencies").Find(query).All(&currencies)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	handlers.ReturnJSONMessage(w, currencies)
}
