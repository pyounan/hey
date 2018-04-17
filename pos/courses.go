package pos

import (
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/models"
	"strconv"

	"gopkg.in/mgo.v2/bson"
)

// ListCourses swagger:route GET /api/pos/course/ courses listCourses
//
// List Courses
//
// returns a list of menu courses
//
// Responses:
//   200: []course
func ListCourses(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	idStr, ok := queryParams["id"]
	if ok {
		id, _ := strconv.Atoi(idStr[0])
		query["id"] = id
	}
	courses := []models.Course{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("courses").With(session).Find(query).Sort("id").All(&courses)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, courses)
}

// GetCourse swagger:route GET /api/pos/course/{id}/ courses getCourse
//
// Get Course
//
// returns details of a Course
//
// Responses:
//   200: course
func GetCourse(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	idStr, ok := queryParams["id"]
	if ok {
		id, _ := strconv.Atoi(idStr[0])
		query["id"] = id
	}
	course := models.Course{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("courses").With(session).Find(query).Sort("id").One(&course)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, course)
}
