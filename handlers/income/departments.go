package income

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/handlers"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

func ListDepartments(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	for key, val := range queryParams {
		query[key] = val
	}
	var departments []map[string]interface{}
	err := db.DB.C("departments").Find(query).All(&departments)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	handlers.ReturnJSONMessage(w, departments)
}

func GetDepartment(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["id"]
	q["id"] = id

	var department map[string]interface{}
	err := db.DB.C("departments").Find(q).One(&department)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	handlers.ReturnJSONMessage(w, department)
}
