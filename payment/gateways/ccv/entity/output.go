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
	XMLName         xml.Name        `xml:"E-Journal,omitempty" json:"-"`
	ShopInfo        ShopInfo        `json:"shop_info"`
	TransactionInfo TransactionInfo `json:"transaction_info"`
	CardInfo        CardInfo        `json:"card_info"`
}

type ShopInfo struct {
	XMLName                xml.Name `xml:"ShopInfo" json:"-"`
	ShopLocation           string   `json:"shop_location"`
	TerminalIdentifier     string   `json:"terminal_identifier"`
	MerchantUserIdentifier string   `json:"merchant_user_identifier"`
}

type TransactionInfo struct {
	XMLName                xml.Name `xml:"TransactionInfo" json:"-"`
	TransactionIdentifier  string   `json:"transaction_identifier"`
	ServiceLabelName       string   `json:"service_label_name"`
	DateAndTime            string   `json:"date_and_time"`
	DetailedAmount         string   `json:"detailed_amount"`
	TotalAmount            string   `json:"total_amount"`
	TransactionResultText  string   `json:"transaction_result_text"`
	AcquirerIdentifier     string   `json:"acquirer_identifier"`
	ErrorDiagnosisCode     string   `json:"error_diagonsis_code"`
	TrxTermTreatmentResult string   `json:"trx_term_treatment_result"`
}

type CardInfo struct {
	XMLName               xml.Name `xml:"CardInfo" json:"-"`
	CardLabelName         string   `json:"card_label_name"`
	ApplicationIdentifier string   `json:"application_idenfitifer"`
	IssuerLabelName       string   `json:"issuer_label_name"`
	CardNumber            string   `json:"card_number"`
	CardSequenceNumber    string   `json:"card_sequence_number"`
	ExpirationDate        string   `json:"expiration_date"`
	CardHolderName        string   `json:"card_holder_name"`
	Account               string   `json:"account"`
}
