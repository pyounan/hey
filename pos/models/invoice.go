package models

import "time"

type PaymentLog struct {
}

type Invoice struct {
	ID               *int64        `json:"id" bson:"id,omitempty"`
	InvoiceNumber    string        `json:"invoice_number" bson:"invoice_number"`
	Items            []POSLineItem `json:"posinvoicelineitem_set" bson:"posinvoicelineitem_set"`
	GroupedLineItems []EJEvent     `json:"grouped_lineitems" bson:"grouped_lineitems"`
	TableNumber      *int64        `json:"table" bson:"table"`

	Events []EJEvent `json:"events" bson:"events"`

	AuditDate           string                 `json:"audit_date" bson:"audit_date"`
	Cashier             int                    `json:"cashier" bson:"cashier"`
	CashierDetails      string                 `json:"cashier_details" bson:"cashier_details"`
	CashierNumber       int                    `json:"cashier_number" bson:"cashier_number"`
	CreatedOn           string                 `json:"created_on" bson:"created_on"`
	FrontendID          string                 `json:"frontend_id" bson:"frontend_id"`
	IsSettled           bool                   `json:"is_settled" bson:"is_settled"`
	PaidAmount          float64                `json:"paid_amount" bson:"paid_amount"`
	Pax                 int                    `json:"pax" bson:"pax"`
	WalkinName          *string                `json:"walkin_name" bson:"walkin_name"`
	ProfileName         *int64                 `json:"profile_name" bson:"profile_name"`
	ProfileDetails      *string                `json:"profile_details" bson:"profile_details"`
	Store               int                    `json:"store" bson:"store"`
	StoreDescription    string                 `json:"store_description" bson:"store_description"`
	Subtotal            float64                `json:"subtotal" bson:"subtotal"`
	TableID             *int64                 `json:"table_number" bson:"table_number"`
	TakeOut             bool                   `json:"takeout" bson:"takeout"`
	TerminalID          int                    `json:"terminal_id" bson:"terminal_id"`
	TerminalDescription string                 `json:"terminal_description" bson:"terminal_description"`
	Total               float64                `json:"total" bson:"total"`
	CreateLock          bool                   `json:"create_lock" bson:"create_lock"`
	FDMResponses        []FDMResponse          `json:"fdm_responses" bson:"fdm_responses"`
	Postings            []Posting              `json:"pospayment" bson:"pospayment"`
	Room                *int64                 `json:"room,omitempty" bson:"room,omitempty"`
	RoomNumber          *int64                 `json:"room_number,omitempty" bson:"room_number,omitempty"`
	RoomDetails         *string                `json:"room_details,omitempty" bson:"room_details,omitempty"`
	HouseUse            bool                   `json:"house_use" bson:"house_use"`
	PrintCount          int                    `json:"print_count" bson:"print_count"`
	Taxes               map[string]interface{} `json:"taxes" bson:"taxes"`
	VoidReason          string                 `json:"void_reason,omitempty" bson:"void_reason,omitempty"`
	Change              float64                `json:"change,omitempty" bson:"change,omitempty"`
	ClosedOn            *time.Time             `json:"closed_on" bson:"closed_on"`
	UpdatedOn           time.Time              `json:"updated_on" bson:"updated_on"`
	OperaReservation    string                 `json:"opera_reservation" bson:"opera_reservation"`
	OperaRoomNumber     string                 `json:"opera_room_number" bson:"opera_room_number"`
}
