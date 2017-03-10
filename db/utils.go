package db

import (
	"gopkg.in/mgo.v2/bson"
)

type MetaData struct {
	Key   string `bson:"key"`
	Value int    `bson:"value"`
}

func GetNextSequence() (int, error) {
	var data MetaData
	err := DB.C("metadata").Find(bson.M{"key": "last_sequence"}).One(&data)
	if err != nil {
		return 0, err
	}

	if data.Value == 99 {
		return 1, nil
	} else {
		return data.Value + 1, nil
	}
}

func GetNextTicketNumber() (int, error) {
	var data MetaData
	err := DB.C("metadata").Find(bson.M{"key": "last_ticket_number"}).One(&data)
	if err != nil {
		return 0, err
	}

	if data.Value == 999999 {
		return 1, nil
	} else {
		return data.Value + 1, nil
	}
}

func UpdateLastSequence(val int) error {
	err := DB.C("metadata").Update(bson.M{"key": "last_sequence"},
		bson.M{"$set": bson.M{"value": val}})
	return err
}

func UpdateLastTicketNumber(val int) error {
	err := DB.C("metadata").Update(bson.M{"key": "last_ticket_number"},
		bson.M{"$set": bson.M{"value": val}})
	return err
}
