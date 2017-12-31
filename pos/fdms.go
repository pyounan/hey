package pos

import (
	"encoding/json"
	"net/http"
	"pos-proxy/helpers"
	"pos-proxy/libs/libfdm"
	"pos-proxy/pos/fdm"
)

// FDMSetPin api returns a json response of FDM SetPIN request
func FDMSetPin(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		ProductionNumber string `json:"production_number"`
		Pin              string `json:"pin"`
	}
	body := reqBody{}
	// read request body
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()
	// create FDM connection
	conn, err := fdm.Connect(body.ProductionNumber)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer conn.Close()
	// get next sequence number for this production number
	sn, err := fdm.GetNextSequence(body.ProductionNumber)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	// send setpin request to FDM
	resp, err := libfdm.SetPin(conn, sn, body.Pin)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, resp)
}
