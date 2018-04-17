package income

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
)

// Company swagger:model company
// defines attributes of a company model
type Company struct {
	Name       string `json:"name" bson:"name"`
	Phone      string `json:"phone" bson:"phone"`
	Country    string `json:"country" bson:"country"`
	City       string `json:"city" bson:"city"`
	Address    string `json:"address" bson:"address"`
	Email      string `json:"email" bson:"email"`
	Fax        string `json:"fax" bson:"fax"`
	Logo       string `json:"logo" bson:"logo"`
	PostalCode string `json:"postal_code" bson:"postal_code"`
	VATNumber  string `json:"vat_number" bson:"vat_number"`
}

// GetCompany swagger:route GET /shadowinn/api/company/ shadowinn company
//
// Get Company
//
// returns the current audit date of the instance
//
// Responses:
// 200: company
func GetCompany(w http.ResponseWriter, r *http.Request) {
	session := db.Session.Copy()
	defer session.Close()
	company := Company{}
	db.DB.C("company").With(session).Find(nil).One(&company)

	helpers.ReturnSuccessMessage(w, company)
}
