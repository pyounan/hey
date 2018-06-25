package entity

import "encoding/xml"

type Attrs struct {
	WorkstationID string `xml:",attr"`
	RequestID     string `xml:",attr"`
	RequestType   string `xml:",attr"`
	XMLNS         string `xml:"xmlns,attr"`
	OverallResult string `xml:",attr,omitempty"`
}

type CardServiceRequest struct {
	Attrs
	XMLName xml.Name `xml:"CardServiceRequest"`
	POSdata POSdata
}

type CardServiceResponse struct {
	Attrs
	XMLName xml.Name `xml:"CardServiceResponse"`
}

type DeviceRequest struct {
	XMLName xml.Name `xml:"DeviceRequest"`
	Attrs
	Output Output
}

type DeviceResponse struct {
	XMLName xml.Name `xml:"DeviceResponse"`
	Attrs
	Output Output
	Input  Input
}

type POSdata struct {
	XMLName        xml.Name `xml:"POSdata"`
	LanguageCode   string   `xml:",attr"`
	ApprovalCode   string   `xml:",attr,omitempty"`
	POSTimeStamp   string
	ClerkID        int
	ShiftNumber    int
	PrinterStatus  string
	EJournalStatus string `xml:"E-JournalStatus"`
}
