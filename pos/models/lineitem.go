package models

import (
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// CondimentLineItem maps condimentlineitem_set in POSInvoiceLineItem
type CondimentLineItem struct {
	ID          int    `json:"id,omitempty" bson:"id,omitempty"`
	Condiment   int    `json:"condiment" bson:"condiment"`
	LineItem    int    `json:"posinvoicelineitem" bson:"posinvoicelineitem"`
	Description string `json:"name" bson:"name"`
	// Item          *int64 `json:"item" bson:"item"`
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

type GroupedAppliedDiscount struct {
	Amount        float64 `json:"amount" bson:"amount"`
	Percentage    float64 `json:"percentage" bson:"percentage"`
	Type          string  `json:"type" bson:"type"`
	VAT           string  `json:"vat_code" bson:"vat_code"`
	VATPercentage float64 `json:"vat_percentage" bson:"vat_percentage"`
	NetAmount     float64 `json:"net_amount" bson:"net_amount"`
	TaxAmount     float64 `json:"tax_amount" bson:"tax_amount"`
}

type EJItem struct {
	Description   string  `json:"description" bson:"description"`
	Price         float64 `json:"price" bson:"price"`
	NetAmount     float64 `json:"net_amount" bson:"net_amount"`
	TaxAmount     float64 `json:"tax_amount" bson:"tax_amount"`
	VAT           string  `json:"vat_code" bson:"vat_code"`
	VATPercentage float64 `json:"vat_percentage" bson:"vat_percentage"`
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
	// used for waste
	PosinvoiceID *int64 `json:"posinvoice,omitempty" bson:"posinvoice,omitempty"`
	CashierID    *int64 `json:"cashier,omitempty" bson:"cashier,omitempty"`
	Reason       string `json:"reason,omitempty" bson:"reason,omitempty"`
}

// String generates a text for a line item in a format for the FDM.
func (l POSLineItem) String() string {
	// quantity length should be 4 letters, if len is smaller than 4, prepend zeros
	l.Quantity = math.Abs(l.Quantity)
	q := strconv.FormatFloat(l.Quantity, 'f', 0, 64)
	q = strings.Replace(q, ".", "", 1)
	q = strings.Replace(q, "-", "", 1)
	if len(q) < 4 {
		diff := 4 - len(q)
		for i := 0; i < diff; i++ {
			q = "0" + q
		}
	}
	// desc len should be 20, if len is smaller than 20, append spaces to the right
	// remove all spaces from the description
	reg := regexp.MustCompile(`[^A-Za-z0-9]`)
	desc := reg.ReplaceAllString(l.Description, "")
	desc = strings.ToUpper(desc)
	d := desc
	if len(d) > 20 {
		d = d[:20]
	} else if len(d) < 20 {
		diff := 20 - len(d)
		for i := 0; i < diff; i++ {
			d += " "
		}
	}
	// price len should be 8, if len is smaller than 8, prepend zeros
	l.Price = math.Abs(l.Price)
	p := strconv.FormatFloat(l.Price, 'f', 2, 64)
	p = strings.Replace(p, ".", "", 1)
	p = strings.Replace(p, "-", "", 1)
	if len(p) < 8 {
		diff := 8 - len(p)
		for i := 0; i < diff; i++ {
			p = "0" + p
		}
	}

	l.NetAmount = math.Abs(l.NetAmount)

	// make sure the len of res = 33
	log.Println("Item VAT", l.Description, l.VAT)
	result := q + d + p + string(l.VAT[0])
	return result
}
