package entity

import (
	"encoding/xml"
)

func NewSaleRequest() *SaleRequest {
	r := SaleRequest{}
	r.RequestType = "CardPayment"
	r.XMLNS = "http://www.nrf-arts.org/IXRetail/namespace"
	r.POSdata.POSTimeStamp = "2018-03-14T10:44:58.3913175-07:00" // time.Now().String()
	r.POSdata.LanguageCode = "en"
	return &r
}

type SaleRequest struct {
	CardServiceRequest
	TotalAmount *TotalAmount
}

type SaleResponse struct {
	CardServiceResponse
	Terminal Terminal
	Tender   Tender
}

type TotalAmount struct {
	XMLName  xml.Name `xml:"TotalAmount,omitempty"`
	Currency string   `xml:",attr"`
	Amount   string   `xml:",chardata"`
}

type Terminal struct {
	XMLName    xml.Name
	TerminalID string `xml:",attr"`
	STAN       string `xml:",attr"`
}

type Tender struct {
	XMLName       xml.Name
	Language      string `xml:",attr"`
	TotalAmount   TotalAmount
	Authorisation Authorisation
}

type Authorisation struct {
	XMLName          xml.Name
	TimeStamp        string `xml:",attr"`
	ApprovalCode     string `xml:",attr"`
	AcquireBatch     int    `xml:",attr"`
	CardCircuit      string `xml:",attr"`
	ReceiptCopies    int    `xml:",attr"`
	MaskedCardNumber string `xml:",attr"`
}
