package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"

	"gopkg.in/mgo.v2/bson"
)

func ListCourses(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	for key, val := range queryParams {
		query[key] = val
	}
	var courses []map[string]interface{}
	err := db.DB.C("courses").Find(query).All(&courses)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, courses)
}
