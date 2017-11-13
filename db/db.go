package db

import (
	"gopkg.in/mgo.v2"
)

// DB instance of mongo database connection
var DB *mgo.Database

// Connect sets the connection to mongodb and make it
// available as a global variable to be used by other
// packages
func Connect() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	DB = session.DB("cloudinn")
}
