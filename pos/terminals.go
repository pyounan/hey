package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/pos/locks"
	"pos-proxy/helpers"
	"strconv"

	"gopkg.in/mgo.v2/bson"
	"github.com/gorilla/mux"
)

func ListTerminals(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	for key, val := range queryParams {
		if key == "store" {
			num, err := strconv.Atoi(val[0])
			if err != nil {
				helpers.ReturnErrorMessage(w, err.Error())
				return
			}
			query[key] = num
		} else {
			query[key] = val
		}
	}
	var terminals []map[string]interface{}
	err := db.DB.C("terminals").Find(query).All(&terminals)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, terminals)
}

func UnlockTerminal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, _ := vars["id"]
	locks.UnlockTerminal(idStr)
	helpers.ReturnSuccessMessage(w, true)
}
