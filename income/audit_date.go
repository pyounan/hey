package income

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
)

func GetAuditDate(w http.ResponseWriter, req *http.Request) {
	auditDate := make(map[string]interface{})
	err := db.DB.C("audit_date").With(db.Session.Copy()).Find(nil).One(&auditDate)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	helpers.ReturnSuccessMessage(w, auditDate)
}
