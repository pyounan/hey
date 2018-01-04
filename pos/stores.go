package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"strconv"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

func ListStores(w http.ResponseWriter, r *http.Request) {
	stores := []map[string]interface{}{}
	err := db.DB.C("stores").With(db.Session.Copy()).Find(bson.M{}).All(&stores)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, stores)
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
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	q["id"] = val

	var storedetails map[string]interface{}
	err = db.DB.C("storedetails").With(db.Session.Copy()).Find(q).One(&storedetails)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, storedetails)
}
