package opera

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
	TotalAmount         float64  `xml:"Totalamount,attr"`
	CreditLimitOverride string   `xml:"CreditLimitOverride,attr"`
	PaymentMethod       int      `xml:"PaymentMethod,attr"`
	Covers              int      `xml:"Covers,attr"`
	RevenueCenter       int      `xml:"RevenueCenter,attr"`
	ServingTime         int      `xml:"ServingTime,attr"`
	CheckNumber         string   `xml:"CheckNumber,attr"`
	Subtotal1           float64  `xml:"Subtotal1,attr"`
	Subtotal2           float64  `xml:"Subtotal2,attr"`
	Subtotal3           float64  `xml:"Subtotal3,attr"`
	Subtotal4           float64  `xml:"Subtotal4,attr"`
	Discount1           float64  `xml:"Discount1,attr"`
	Discount2           float64  `xml:"Discount2,attr"`
	Discount3           float64  `xml:"Discount3,attr"`
	Discount4           float64  `xml:"Discount4,attr"`
	Tip                 float64  `xml:"Tip,attr"`
	ServiceCharge1      float64  `xml:"ServiceCharge1,attr"`
	ServiceCharge2      float64  `xml:"ServiceCharge2,attr"`
	ServiceCharge3      float64  `xml:"ServiceCharge3,attr"`
	ServiceCharge4      float64  `xml:"ServiceCharge4,attr"`
	Tax1                float64  `xml:"Tax1,attr"`
	Tax2                float64  `xml:"Tax2,attr"`
	Tax3                float64  `xml:"Tax3,attr"`
	Tax4                float64  `xml:"Tax4,attr"`
	Date                string   `xml:"Date,attr"`
	Time                string   `xml:"Time,attr"`
	WorkstationId       string   `xml:"WorkstationId,attr"`
}

type PostAnswer struct {
	XMLName        xml.Name `xml:"PostAnswer"`
	RoomNumber     string   `xml:"RoomNumber,attr,omitempty" json:"room_number,omitempty"`
	ReservationId  string   `xml:"ReservationId,attr,omitempty" json:"reservation_id,omitempty"`
	LastName       string   `xml:"LastName,attr,omitempty" json:"last_name,omitempty"`
	AnswerStatus   string   `xml:"AnswerStatus,attr" json:"answer_status"`
	ResponseText   string   `xml:"ResponseTest,attr" json:"response_text"`
	CheckNumber    string   `xml:"CheckNumber,attr,omitempty" json:"check_number,omitempty"`
	SequenceNumber int      `xml:"SequenceNumber,attr" json:"squence_number"`
	Date           string   `xml:"Date,attr" json:"date"`
	Time           string   `xml:"Time,attr" json:"time"`
	WorkstationId  string   `xml:"WorkstationId,attr" json:"workstation_id"`
	RevenueCenter  int      `xml:"RevenueCenter,attr" json:"revenue_center"`
	PaymentMethod  int      `xml:"PaymentMethod,attr" json:"payment_method"`
}
