package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"

	"gopkg.in/mgo.v2/bson"
)

func ListCondiments(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	for key, val := range queryParams {
		query[key] = val
	}
	var condiments []map[string]interface{}
	err := db.DB.C("condiments").With(db.Session.Copy()).Find(query).All(&condiments)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, condiments)
}
