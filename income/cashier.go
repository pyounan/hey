package income

import (
	"fmt"
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/locks"
	"strconv"

	"log"

	"gopkg.in/mgo.v2/bson"
)

type Cashier struct {
	ID         int    `json:"id" bson:"id"`
	Name       string `json:"name" bson:"name"`
	Number     int    `json:"number" bson:"number"`
	EmployeeID string `json:"employee_id" bson:"employee_id"`
}

func GetPosCashier(w http.ResponseWriter, req *http.Request) {
	cashier := Cashier{}
	q := bson.M{}
	store, _ := strconv.Atoi(req.URL.Query().Get("store"))
	pin := req.URL.Query().Get("pin")
	terminal := req.URL.Query().Get("terminal")
	terminalID, _ := strconv.Atoi(terminal)
	q["pin"] = pin
	q["store_set"] = store
	err := db.DB.C("cashiers").Find(q).One(&cashier)
	if err != nil {
		resp := bson.M{"ok": false, "details": "No matching PIN code to selected store."}
		helpers.ReturnSuccessMessage(w, resp)
		return
	}

	cashierHashExists := true
	_, err = req.Cookie("cashier_hash")
	if err != nil {
		cashierHashExists = false
	}
	otherCashier, err := locks.LockTerminal(terminalID, cashier.ID)
	if err != nil && ((cashierHashExists && otherCashier == cashier.ID) || otherCashier != cashier.ID) {
		log.Println(err)
		resp := bson.M{"ok": false, "details": "Terminal is locked."}
		helpers.ReturnErrorMessage(w, resp)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "cashier_id",
		Value: fmt.Sprintf("%d", cashier.ID),
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
