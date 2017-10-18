package opera

import (
	"net/http"
	"pos-proxy/helpers"
	"strconv"

	"bytes"
	"encoding/xml"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"log"
	"pos-proxy/db"
	"time"
)

// ListOperaRooms returns a list of reservations on Opera
func ListOperaRooms(w http.ResponseWriter, r *http.Request) {
	postInquiry := PostInquiry{}

	urlQuery := r.URL.Query()
	if _, ok := urlQuery["room_number"]; ok {
		postInquiry.InquiryInformation = urlQuery["room_number"][0]
	}
	if _, ok := urlQuery["terminal"]; ok {
		postInquiry.WorkstationId = urlQuery["terminal"][0]
	}
	if _, ok := urlQuery["store"]; ok {
		postInquiry.RevenueCenter, _ = strconv.Atoi(urlQuery["store"][0])
	}
	postInquiry.MaximumReturnedMatches = 16
	postInquiry.SequenceNumber = 0
	postInquiry.RequestType = 4
	postInquiry.PaymentMethod = 16
	t := time.Now()
	val := fmt.Sprintf("%02d%02d%02d", t.Year(), t.Month(), t.Day())
	val = val[2:]
	postInquiry.Date = val

	val = fmt.Sprintf("%02d%02d%02d", t.Hour(), t.Minute(), t.Second())
	postInquiry.Time = val

	buf := bytes.NewBufferString("")
	if err := xml.NewEncoder(buf).Encode(postInquiry); err != nil {
		log.Println("Error occurred while encoding ", err)
		helpers.ReturnErrorMessage(w, err)
		return
	}
	response, err := SendRequest([]byte(buf.String()))
	response = response[1 : len(response)-1]
	log.Printf("New response '%s'\n", response)
	if err != nil {
		log.Println("Error while sending request", err)
		helpers.ReturnErrorMessage(w, err)
		return
	}
	postList := PostList{}
	responseBuf := bytes.NewBufferString(response)
	if err := xml.NewDecoder(responseBuf).Decode(&postList); err != nil {
		postAnswer := PostAnswer{}
		responseBuf = bytes.NewBufferString(response)
		if err := xml.NewDecoder(responseBuf).Decode(&postAnswer); err != nil {
			log.Println("Couldn't parse as PostAnswer", err)
			helpers.ReturnErrorMessage(w, err)
			return
		}
		log.Println("Writing postAnswer", postAnswer)
		helpers.ReturnErrorMessage(w, postAnswer)
		return
	}
	helpers.ReturnSuccessMessage(w, postList)
}

// Return configured room department ID
func GetRoomDepartment(w http.ResponseWriter, r *http.Request) {
	var roomDepartment RoomDepartmentConfig
	err := db.DB.C("operasettings").Find(bson.M{"config_name": "room_department"}).One(&roomDepartment)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	helpers.ReturnSuccessMessage(w, roomDepartment.Value)
}
