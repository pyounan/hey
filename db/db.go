package db

import (
	"gopkg.in/mgo.v2"
)

// DB instance of mongo database connection
var DB *mgo.Database

func init() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	DB = session.DB("cloudinn")
}
