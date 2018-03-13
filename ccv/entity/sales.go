package entity

import "encoding/xml"

type SaleRequest struct {
	CardServiceRequest
	TotalAmount TotalAmount
}

type SaleResponse struct {
	CardServiceResponse
	Terminal Terminal
	Tender   Tender
}

type TotalAmount struct {
	XMLName  xml.Name `xml:"TotalAmount"`
	Currency string
	Amount   float64 `xml:",chardata"`
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
