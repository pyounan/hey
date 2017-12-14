package auth

import (
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"strconv"
)

func EnsureTenantSelected(w http.ResponseWriter, req *http.Request) {
	helpers.ReturnSuccessMessage(w, true)
}

func GetUserPermissions(w http.ResponseWriter, req *http.Request) {
	var perms map[string]interface{}
	vars := mux.Vars(req)
	idStr := vars["id"]
	id, _ := strconv.Atoi(idStr)
	q := bson.M{}
	q["id"] = id
	err := db.DB.C("usergroups").Find(q).One(&perms)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, perms["permissions"])
}
