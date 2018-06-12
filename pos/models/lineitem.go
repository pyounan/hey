package models

// Item swagger:model item
// defines attributes of Item entity
type Item struct {
	ID                    int                    `json:"id" bson:"id"`
	Number                string                 `json:"number" bson:"number"`
	Description           string                 `json:"description" bson:"description"`
	UnitPrince            float64                `json:"unit_price" bson:"unit_price"`
	StoreUnitID           int                    `json:"store_unit" bson:"store_unit"`
	ItemID                int                    `json:"item" bson:"item"`
	OpenItem              bool                   `json:"open_item" bson:"open_item"`
	OpenPrice             bool                   `json:"open_price" bson:"open_price"`
	AttachedAttributes    map[string]interface{} `json:"attached_attributes" bson:"attached_attributes"`
	StoreMenuItemConfig   int                    `json:"storemenuitemconfig" bson:"storemenuitemconfig"`
	BaseUnit              string                 `json:"base_unit,omitempty" bson:"base_unit,omitempty"`
	CategoryID            int                    `json:"category_id" bson:"category_id"`
	GroupID               int                    `json:"group_id" bson:"group_id"`
	GroupName             string                 `json:"group_name" bson:"group_name"`
	ItemCondimentGroupSet []ItemCondimentGroup   `json:"itemcondimentgroup_set" bson:"itemcondimentgroup_set"`
}

// ItemCondimentGroup swagger:model itemCondimentGroup
// defines attributes of ItemCondimentGroup entity combinaation
type ItemCondimentGroup struct {
	CondimentGroupID int    `json:"condiment_group" bson:"condiment_group"`
	Min              int    `json:"min" bson:"min"`
	Max              int    `json:"max" bson:"max"`
	Name             string `json:"name" bson:"name"`
}

// CondimentLineItem maps condimentlineitem_set in POSInvoiceLineItem
type CondimentLineItem struct {
	ID                  int                    `json:"id,omitempty" bson:"id,omitempty"`
	Condiment           int                    `json:"condiment" bson:"condiment"`
	LineItem            int                    `json:"posinvoicelineitem" bson:"posinvoicelineitem"`
	Description         string                 `json:"name" bson:"name"`
	Item                string                 `json:"item" bson:"item"`
	UnitPrice           float64                `json:"unit_price,omitempty" bson:"unit_price,omitempty"`
	Price               float64                `json:"price" bson:"price"`
	NetAmount           float64                `json:"net_amount" bson:"net_amount"`
	TaxAmount           float64                `json:"tax_amount" bson:"tax_amount"`
	VAT                 string                 `json:"vat_code" bson:"vat_code"`
	VATPercentage       float64                `json:"vat_percentage" bson:"vat_percentage"`
	AttachedAttributes  map[string]interface{} `json:"attached_attributes" bson:"attached_attributes"`
	StoreMenuItemConfig *int64                 `json:"storemenuitemconfig" bson:"storemenuitemconfig"`
}

// AppliedDiscount maps a discount of posinvoicelineitem
type AppliedDiscount struct {
	Amount     float64 `json:"amount" bson:"amount"`
	Percentage float64 `json:"percentage" bson:"percentage"`
	Type       string  `json:"type" bson:"type"`
}

// GroupedAppliedDiscount represents a group of applied discounts grouped
// by vat code
type GroupedAppliedDiscount struct {
	Amount        float64 `json:"amount" bson:"amount"`
	Percentage    float64 `json:"percentage" bson:"percentage"`
	Type          string  `json:"type" bson:"type"`
	VAT           string  `json:"vat_code" bson:"vat_code"`
	VATPercentage float64 `json:"vat_percentage" bson:"vat_percentage"`
	NetAmount     float64 `json:"net_amount" bson:"net_amount"`
	TaxAmount     float64 `json:"tax_amount" bson:"tax_amount"`
}

// POSLineItem maps POSInvoiceLineItem of the backend
type POSLineItem struct {
	ID                      int                      `json:"id,omitempty" bson:"id,omitempty"`
	Item                    int                      `json:"item" bson:"item"`
	Quantity                float64                  `json:"qty" bson:"qty"`
	SubmittedQuantity       float64                  `json:"submitted_qty" bson:"submitted_qty"`
	ReturnedQuantity        float64                  `json:"returned_qty" bson:"returned_qty"`
	Description             string                   `json:"description" bson:"description"`
	Comment                 string                   `json:"comment" bson:"comment"`
	UnitPrice               float64                  `json:"unit_price" bson:"unit_price"`
	Price                   float64                  `json:"price" bson:"price"`
	NetAmount               float64                  `json:"net_amount" bson:"net_amount"`
	TaxAmount               float64                  `json:"tax_amount" bson:"tax_amount"`
	VAT                     string                   `json:"vat_code" bson:"vat_code"`
	VATPercentage           float64                  `json:"vat_percentage" bson:"vat_percentage"`
	LineItemType            string                   `json:"lineitem_type" bson:"lineitem_type"`
	IsCondiment             bool                     `json:"is_condiment" bson:"is_condiment"`
	CondimentLineItems      []CondimentLineItem      `json:"condimentlineitem_set" bson:"condimentlineitem_set"`
	CondimentGroup          []map[string]interface{} `json:"itemcondimentgroup_set" bson:"itemcondimentgroup_set"`
	IsDiscount              bool                     `json:"is_discount" bson:"is_discount"`
	IsVoid                  bool                     `json:"is_void,omitempty" bson:"is_void,omitempty"`
	AppliedDiscounts        []AppliedDiscount        `json:"applied_discounts" bson:"applied_discounts"`
	GroupedAppliedDiscounts []GroupedAppliedDiscount `json:"grouped_applieddiscounts" bson:"grouped_applieddiscounts"`
	AttachedAttributes      map[string]interface{}   `json:"attached_attributes" bson:"attached_attributes"`
	Course                  int                      `json:"course,omitempty" bson:"course,omitempty"`
	StoreMenuItemConfig     int                      `json:"storemenuitemconfig" bson:"storemenuitemconfig"`
	OpenItem                bool                     `json:"open_item" bson:"open_item"`
	OpenPrice               bool                     `json:"open_price" bson:"open_price"`
	ReturnedIDs             []string                 `json:"returned_ids" bson:"returned_ids"`
	FrontendID              string                   `json:"frontend_id" bson:"frontend_id"`
	UpdatedOn               string                   `json:"updated_on" bson:"updated_on"`
	StoreUnit               int                      `json:"store_unit,omitempty" bson:"store_unit,omitempty"`
	BaseUnit                string                   `json:"base_unit,omitempty" bson:"base_unit,omitempty"`
	OriginalFrontendID      *string                  `json:"original_frontend_id" bson:"original_frontend_id"`
	OriginalLineItemID      *int64                   `json:"original_line_item_id" bson:"original_line_item_id"`
	MenuID                  *int64                   `json:"menu" bson:"menu"`
	CashierID               *int64                   `json:"cashier" bson:"cashier"`
	// used for waste
	PosinvoiceID      *int64 `json:"posinvoice" bson:"posinvoice"`
	Reason            string `json:"reason,omitempty" bson:"reason,omitempty"`
	CondimentsComment string `json:"condiments_comment" bson:"-"`
	LastChildInCourse bool   `json:"last_child_in_course" bson:"-"`
}
