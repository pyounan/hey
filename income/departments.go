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
		query[key] = val
	}
	departments := []map[string]interface{}{}
	err := db.DB.C("departments").Find(query).All(&departments)
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
	err := db.DB.C("departments").Find(q).One(&department)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, department)
}
