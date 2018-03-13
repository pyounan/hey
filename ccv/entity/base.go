package entity

import "encoding/xml"

type CardServiceRequest struct {
	XMLName       xml.Name `xml:"CardServiceRequest"`
	WorkstationID string
	RequestID     string
	RequestType   string
	POSdata       POSdata
}

type CardServiceResponse struct {
	XMLName       xml.Name `xml:"CardServiceRequest"`
	WorkstationID string
	RequestID     string
	RequestType   string
	OverallResult string
}

type DeviceRequest struct {
	ID int
}

type DeviceResponse struct {
}

type POSdata struct {
	XMLName        xml.Name `xml:"POSdata"`
	LanguageCode   string   `xml:",attr"`
	POSTimeStamp   string
	ClerkID        int
	PrinterStatus  string
	EJournalStatus string `xml:"E-JournalStatus"`
}
