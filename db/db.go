package db

import (
	"gopkg.in/mgo.v2"
)

// Session represents the main db session that we take copies from
var Session *mgo.Session

// DB instance of mongo database connection
var DB *mgo.Database

// Connect sets the connection to mongodb and make it
// available as a global variable to be used by other
// packages
func Connect() error {
	var err error
	Session, err = mgo.Dial("localhost")
	if err != nil {
		return err
	}
	Session.SetMode(mgo.Monotonic, true)
	DB = Session.DB("cloudinn")
	return nil
}

// Close closes the main session
func Close() {
	Session.Close()
}
