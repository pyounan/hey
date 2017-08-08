package income

import (
	"fmt"
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/locks"
	"strconv"

	"gopkg.in/mgo.v2/bson"
	"log"
)

func GetPosCashier(w http.ResponseWriter, req *http.Request) {
	cashier := make(map[string]interface{})
	q := bson.M{}
	store, _ := strconv.Atoi(req.URL.Query().Get("store"))
	pin := req.URL.Query().Get("pin")
	terminal := req.URL.Query().Get("terminal")
	q["pin"] = pin
	q["store_set"] = store
	err := db.DB.C("cashiers").Find(q).One(&cashier)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	err = locks.LockTerminal(terminal)
	if err != nil {
		log.Println(err)
		if err.Error() == "Couldn't obtain terminal lock." {
			resp := bson.M{"ok": false, "details": err.Error()}
			helpers.ReturnErrorMessage(w, resp)
			return
		}
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "cashier_id",
		Value: fmt.Sprintf("%s", cashier["id"]),
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "cashier_hash",
		Value: "1",
	})

	resp := bson.M{"ok": true, "details": cashier}

	helpers.ReturnSuccessMessage(w, resp)
}

func GetCashierPermissions(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	for key, val := range queryParams {
		query[key] = val
	}
	permissions := []map[string]interface{}{}
	err := db.DB.C("permissions").Find(query).All(&permissions)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, permissions)
}
