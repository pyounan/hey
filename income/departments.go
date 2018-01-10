package income

import (
	"net/http"
	"pos-proxy/db"

	"github.com/gorilla/mux"

	"pos-proxy/helpers"

	"gopkg.in/mgo.v2/bson"
)

func ListDepartments(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	for key, val := range queryParams {
		if key == "pos_payment" {
			query[key] = true
		} else if key == "type" && (val[0] == "waste" || val[0] == "revenue") {
			query[key] = "debit"
		} else {
			query[key] = val[0]
		}
	}
	departments := []map[string]interface{}{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("departments").With(session).Find(query).All(&departments)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, departments)
}

func GetDepartment(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["id"]
	q["id"] = id

	department := make(map[string]interface{})
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("departments").With(session).Find(q).One(&department)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, department)
}
