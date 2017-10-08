package models

// Posting models POSPayment
type Posting struct {
	ID                     int        `json"id,omitempty" bson:"id,omitempty"`
	Amount                 float64    `json:"amount" bson:"amount"`
	AuditDate              string     `json:"audit_date,omitempty" bson:"audit_date,omitempty"`
	CashierDetails         string     `json:"cashier_details" bson:"cashier_details"`
	Cashier                int        `json:"cashier" bson:"cashier"`
	CashierID              int        `json:"cashier_id" bson:"cashier_id"`
	Comments               string     `json:"comments" bson:"comments"`
	CurrencyID             int        `json:"currency_id" bson:"currency_id"`
	Currency               int        `json:"currency" bson:"currency"`
	CurrencyDetails        string     `json"currency_details" bson:"currency_details"`
	Department             int        `json:"department,omitempty" bson:"department,omitempty"`
	DepartmentDetails      string     `json:"department_details" bson:"department_details"`
	ForeignAmount          float64    `json:"foreign_amount" bson:"foreign_amount"`
	FrontendID             string     `json:"frontend_id" bson:"frontend_id"`
	PosinvoiceID           int        `json:"posinvoice_id" bson:"posinvoice_id"`
	PostingType            string     `json:"posting_type" bson:"posting_type"`
	Room                   *int64     `json:"room" bson:"room"`
	RoomNumber             *int64     `json:"room_number" bson:"room_number"`
	RoomDetails            *string    `json:"room_details" bson:"room_details"`
	PosPostingInformations []Posting  `json:"pospostinginformations" bson:"pospostinginformations"`
	PaymentLog             PaymentLog `json:"paymentlog" bson:"paymentlog"`
	// pospostinginformatios only
	Sign      string `json:"sign,omitempty" bson:"sign,omitempty"`
	Type      string `json:"type,omitempty" bson:"type,omitempty"`
	Cancelled bool   `json:"cancelled,omitempty" bson:"cancelled,omitempty"`
}
