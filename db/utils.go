package db

import (
	"gopkg.in/mgo.v2/bson"
	"sync"
)

type MetaData struct {
	Key   string `bson:"key"`
	Value int    `bson:"value"`
	RCRS  string `bson:"rcrs"`
}

// GetNextSequence checks the database for the last used sequence
// and returns the next one to be used, the counter fallsback to 1
// if it exceeds 99.
func GetNextSequence(rcrs string) (int, error) {
	var mutex = &sync.Mutex{}
	mutex.Lock()
	var data MetaData
	q := bson.M{"key": "last_sequence", "rcrs": rcrs}
	err := DB.C("metadata").Find(q).One(&data)
	if err != nil {
		// if sequence number doesnt exist for this rcrs, create new one with zero value
		data = MetaData{}
		data.Key = "last_sequence"
		data.Value = 1
		data.RCRS = rcrs
		DB.C("metadata").Insert(data)
	} else if data.Value == 99 {
		data.Value += 1
		DB.C("metadata").Update(q, bson.M{"$set": bson.M{"value": data.Value}})
	}
	mutex.Unlock()
	return data.Value, nil
}

// GetNextTicketNumber checks the database for the last ticket number used for the passed RCRS
// if it doesn't exists, it creates a new one with zero value;
// Then increase the retrieved number by one, if number exceeds 999999, it fallsback
// to one again.
func GetNextTicketNumber(rcrs string) (int, error) {
	var data MetaData
	q := bson.M{"key": "last_ticket_number", "rcrs": rcrs}
	err := DB.C("metadata").Find(q).One(&data)
	if err != nil {
		// if ticket number doesnt exist for this rcrs, create new one with zero value
		data = MetaData{}
		data.Key = "last_ticket_number"
		data.Value = 0
		data.RCRS = rcrs
		DB.C("metadata").Insert(data)

	}

	if data.Value == 999999 {
		return 1, nil
	} else {
		return data.Value + 1, nil
	}
}

// UpdateLastSequence updates the last used sequence in the database
func UpdateLastSequence(rcrs string, val int) error {
	q := bson.M{"key": "last_sequence", "rcrs": rcrs}
	err := DB.C("metadata").Update(q,
		bson.M{"$set": bson.M{"value": val}})
	return err
}

// UpdateLastTicketNumber update the last ticket number in database for the passed RCRS.
func UpdateLastTicketNumber(rcrs string, val int) error {
	q := bson.M{"key": "last_ticket_number", "rcrs": rcrs}
	err := DB.C("metadata").Update(q,
		bson.M{"$set": bson.M{"value": val}})
	return err
}
