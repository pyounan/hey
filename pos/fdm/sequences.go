package fdm

import (
	"pos-proxy/db"
	"pos-proxy/pos/models"
	"pos-proxy/syncer"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/mgo.v2/bson"
)

var sequenceMutex = &sync.Mutex{}
var ticketMutex = &sync.Mutex{}

// GetNextSequence checks the database for the last used sequence
// and returns the next one to be used, the counter fallsback to 1
// if it exceeds 99.
func GetNextSequence(rcrs string) (int, error) {
	var data models.Sequence
	sequenceMutex.Lock()
	defer sequenceMutex.Unlock()
	q := bson.M{"key": "last_sequence", "rcrs": rcrs}
	err := db.DB.C("metadata").With(db.Session.Copy()).Find(q).One(&data)
	if err != nil {
		// if sequence number doesnt exist for this rcrs, create new one with zero value
		data = models.Sequence{}
		data.Key = "last_sequence"
		data.Value = 0
		atomic.AddUint64(&data.Value, 1)
		data.RCRS = rcrs
		data.UpdatedAt = time.Now()
		db.DB.C("metadata").With(db.Session.Copy()).Insert(data)
		go syncer.QueueRequest(syncer.SequencesAPI, "POST", nil, data)
	} else if data.Value == 99 {
		data.Value = 0
		atomic.AddUint64(&data.Value, 1)
		data.UpdatedAt = time.Now()
		db.DB.C("metadata").With(db.Session.Copy()).Update(q, bson.M{"$set": bson.M{"value": data.Value}})
		go syncer.QueueRequest(syncer.SequencesAPI, "POST", nil, data)
	} else {
		atomic.AddUint64(&data.Value, 1)
		data.UpdatedAt = time.Now()
		db.DB.C("metadata").With(db.Session.Copy()).Update(q, bson.M{"$set": bson.M{"value": data.Value}})
		go syncer.QueueRequest(syncer.SequencesAPI, "POST", nil, data)
	}
	return int(data.Value), nil
}

// GetNextTicketNumber checks the database for the last ticket number used for the passed RCRS
// if it doesn't exists, it creates a new one with zero value;
// Then increase the retrieved number by one, if number exceeds 999999, it fallsback
// to one again.
func GetNextTicketNumber(rcrs string) (int, error) {
	ticketMutex.Lock()
	defer ticketMutex.Unlock()
	var data models.Sequence
	q := bson.M{"key": "last_ticket_number", "rcrs": rcrs}
	err := db.DB.C("metadata").With(db.Session.Copy()).Find(q).One(&data)
	if err != nil {
		// if ticket number doesnt exist for this rcrs, create new one with zero value
		data = models.Sequence{}
		data.Key = "last_ticket_number"
		data.Value = 0
		data.RCRS = rcrs
		db.DB.C("metadata").With(db.Session.Copy()).Insert(data)
	}

	if data.Value == 999999 {
		return 1, nil
	} else {
		return int(data.Value) + 1, nil
	}
}

// UpdateLastTicketNumber update the last ticket number in database for the passed RCRS.
func UpdateLastTicketNumber(rcrs string, val int) error {
	q := bson.M{"key": "last_ticket_number", "rcrs": rcrs}
	t := time.Now()
	err := db.DB.C("metadata").With(db.Session.Copy()).Update(q,
		bson.M{"$set": bson.M{"value": val, "updated_at": t}})
	if err != nil {
		return err
	}
	q["value"] = val
	q["updated_at"] = t
	go syncer.QueueRequest(syncer.SequencesAPI, "POST", nil, q)
	return nil
}
