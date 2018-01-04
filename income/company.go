package income

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
)

func GetCompany(w http.ResponseWriter, r *http.Request) {
	company := make(map[string]interface{})
	db.DB.C("company").With(db.Session.Copy()).Find(nil).One(&company)

	helpers.ReturnSuccessMessage(w, company)
}
