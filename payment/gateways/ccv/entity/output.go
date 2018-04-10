package entity

import "encoding/xml"

type Output struct {
	XMLName   xml.Name   `xml:"Output"`
	Target    string     `xml:"OutDeviceTarget,attr"`
	OutResult string     `xml:",omitempty,attr"`
	TextLines []TextLine `xml:"TextLine,omitempty"`
	EJournal  *EJournal  `xml:"E-Journal,omitempty"`
}

type TextLine struct {
	XMLName xml.Name `xml:"TextLine"`
	Height  string   `xml:",attr"`
	Width   string   `xml:",attr"`
	Text    string   `xml:",chardata"`
}

type EJournal struct {
	XMLName         xml.Name `xml:"E-Journal,omitempty"`
	ShopInfo        ShopInfo
	TransactionInfo TransactionInfo
	CardInfo        CardInfo
}

type ShopInfo struct {
	XMLName                xml.Name `xml:"ShopInfo"`
	ShopLocation           string
	TerminalIdentifier     string
	MerchantUserIdentifier string
}

type TransactionInfo struct {
	XMLName                xml.Name `xml:"TransactionInfo"`
	TransactionIdentifier  string
	ServiceLabelName       string
	DateAndTime            string
	DetailedAmount         string
	TotalAmount            string
	TransactionResultText  string
	AcquirerIdentifier     string
	ErrorDiagnosisCode     string
	TrxTermTreatmentResult string
}

type CardInfo struct {
	XMLName               xml.Name `xml:"CardInfo"`
	CardLabelName         string
	ApplicationIdentifier string
	IssuerLabelName       string
	CardNumber            string
	CardSequenceNumber    string
	ExpirationDate        string
	CardHolderName        string
	Account               string
}
