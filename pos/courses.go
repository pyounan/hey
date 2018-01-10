package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"strconv"

	"gopkg.in/mgo.v2/bson"
)

func ListCourses(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	idStr, ok := queryParams["id"]
	if ok {
		id, _ := strconv.Atoi(idStr[0])
		query["id"] = id
	}
	courses := []map[string]interface{}{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("courses").With(session).Find(query).Sort("id").All(&courses)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, courses)
}
