package models

import (
	"encoding/xml"
)

type PostRequest struct {
	XMLName             xml.Name `xml:"PostRequest"`
	RoomNumber          string   `xml:"RoomNumber,attr"`
	ReservationId       string   `xml:"ReservationId,attr"`
	LastName            string   `xml:"LastName,attr"`
	RequestType         int      `xml:"RequestType,attr"`
	InquiryInformation  string   `xml:"InquiryInformation,attr"`
	MatchfromPostList   int      `xml:"MatchfromPostList,attr"`
	SequenceNumber      int      `xml:"SequenceNumber,attr"`
	TotalAmount         float32  `xml:"Totalamount,attr"`
	CreditLimitOverride string   `xml:"CreditLimitOverride,attr"`
	PaymentMethod       int      `xml:"PaymentMethod,attr"`
	Covers              int      `xml:"Covers,attr"`
	RevenueCenter       int      `xml:"RevenueCenter,attr"`
	ServingTime         int      `xml:"ServingTime,attr"`
	CheckNumber         string   `xml:"CheckNumber,attr"`
	Subtotal1           float32  `xml:"Subtotal1,attr"`
	Subtotal2           float32  `xml:"Subtotal2,attr"`
	Subtotal3           float32  `xml:"Subtotal1,attr"`
	Subtotal4           float32  `xml:"Subtotal4,attr"`
	Discount1           float32  `xml:"Discount1,attr"`
	Discount2           float32  `xml:"Discount2,attr"`
	Discount3           float32  `xml:"Discount1,attr"`
	Discount4           float32  `xml:"Discount4,attr"`
	Tip                 float32  `xml:"Tip,attr"`
	ServiceCharge1      float32  `xml:"ServiceCharge1,attr"`
	ServiceCharge2      float32  `xml:"ServiceCharge2,attr"`
	ServiceCharge3      float32  `xml:"ServiceCharge1,attr"`
	ServiceCharge4      float32  `xml:"ServiceCharge4,attr"`
	Tax1                float32  `xml:"Tax1,attr"`
	Tax2                float32  `xml:"Tax2,attr"`
	Tax3                float32  `xml:"Tax1,attr"`
	Tax4                float32  `xml:"Tax4,attr"`
	Date                string   `xml:"Date,attr"`
	Time                string   `xml:"Time,attr"`
	WorkstationId       string   `xml:"WorkstationId,attr"`
}

type PostAnswer struct {
	XMLName        xml.Name `xml:"PostAnswer"`
	RoomNumber     string   `xml:"RoomNumber,attr"`
	ReservationId  string   `xml:"ReservationId,attr"`
	LastName       string   `xml:"LastName,attr"`
	AnswerStatus   string   `xml:"AnswerStatus,attr"`
	ResponseText   string   `xml:"ResponseTest,attr"`
	CheckNumber    string   `xml:"CheckNumber,attr"`
	SequenceNumber int      `xml:"SequenceNumber,attr"`
	Date           string   `xml:"Date,attr"`
	Time           string   `xml:"Time,attr"`
	WorkstationId  string   `xml:"WorkstationId,attr"`
	RevenueCenter  int      `xml:"RevenueCenter,attr"`
	PaymentMethod  int      `xml:"PaymentMethod,attr"`
}
