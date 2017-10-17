package opera

import (
	"bytes"
	"encoding/xml"
	"pos-proxy/opera/models"
	"testing"
)

func TestSerializerPostInquiry(t *testing.T) {
	postInquiry := models.PostInquiry{}
	postInquiry.InquiryInformation = "217"
	postInquiry.MaximumReturnedMatches = 16
	postInquiry.SequenceNumber = 1
	postInquiry.RequestType = 4
	postInquiry.PaymentMethod = 16
	postInquiry.Date = "070905"
	postInquiry.Time = "194121"
	postInquiry.RevenueCenter = 1
	postInquiry.WorkstationId = "POS1"

	buf := bytes.NewBufferString("")
	if err := xml.NewEncoder(buf).Encode(postInquiry); err != nil {
		t.Error(err)

	}
	t.Log(buf.String())

}

func TestDeserializerPostInquiry(t *testing.T) {
	postInquiry := models.PostInquiry{}
	xmlStr := `<PostInquiry InquiryInformation="217" 
	MaximumReturnedMatches="16" SequenceNumber="1" RequestType="4" 
	PaymentMethod="16" Date="070905" Time="194121" 
	RevenueCenter="1" WorkstationId="POS1"></PostInquiry>`
	buf := bytes.NewBufferString(xmlStr)
	if err := xml.NewDecoder(buf).Decode(&postInquiry); err != nil {
		t.Error(err)
	}
	t.Log(postInquiry.PaymentMethod)
	t.Log(postInquiry.WorkstationId)
}

func TestSerializerPostList(t *testing.T) {
	postList := models.PostList{}
	postList.SequenceNumber = 1
	for i := 0; i < 5; i++ {
		postListItem := models.PostListItem{}
		postListItem.RoomNumber = "217"
		postListItem.ReservationId = "24331"
		postListItem.LastName = "Hundt"
		postListItem.FirstName = "Heike"
		postList.PostListItems = append(postList.PostListItems, postListItem)
	}
	buf := bytes.NewBufferString("")
	if err := xml.NewEncoder(buf).Encode(postList); err != nil {
		t.Error(err)
	}
	t.Log(buf.String())

}

func TestDeserializerPostList(t *testing.T) {
	postList := models.PostList{}
	xmlStr := `
	<PostList SequenceNumber="1">
		<PostListItem RoomNumber="217" ReservationId="24331" FirstName="Heike" LastName="Hundt"></PostListItem>
		<PostListItem RoomNumber="217" ReservationId="24331" FirstName="Heike" LastName="Hundt"></PostListItem>
		<PostListItem RoomNumber="217" ReservationId="24331" FirstName="Heike" LastName="Hundt"></PostListItem>
		<PostListItem RoomNumber="217" ReservationId="24331" FirstName="Heike" LastName="Hundt"></PostListItem>
		<PostListItem RoomNumber="217" ReservationId="24331" FirstName="Heike" LastName="Hundt"></PostListItem>
	</PostList>
	`
	buf := bytes.NewBufferString(xmlStr)
	if err := xml.NewDecoder(buf).Decode(&postList); err != nil {
		t.Error(err)
	}
	t.Log(postList)
}
