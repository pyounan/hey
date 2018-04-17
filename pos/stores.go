package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/models"
	"strconv"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

// ListStores swagger:route GET /api/pos/store/ stores getStoreList
//
// List Stores
//
// returns a list of POS stores
//
// Responses:
//   200: []store
func ListStores(w http.ResponseWriter, r *http.Request) {
	stores := []models.Store{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("stores").With(session).Find(bson.M{}).All(&stores)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, stores)
}

// GetStore swagger:route GET /api/pos/store/{id}/ stores getStore
//
// Get Store
//
// returns a POS store data based on ID
//
// Parameters:
// + name: id
//   in: path
//   required: true
//   schema:
//     type: integer
// Responses:
//   200: store
func GetStore(w http.ResponseWriter, r *http.Request) {
	store := models.Store{}
	query := bson.M{}
	idStr := mux.Vars(r)["id"]
	query["id"], _ = strconv.Atoi(idStr)
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("stores").With(session).Find(query).One(&store)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, store)

}

// GetStoreDetails swagger:route GET /api/pos/storedetails/{id}/ stores getStoreDetails
//
// Get Store Details
//
// returns a POS stores details based on ID
//
// Parameters:
// + name: id
//   in: path
//   required: true
//   schema:
//     type: integer
//
// Responses:
//   200: storeDetails
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

	var storedetails models.StoreDetails
	session := db.Session.Copy()
	defer session.Close()
	err = db.DB.C("storedetails").With(session).Find(q).One(&storedetails)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, storedetails)
}
