package models

// Course swagger:model course
// defines attributes of Course model
type Course struct {
	ID    int    `json:"id" bson:"id"`
	Name  string `json:"name" bson:"name"`
	Order int    `json:"order" bson:"order"`
}
