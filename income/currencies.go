package income

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/income/models"
	"strconv"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

// GetCurrency swagger:route GET /income/api/currency/{id}/ income getCurrency
//
// Get Currency
//
// returns a details of a currency by id
//
// Paramters:
// + name: id
//   in: path
//   required: true
//   schema:
//      type: integer
// Responses:
//   200: currency
func GetCurrency(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["id"]
	q["id"], _ = strconv.Atoi(id)

	c := models.Currency{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("currencies").With(session).Find(q).One(&c)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, c)
}

// ListCurrencies swagger:route GET /income/api/currency/ income getCurrencyList
//
// List Currencies
//
// returns a list with available currencies
//
// Responses:
//   200: []currency
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
	currencies := []models.Currency{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("currencies").With(session).Find(query).All(&currencies)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, currencies)
}
