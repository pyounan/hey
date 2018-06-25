package entity

import "encoding/xml"

type Input struct {
	XMLName    xml.Name `xml:"Input"`
	Target     string   `xml:"InDeviceTarget,attr"`
	InResult   string   `xml:",omitempty,attr"`
	InputValue InputValue
}

type InputValue struct {
	XMLName  xml.Name `xml:"InputValue"`
	InNumber int
}
