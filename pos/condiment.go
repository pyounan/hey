package pos

import (
	"net/http"
)

func ListCondiments(w http.ResponseWriter, r *http.Request) {
	//	query := bson.M{}
	//	queryParams := r.URL.Query()
	//	for key, val := range queryParams {
	//		query[key] = val
	//	}
	//	var condiments []map[string]interface{}
	//	err := db.DB.C("condiments").Find(query).All(&condiments)
	//	if err != nil {
	//		helpers.ReturnErrorMessage(w, err.Error())
	//		return
	//	}
	//	helpers.ReturnSuccessMessage(w, condiments)
}
