package opera

import (
	"net/http"
	"pos-proxy/helpers"
	"strconv"

	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"time"
)

// ListOperaRooms swagger:route GET /api/opera/rooms/ opera listRooms
//
// List Opera Rooms
//
// returns a list of reservations on Opera
//
// Parameters:
// + name: inquiry
//   in: query
//   required: true
//
// + name: terminal
//   in: query
//   required: true
//
// + name: store
//   in: query
//   required: true
//
// Responses:
// 200: []postListItem
func ListOperaRooms(w http.ResponseWriter, r *http.Request) {
	postInquiry := PostInquiry{}

	urlQuery := r.URL.Query()
	if _, ok := urlQuery["inquiry"]; ok {
		postInquiry.InquiryInformation = urlQuery["inquiry"][0]
	} else {
		helpers.ReturnErrorMessage(w, "Insufficient parameters")
		return
	}
	if _, ok := urlQuery["terminal"]; ok {
		postInquiry.WorkstationId = urlQuery["terminal"][0]
	} else {
		helpers.ReturnErrorMessage(w, "Insufficient parameters")
		return
	}
	if _, ok := urlQuery["store"]; ok {
		postInquiry.RevenueCenter, _ = strconv.Atoi(urlQuery["store"][0])
	} else {
		helpers.ReturnErrorMessage(w, "Insufficient parameters")
		return
	}
	postInquiry.MaximumReturnedMatches = 16
	postInquiry.RequestType = 12
	deptID, _ := GetRoomDepartmentID()
	paymentMethodInt, err := GetPaymentMethod(deptID)
	postInquiry.PaymentMethod = paymentMethodInt
	seqNumber, _ := GetNextSequence()
	postInquiry.SequenceNumber = seqNumber
	t := time.Now()
	val := fmt.Sprintf("%02d%02d%02d", t.Year(), t.Month(), t.Day())
	val = val[2:]
	postInquiry.Date = val

	val = fmt.Sprintf("%02d%02d%02d", t.Hour(), t.Minute(), t.Second())
	postInquiry.Time = val

	buf := bytes.NewBufferString("")
	if err := xml.NewEncoder(buf).Encode(postInquiry); err != nil {
		log.Println("Error occurred while encoding ", err)
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	response, err := SendRequest([]byte(buf.String()))
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	response = response[1 : len(response)-1]
	log.Printf("New response '%s'\n", response)
	if err != nil {
		log.Println("Error while sending request", err)
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	postList := PostList{}
	responseBuf := bytes.NewBufferString(response)
	if err := xml.NewDecoder(responseBuf).Decode(&postList); err != nil {
		postAnswer := PostAnswer{}
		responseBuf = bytes.NewBufferString(response)
		if err := xml.NewDecoder(responseBuf).Decode(&postAnswer); err != nil {
			log.Println("Couldn't parse as PostAnswer", err)
			helpers.ReturnErrorMessage(w, err.Error())
			return
		}
		helpers.ReturnSuccessMessage(w, "[]")
		return
	}
	helpers.ReturnSuccessMessage(w, postList.PostListItems)
}

// roomDepartmentResponse swagger:model roomDepartmentResponse
type roomDepartmentResponse struct {
	DepartmentID int `json:"department_id" bson:"department_id"`
}

// GetRoomDeparment swagger:route GET /api/opera/roomdepartment/ opera getRoomDepartment
//
// Get Room Department
//
// Return configured room department ID
//
// Responses:
// 200: roomDepartmentResponse
func GetRoomDepartment(w http.ResponseWriter, r *http.Request) {
	deptID, err := GetRoomDepartmentID()
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	resp := roomDepartmentResponse{
		DepartmentID: deptID,
	}
	helpers.ReturnSuccessMessage(w, resp)
}
