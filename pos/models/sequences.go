package models

import (
	"time"
)

type Sequence struct {
	Key       string    `bson:"key" bson:"key"`
	Value     uint64    `bson:"value" bson:"value"`
	RCRS      string    `json:"rcrs" bson:"rcrs"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
