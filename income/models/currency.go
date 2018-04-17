package models

// Currency swagger:model currency
// Defines the attributes of a currency model
type Currency struct {
	ID            int    `json:"id" bson:"id"`
	Code          string `json:"code" bson:"code"`
	IsDefault     bool   `json:"is_default" bson:"is_default"`
	Factor        string `json:"factor" bson:"factor"`
	HasDepartment bool   `json:"has_department" bson:"has_department"`
}
