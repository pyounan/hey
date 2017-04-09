package db

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var DB *mgo.Database

func init() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	DB = session.DB("cloudinn_pos")
	// make sure that the metadata collection has been created and set default values from last_sequence and last_ticket_number
	count, _ := DB.C("metadata").Count()
	if count == 0 {
		DB.C("metadata").Insert(bson.M{"key": "last_sequence", "value": 0})
		//		DB.C("metadata").Insert(bson.M{"key": "last_ticket_number", "value": 0})
	}

}
