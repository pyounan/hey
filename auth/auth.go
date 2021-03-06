package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

var Token string

// EnsureTenantSelected swagger:route GET /api/ensure_tenant_selected/ auth ensureTenantSelected
//
// Ensure Tenant Selected
//
// in case of proxy always returns true
func EnsureTenantSelected(w http.ResponseWriter, req *http.Request) {
	helpers.ReturnSuccessMessage(w, true)
}

// GetUserPermissions swagger:route GET /core/getloggedinusergroups/{id}/ auth getUserPermissions
//
// List User Permissions
//
// returns a list of permissions assigned for the logged in user
//
// Parameters:
// + name: id
//   in: path
//   required: true
//   schema:
//      type: integer
func GetUserPermissions(w http.ResponseWriter, req *http.Request) {
	var perms map[string]interface{}
	vars := mux.Vars(req)
	idStr := vars["id"]
	id, _ := strconv.Atoi(idStr)
	q := bson.M{}
	q["id"] = id
	err := db.DB.C("usergroups").With(db.Session.Copy()).Find(q).One(&perms)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, perms["permissions"])
}

func FetchToken() {
	type TokenResponse struct {
		Token string `json:"token"`
	}
	netClient := helpers.NewNetClient()
	uri := fmt.Sprintf("%s%s", config.Config.BackendURI, "/api/is_authenticated/")
	req, err := http.NewRequest("GET", uri, strings.NewReader(""))
	if err != nil {
		log.Println("Failed to create new request", err.Error())
		return
	}
	req = helpers.PrepareRequestHeaders(req)
	resp, err := netClient.Do(req)
	if err != nil {
		log.Println("Failed to perform request", err.Error())
		return
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read body", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		data := TokenResponse{}
		err = json.Unmarshal(respBody, &data)
		if err != nil {
			log.Println("Failed to parse update data", string(respBody), err.Error())
			return
		}

		Token = data.Token
	} else {
		log.Println("Failed to fetch token", resp.StatusCode)
	}
}
