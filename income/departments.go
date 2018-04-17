package income

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/income/models"
	"strconv"

	"github.com/gorilla/mux"

	"pos-proxy/helpers"

	"gopkg.in/mgo.v2/bson"
)

// ListDepartment swagger:route GET /income/api/department/ income getDepartmentList
//
// List Departments
//
// returns a list of income departments
//
// Responses:
//   200: []department
func ListDepartments(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	for key, val := range queryParams {
		if key == "pos_payment" {
			query[key] = true
		} else if key == "type" && (val[0] == "waste" || val[0] == "revenue") {
			query[key] = "debit"
		} else {
			query[key] = val[0]
		}
	}
	departments := []models.Department{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("departments").With(session).Find(query).All(&departments)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, departments)
}

// GetDepartment swagger:route GET /income/api/department/{id}/ income getDepartment
//
// Get Department
//
// returns a details of a department by id
//
// Paramters:
// + name: id
//   in: path
//   required: true
//   schema:
//      type: integer
// Responses:
//   200: department
func GetDepartment(w http.ResponseWriter, r *http.Request) {
	q := bson.M{}
	vars := mux.Vars(r)
	id, _ := vars["id"]
	q["id"], _ = strconv.Atoi(id)

	department := models.Department{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("departments").With(session).Find(q).One(&department)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, department)
}
