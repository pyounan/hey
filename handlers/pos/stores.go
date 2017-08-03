package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/handlers"
	"strconv"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

func ListStores(w http.ResponseWriter, r *http.Request) {
	var stores []map[string]interface{}
	err := db.DB.C("stores").Find(bson.M{}).All(&stores)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	handlers.ReturnJSONMessage(w, stores)
}

func GetStore(w http.ResponseWriter, r *http.Request) {

}

func UpdateStore(w http.ResponseWriter, r *http.Request) {

}

func GetStoreDetails(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["id"]
	val, err := strconv.Atoi(id)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	q["id"] = val

	var storedetails map[string]interface{}
	err = db.DB.C("storedetails").Find(q).One(&storedetails)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	handlers.ReturnJSONMessage(w, storedetails)
}
