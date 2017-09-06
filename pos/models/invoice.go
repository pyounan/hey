package models

type Event struct {
	Item POSLineItem `json:"item"`
}


type Posting struct {
	Amount float64 `json:"amount" bson:"amount"`
	AuditDate string `json:"audit_date" bson:"audit_date"`
	CashierDetails string `json:"cashier_details" bson:"cashier_details"`
	CashierID int `json:"cashier_id" bson:"cashier_id"`
	Comments string `json:"comments" bson:"comments"`
	Currency int `json:"currency" bson:"currency"`
	CurrencyDetails string `json"currency_details" bson:"currency_details"`
	Department int `json:"department,omitempty" bson:"department,omitempty"`
	DepartmentDetails string `json:"department_details" bson:"department_details"`
	ForeignAmount float64 `json:"foreign_amount" bson:"foreign_amount"`
	FrontendID string `json:"frontend_id" bson:"frontend_id"`
	PosinvoiceID int `json:"posinvoice_id" bson:"posinvoice_id"`
	PostingType string `json:"posting_type" bson:"posting_type"`
	Room int `json:"room,omitempty" bson:"room,omitempty"`
	RoomNumber string `json:"room_number,omitempty" bson:"room_number,omitempty"`
	RoomDetails string `json:"room_details,omitempty" bson:"room_details,omitempty"`
	PosPostingInformations []Posting `json:"pospostinginformations" bson:"pospostinginformations"`
	PaymentLog PaymentLog `json:"paymentlog" bson:"paymentlog"`
	// pospostinginformatios only
	Sign string `json:"sign,omitempty" bson:"sign,omitempty"`
	Type string `json:"type,omitempty" bson:"type,omitempty"`
	Cancelled bool `json:"cancelled,omitempty" bson:"cancelled,omitempty"`
}


type PaymentLog struct {

}

type Invoice struct {
	ID *int64 `json:"id" bson:"id"`
	InvoiceNumber string        `json:"invoice_number" bson:"invoice_number"`
	Items         []POSLineItem `json:"posinvoicelineitem_set" bson:"posinvoicelineitem_set"`
	TableNumber   *int64           `json:"table" bson:"table"`

	Events []Event `json:"events" bson:"events"`

	AuditDate           string                 `json:"audit_date" bson:"audit_date"`
	Cashier             int                    `json:"cashier" bson:"cashier"`
	CashierDetails      string                 `json:"cashier_details" bson:"cashier_details"`
	CashierNumber       int                    `json:"cashier_number" bson:"cashier_number"`
	CreatedOn           string                 `json:"created_on" bson:"created_on"`
	FrontendID          string                 `json:"frontend_id" bson:"frontend_id"`
	IsSettled           bool                   `json:"is_settled" bson:"is_settled"`
	PaidAmount          float64                `json:"paid_amount" bson:"paid_amount"`
	Pax                 float64                `json:"pax" bson:"pax"`
	Store               int                    `json:"store" bson:"store"`
	StoreDescription    string                 `json:"store_description" bson:"store_description"`
	Subtotal            float64                `json:"subtotal" bson:"subtotal"`
	TableID             *int64                    `json:"table_number" bson:"table_number"`
	TakeOut             bool                   `json:"takeout" bson:"takeout"`
	TerminalID          int                    `json:"terminal_id" bson:"terminal_id"`
	TerminalDescription string                 `json:"terminal_description" bson:"terminal_description"`
	Total               float64                `json:"total" bson:"total"`
	FDMResponses        []FDMResponse          `json:"fdm_responses" bson:"fdm_responses"`
	Postings            []Posting              `json:"pospayment" bson:"pospayment"`
	Room                *int64                 `json:"room,omitempty" bson:"room,omitempty"`
	RoomNumber string `json:"room_number,omitempty" bson:"room_number,omitempty"`
	RoomDetails string `json:"room_details,omitempty" bson:"room_details,omitempty"`
	HouseUse            bool                   `json:"house_use" bson:"house_use"`
	PrintCount          int                    `json:"print_count" bson:"print_count"`
	Taxes               map[string]interface{} `json:"taxes" bson:"taxes"`
}
