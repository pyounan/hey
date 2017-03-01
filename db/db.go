package db

import (
	"gopkg.in/mgo.v2"
)

var DB *mgo.Database

func init() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	DB = session.DB("cloudinn_pos")
}
