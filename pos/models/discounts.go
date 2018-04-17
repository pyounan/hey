package models

// FixedDiscount swagger:model fixedDiscount
// defines attributes of a Fixed Discount entity
type FixedDiscount struct {
	ID         int     `json:"id" bson:"id"`
	Name       string  `json:"name" bson:"name"`
	Amount     float64 `json:"amount" bson:"amount"`
	Percentage float64 `json:"percentage" bson:"percentage"`
	CashierIDs []int   `json:"cashiers" bson:"cashiers"`
	StoreIDs   []int   `json:"stores" bson:"stores"`
	ItemIDs    []int   `json:"items" bson:"items"`
}
