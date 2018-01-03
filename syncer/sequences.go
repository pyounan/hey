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

	pb "gopkg.in/cheggaaa/pb.v1"
	"gopkg.in/mgo.v2/bson"
)

var SequencesAPI = "/api/pos/proxy/sequences/"

func PullSequences() {
	fmt.Println("Pulling FDM Sequences from CloudInn services...")
	// create console progress bar
	bar := pb.StartNew(3)
	netClient := helpers.NewNetClient()
	uri := fmt.Sprintf("%s%s", config.Config.BackendURI, SequencesAPI)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println(err.Error())
	}
	bar.Increment()
	req = helpers.PrepareRequestHeaders(req)
	response, err := netClient.Do(req)
	if err != nil {
		bar.FinishPrint(err.Error())
		return
	}
	defer response.Body.Close()
	bar.Increment()
	if response.StatusCode != 200 {
		log.Printf("Failed to load api from backend: %s\n", uri)
		bar.Finish()
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
	bar.Increment()
	bar.Finish()
}
