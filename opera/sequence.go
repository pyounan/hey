package opera

import (
	"pos-proxy/db"
	"pos-proxy/pos/models"
	"sync"
	"sync/atomic"

	"gopkg.in/mgo.v2/bson"
)

var seqMutex = &sync.Mutex{}

func GetNextSequence() (int, error) {
	seqMutex.Lock()
	defer seqMutex.Unlock()
	session := db.Session.Copy()
	defer session.Close()
	var data models.Sequence
	q := bson.M{"key": "last_sequence"}
	err := db.DB.C("operametadata").With(session).Find(q).One(&data)
	if err != nil {
		data = models.Sequence{}
		data.Key = "last_sequence"
		data.Value = 0
		atomic.AddUint64(&data.Value, 1)
		err = db.DB.C("operametadata").With(session).Insert(data)
	} else {
		atomic.AddUint64(&data.Value, 1)
		db.DB.C("operametadata").Update(q, bson.M{"$set": bson.M{"value": data.Value}})
	}
	return int(data.Value), nil
}
