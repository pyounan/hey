package pos

import (
	"encoding/json"
	"net/http"
	"pos-proxy/db"
	"pos-proxy/handlers"
	"strconv"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

func ListTables(w http.ResponseWriter, r *http.Request) {

}

func GetTable(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["number"]
	q["number"] = id

	var table map[string]interface{}
	err := db.DB.C("tables").Find(q).One(&table)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	handlers.ReturnJSONMessage(w, table)
}

func UpdateTable(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["id"]
	val, err := strconv.Atoi(id)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	q["id"] = val

	var table map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&table)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}

	err = db.DB.C("tables").Update(q, table)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	handlers.ReturnJSONMessage(w, table)
}
