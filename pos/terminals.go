package pos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/locks"
	"pos-proxy/pos/models"
	"pos-proxy/syncer"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

// ListTerminals returns a json response with list of terminals,
// could be queries by store
func ListTerminals(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	for key, val := range queryParams {
		if key == "store" {
			num, err := strconv.Atoi(val[0])
			if err != nil {
				helpers.ReturnErrorMessage(w, err.Error())
				return
			}
			query[key] = num
		} else {
			query[key] = val
		}
	}
	terminals := []models.Terminal{}
	err := db.DB.C("terminals").Find(query).All(&terminals)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	for i := 0; i < len(terminals); i++ {
		key := fmt.Sprintf("terminal_%d", terminals[i].ID)
		_, err := db.Redis.Get(key).Result()
		if err == redis.Nil {
			terminals[i].IsLocked = false
		} else {
			terminals[i].IsLocked = true
		}
	}
	helpers.ReturnSuccessMessage(w, terminals)
}

// GetTerminal returns a json response with the specified terminal id
func GetTerminal(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, _ := strconv.Atoi(idStr)
	query["id"] = id
	terminal := models.Terminal{}
	err := db.DB.C("terminals").Find(query).One(&terminal)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, terminal)
}

// UpdateTerminal returns a json response with the specified terminal id
func UpdateTerminal(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, _ := strconv.Atoi(idStr)
	query["id"] = id
	terminal := models.Terminal{}
	err := db.DB.C("terminals").Find(query).One(&terminal)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	newTerminal := models.Terminal{}
	err = json.NewDecoder(r.Body).Decode(&newTerminal)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	err = db.DB.C("terminals").Update(query, newTerminal)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	// proxy the update to backend
	syncer.QueueRequest(r.URL.Path, r.Method, r.Header, newTerminal)
	helpers.ReturnSuccessMessage(w, newTerminal)
}

// UnlockTerminal removes the terminal redis key and make the terminal
// available for other cashiers again
func UnlockTerminal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, _ := vars["id"]
	id, _ := strconv.Atoi(idStr)
	locks.UnlockTerminal(id)
	helpers.ReturnSuccessMessage(w, true)
}
