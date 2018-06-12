package models

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// EJEvent swagger:model ejevent
// is a model that gets inserted to FDM or EJ.
type EJEvent struct {
	Description   string  `json:"description" bson:"description"`
	Price         float64 `json:"price" bson:"price"`
	Quantity      float64 `json:"qty" bson:"qty"`
	NetAmount     float64 `json:"net_amount" bson:"net_amount"`
	TaxAmount     float64 `json:"tax_amount" bson:"tax_amount"`
	VATCode       string  `json:"vat_code" bson:"vat_code"`
	VATPercentage float64 `json:"vat_percentage" bson:"vat_percentage"`
	IsCondiment   bool    `json:"is_condiment" bson:"is_condiment"`
	IsDiscount    bool    `json:"is_discount" bson:"is_discount"`
	LineItemType  string  `json:"line_item_type" bson:"line_item_type"`
	Item          *int64  `json:"item,omitempty" bson:"item,omitempty"`
	CashierID     *int64  `json:"cashier" bson:"cashier"`
}

// String generates a text for a ej event in a format for the FDM.
func (l EJEvent) String() string {
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

	result := q + d + p + string(l.VATCode[0])
	return result
}
