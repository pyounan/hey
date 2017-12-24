package auth

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"strconv"
	"strings"
)

var Token string

func EnsureTenantSelected(w http.ResponseWriter, req *http.Request) {
	helpers.ReturnSuccessMessage(w, true)
}

func GetUserPermissions(w http.ResponseWriter, req *http.Request) {
	var perms map[string]interface{}
	vars := mux.Vars(req)
	idStr := vars["id"]
	id, _ := strconv.Atoi(idStr)
	q := bson.M{}
	q["id"] = id
	err := db.DB.C("usergroups").Find(q).One(&perms)
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
	req = helpers.PrepareRequestHeaders(req)
	resp, err := netClient.Do(req)
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read body", err.Error())
	}
	data := TokenResponse{}
	err = json.Unmarshal(respBody, &data)
	if err != nil {
		log.Println("Failed to parse update data", string(respBody), err.Error())
	}

	log.Println("resp", data.Token)
	Token = data.Token
	log.Println("Token", Token)
}
