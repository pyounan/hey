package fdm

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

type POSLineItem struct {
	UUID              string  `json:"frontend_id" bson:"frontend_id"`
	Quantity          float64 `json:"qty" bson:"qty"`
	SubmittedQuantity float64 `json:"submitted_qty" bson:"submitted_qty"`
	Description       string  `json:"description" bson:"description"`
	Price             float64 `json:"price" bson:"price"`
	NetAmount         float64 `json:"net_amount" bson:"net_amount"`
	VAT               string  `json:"vat_code" bson:"vat_code"`
	VATPercentage     float64 `json:"vat_percentage" bson:"vat_percentage"`
}

// String generates a text for a line item in a format for the FDM.
func (l *POSLineItem) String() string {
	// quantity length should be 4 letters, if len is smaller than 4, prepend zeros
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
	p := strconv.FormatFloat(l.Price, 'f', 2, 64)
	p = strings.Replace(p, ".", "", 1)
	p = strings.Replace(p, "-", "", 1)
	if len(p) < 8 {
		diff := 8 - len(p)
		for i := 0; i < diff; i++ {
			p = "0" + p
		}
	}

	// make sure the len of res = 33
	result := q + d + p + string(l.VAT[0])
	log.Printf("Item: %s\n", result)
	return result
}
