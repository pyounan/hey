package income

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
)

func GetCompany(w http.ResponseWriter, r *http.Request) {
	session := db.Session.Copy()
	defer session.Close()
	company := make(map[string]interface{})
	db.DB.C("company").With(session).Find(nil).One(&company)

	helpers.ReturnSuccessMessage(w, company)
}
