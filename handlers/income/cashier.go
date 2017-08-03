package income

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/handlers"

	"gopkg.in/mgo.v2/bson"
)

func GetPosCashier(w http.ResponseWriter, req *http.Request) {
	var cashier map[string]interface{}
	q := bson.M{}
	for key, val := range req.URL.Query() {
		q[key] = val
	}
	err := db.DB.C("cashiers").Find(q).One(&cashier)
	if err != nil {
		handlers.ReturnJSONError(w, err.Error())
		return
	}
	// TODO: lock terminal, if terminal is locked return error with
	// the cashier currently using the locked terminal.
	handlers.ReturnJSONMessage(w, cashier)
}
