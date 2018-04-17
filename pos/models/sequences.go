package models

import (
	"time"
)

// Sequence swagger:model sequence
// defines a Sequence entity which holds values for sequence per terminal
// a Sequence is used to keep track for last sequence sent to FDM
type Sequence struct {
	Key       string    `bson:"key" bson:"key"`
	Value     uint64    `bson:"value" bson:"value"`
	RCRS      string    `json:"rcrs" bson:"rcrs"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
