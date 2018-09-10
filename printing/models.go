package printing

import (
	"encoding/xml"
	"pos-proxy/income"
	"pos-proxy/pos/models"
)

//FolioPrint defines the objects and requried data for folio printing
type FolioPrint struct {
	Items          []models.EJEvent `bson:"items"`
	Invoice        models.Invoice   `bson:"invoice"`
	Terminal       models.Terminal  `bson:"termianl"`
	Store          models.Store     `bson:"store"`
	Cashier        income.Cashier   `bson:"cashier"`
	Company        income.Company   `bson:"company"`
	Printer        models.Printer   `bson:"printer"`
	TotalDiscounts float64          `bson:"total_discounts"`
	Timezone       string           `bson:"timezone"`
}

//KitchenPrint defines the objects and requried data for kitchen printing
type KitchenPrint struct {
	GropLineItems []models.EJEvent `bson:"group_lineitems`
	Invoice       models.Invoice   `bson:"invoice"`
	Printer       models.Printer   `bson:"printer"`
	Cashier       income.Cashier   `bson:"cashier"`
	Timezone      string           `bson:"timezone"`
}

func New() EposPrint {
	req := EposPrint{}
	return req
}

type EposPrint struct {
	XMLName xml.Name `xml:"epos-print"`
	XMLns   string   `xml:"xmlns,attr"`
	Text    []Text   `xml:""`
	Feed    Feed     `xml:"feed"`
	Cut     Cut      `xml:"cut"`
}

type Text struct {
	XMLName     xml.Name `xml:"text"`
	Text        string   `xml:",chardata"`
	Font        string   `xml:"font,attr,omitempty"`
	Align       string   `xml:"align,attr,omitempty"`
	Linespc     string   `xml:"linespc,attr,omitempty"`
	Reverse     string   `xml:"reverse,attr,omitempty"`
	UnderLine   string   `xml:"ul,attr,omitempty"`
	Emphasized  string   `xml:"em,attr,omitempty"`
	Color       string   `xml:"color,attr,omitempty"`
	DoubleWidth string   `xml:"dw,attr,omitempty"`
	DoubleHight string   `xml:"dh,attr,omitempty"`
}

type Cut struct {
	XMLName xml.Name `xml:"cut"`
	Type    string   `xml:"type,attr"`
}

type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Line    string   `xml:"line,attr,omitempty"`
}
