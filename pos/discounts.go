package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"strconv"

	"gopkg.in/mgo.v2/bson"
)

// ListFixedDiscounts returns a list of fixed discounts for a store
func ListFixedDiscounts(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	urlQuery := r.URL.Query()
	if _, ok := urlQuery["store"]; ok {
		id, _ := strconv.Atoi(urlQuery["store"][0])
		query["stores"] = id
	}
	if _, ok := urlQuery["cashier"]; ok {
		id, _ := strconv.Atoi(urlQuery["cashier"][0])
		query["cashiers"] = id
	}

	fixedDiscounts := []map[string]interface{}{}
	err := db.DB.C("fixeddiscounts").Find(query).Sort("id").All(&fixedDiscounts)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, fixedDiscounts)
}
