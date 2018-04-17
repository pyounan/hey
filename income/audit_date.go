package income

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
)

// auditDateResponse swagger:model auditDateResponse
type auditDateResponse struct {
	AuditDate string `json:"audit_date" bson:"audit_date"`
}

// GetAuditDate swagger:route GET /shadowinn/api/auditdate/ shadowinn auditDate
//
// Get Audit Date
//
// returns the current audit date of the instance
//
// Responses:
// 200: auditDateResponse
func GetAuditDate(w http.ResponseWriter, req *http.Request) {
	auditDate := auditDateResponse{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("audit_date").With(session).Find(nil).One(&auditDate)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	helpers.ReturnSuccessMessage(w, auditDate)
}
