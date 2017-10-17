package opera

import (
	"encoding/xml"
)

type PostInquiry struct {
	XMLName                xml.Name `xml:"PostInquiry"`
	InquiryInformation     string   `xml:"InquiryInformation,attr"`
	MaximumReturnedMatches int      `xml:"MaximumReturnedMatches,attr"`
	SequenceNumber         int      `xml:"SequenceNumber,attr"`
	RequestType            int      `xml:"RequestType,attr"`
	PaymentMethod          int      `xml:"PaymentMethod,attr"`
	Date                   string   `xml:"Date,attr"`
	Time                   string   `xml:"Time,attr"`
	RevenueCenter          int      `xml:"RevenueCenter,attr"`
	WorkstationId          string   `xml:"WorkstationId,attr"`
}

type PostListItem struct {
	XMLName       xml.Name `xml:"PostListItem"`
	RoomNumber    string   `xml:"RoomNumber,attr" json:"room_number"`
	ReservationId string   `xml:"ReservationId,attr" json:"reservation_id"`
	FirstName     string   `xml:"FirstName,attr" json:"first_name"`
	LastName      string   `xml:"LastName,attr" json:"last_name"`
}

type PostList struct {
	XMLName        xml.Name       `xml:"PostList"`
	SequenceNumber int            `xml:"SequenceNumber,attr" json:"sequence_numner"`
	PostListItems  []PostListItem `xml:"PostListItem" json:"post_list_items"`
}
