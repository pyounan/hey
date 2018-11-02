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

//New returns request of envelope object
func New() Envelope {
	req := Envelope{}
	return req
}

//Envelope defines body of Envelope tag
type Envelope struct {
	XMLName xml.Name `xml:"s:Envelope"`
	XMLns   string   `xml:"xmlns:s,attr"`
	Body    Body     `xml:"s:Body"`
}

//Body defines body of body tag
type Body struct {
	XMLName   xml.Name  `xml:"s:Body"`
	EposPrint EposPrint `xml:"epos-print"`
}

//EposPrint defines body of EposPrint tag
type EposPrint struct {
	XMLName xml.Name `xml:"epos-print"`
	XMLns   string   `xml:"xmlns,attr"`
	Layout  *Layout  `xml:"layout"`
	Align   *Text    `xml:""`
	Image   []Image  `xml:"image,omitempty"`
	Text    []Text   `xml:""`
	Feed    *Feed    `xml:"feed,omitempty"`
	Cut     Cut      `xml:"cut"`
}

//Text defines body of Text tag
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
	Lang        string   `xml:"lang,attr,omitempty"`
}

//Cut defines body of Cut tag
type Cut struct {
	XMLName xml.Name `xml:"cut"`
	Type    string   `xml:"type,attr"`
}

//Feed defines body of Feed tag
type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Line    string   `xml:"line,attr,omitempty"`
}

//Image defines body of Image tag
type Image struct {
	XMLName xml.Name `xml:"image,omitempty"`
	Image   string   `xml:",chardata"`
	Width   string   `xml:"width,attr,omitempty"`
	Height  string   `xml:"height,attr,omitempty"`
	Color   string   `xml:"color,attr,omitempty"`
	Mode    string   `xml:"mode,attr,omitempty"`
}

//Layout defines body of Layout tag
type Layout struct {
	XMLName      xml.Name `xml:"layout"`
	Type         string   `xml:"type,attr,omitempty"`
	Width        string   `xml:"width,attr,omitempty"`
	Height       string   `xml:"height,attr,omitempty"`
	MarginTop    string   `xml:"margin-top,attr,omitempty"`
	MarginBottom string   `xml:"margin-bottom,attr,omitempty"`
	OffsetCut    string   `xml:"offset-cut,attr,omitempty"`
	OffsetLabel  string   `xml:"offset-label,attr,omitempty"`
}
