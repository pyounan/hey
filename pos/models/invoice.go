package models

type Event struct {
	Item POSLineItem `json:"item"`
}

type Invoice struct {
	InvoiceNumber string        `json:"invoice_number" bson:"invoice_number"`
	Items         []POSLineItem `json:"posinvoicelineitem_set" bson:"posinvoicelineitem_set"`
	TableNumber   int           `json:"table" bson:"table"`

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
	Subtotal            float64                `json:"subtotal" bson:"Subtotal"`
	TableID             int                    `json:"table_number" bson:"table_number"`
	TakeOut             bool                   `json:"takeout" bson:"takeout"`
	TerminalID          int                    `json:"terminal_id" bson:"terminal_id"`
	TerminalDescription string                 `json:"terminal_description" bson:"terminal_description"`
	Total               float64                `json:"total" bson:"total"`
	FDMResponses        []FDMResponse          `json:"fdm_responses" bson:"fdm_responses"`
	Payments            []Payment              `json:"pospayment" bson:"pospayment"`
	RoomID              int                    `json:"room_number" bson:"room_number"`
	Room                string                 `json:"room" bson:"room"`
	HouseUse            bool                   `json:"house_use" bson:"house_use"`
	PrintCount          int                    `json:"print_count" bson:"print_count"`
	Taxes               map[string]interface{} `json:"taxes" bson:"taxes"`
}
