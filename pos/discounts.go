package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"

	"gopkg.in/mgo.v2/bson"
)

func ListFixedDiscounts(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	for key, val := range queryParams {
		query[key] = val
	}
	fixedDiscounts := []map[string]interface{}{}
	err := db.DB.C("fixeddiscounts").Find(query).All(&fixedDiscounts)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, fixedDiscounts)
}
