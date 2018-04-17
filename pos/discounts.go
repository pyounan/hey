package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/models"
	"pos-proxy/proxy"
	"pos-proxy/syncer"
	"strconv"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

// ListFixedDiscounts swagger:route GET /api/pos/fixeddiscount/ discounts listFixedDiscoints
//
// List Fixed Discounts
//
// returns a list of fixed discounts for a store
//
// Parameters:
// + name: id
//   description: filter discounts by id
//   schema:
//     type: integer
//
// + name: store
//   description: filter discounts by store
//   schema:
//     type: integer
//
// + name: poscashier_id
//   description: filter discounts by cashier id
//   schema:
//     type: integer
//
// Responses:
// 200: []fixedDiscount
func ListFixedDiscounts(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	urlQuery := r.URL.Query()
	if _, ok := urlQuery["id"]; ok {
		id, _ := strconv.Atoi(urlQuery["id"][0])
		query["id"] = id
	}
	if _, ok := urlQuery["store"]; ok {
		id, _ := strconv.Atoi(urlQuery["store"][0])
		query["stores"] = id
	}
	if _, ok := urlQuery["poscashier_id"]; ok {
		id, _ := strconv.Atoi(urlQuery["poscashier_id"][0])
		query["cashiers"] = id
	}

	fixedDiscounts := []models.FixedDiscount{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("fixeddiscounts").With(session).Find(query).Sort("id").All(&fixedDiscounts)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, fixedDiscounts)
}

// GetFixedDiscount swagger:route GET /api/pos/fixeddiscount/{id}/ discounts getFixedDiscount
//
// Get Fixed Discount
//
// returns an object of a FixedDiscount based on ID
//
// Parameters:
// + name: id
//   in: path
//   required: true
//   schema:
//     type: integer
//
// Responses:
// 200: fixedDiscount
func GetFixedDiscount(w http.ResponseWriter, r *http.Request) {
	idStr, _ := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	d := models.FixedDiscount{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("fixeddiscounts").With(session).Find(bson.M{"id": id}).One(&d)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, d)
}

// CreateFixedDiscount swagger:route POST /api/pos/fixeddiscount/ discounts createFixedDiscount
//
// Create Fixed Discount
//
// creates a new fixed discount
//
// Parameters:
// + name: body
//   in: body
//   type: fixedDiscount
//   required: true
//
// Responses:
//   200: fixedDiscount
func CreateFixedDiscount(w http.ResponseWriter, r *http.Request) {
	proxy.ProxyToBackend(w, r)
}

// UpdateFixedDiscount swagger:route PUT /api/pos/fixeddiscount/{id}/ discounts updateFixedDiscount
//
// Update Fixed Discount
//
// updates a fixed discount based on ID
//
// Parameters:
// + name: id
//   in: path
//   required: true
//   schema:
//      type: integer
//
// + name: body
//   in: body
//   type: fixedDiscount
//   required: true
//
// Responses:
//   200: fixedDiscount
func UpdateFixedDiscount(w http.ResponseWriter, r *http.Request) {
	proxy.ProxyToBackend(w, r)
}

// DeleteFixedDiscount swagger:route DELETE /api/pos/fixeddiscount/{id}/ discounts deleteFixedDiscounts
//
// Delete Fixed Discount
//
// deletes a fixed discount by id from mongodb
// then proxy the request to the backend
//
// Parameters:
// + name: id
//   in: path
//   required: true
//   schema:
//     type: integer
func DeleteFixedDiscount(w http.ResponseWriter, r *http.Request) {
	idStr, _ := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("fixeddiscounts").With(session).Remove(bson.M{"id": id})
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, nil)
	helpers.ReturnSuccessMessage(w, true)
}
