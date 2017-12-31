package syncer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/models"

	"gopkg.in/mgo.v2/bson"
)

var SequencesAPI = "/api/pos/proxy/sequences/"

func PullSequences() {
	netClient := helpers.NewNetClient()
	uri := fmt.Sprintf("%s%s", config.Config.BackendURI, SequencesAPI)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println(err.Error())
	}
	req = helpers.PrepareRequestHeaders(req)
	response, err := netClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		log.Printf("Failed to load api from backend: %s\n", uri)
		return
	}
	list := []models.Sequence{}
	err = json.NewDecoder(response.Body).Decode(&list)
	if err != nil {
		log.Println(err.Error())
	}
	for _, s := range list {
		oldS := models.Sequence{}
		q := bson.M{"key": s.Key, "rcrs": s.RCRS}
		err := db.DB.C("metadata").Find(q).One(&oldS)
		if err != nil {
			log.Println("inserting new sequnce")
			db.DB.C("metadata").Insert(s)
			continue
		}
		if s.UpdatedAt.After(oldS.UpdatedAt) {
			log.Println("updating sequnce")
			db.DB.C("metadata").Upsert(q, s)
		}
	}
}
