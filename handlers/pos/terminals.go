package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/handlers"
	"strconv"

	"gopkg.in/mgo.v2/bson"
)

func ListTerminals(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	for key, val := range queryParams {
		if key == "store" {
			num, err := strconv.Atoi(val[0])
			if err != nil {
				handlers.ReturnJSONError(w, err.Error())
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
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	handlers.ReturnJSONMessage(w, terminals)
}
