package pos

import (
	"log"
	"net/http"
	"pos-proxy/helpers"
	"pos-proxy/pos/fdm"

	"github.com/gorilla/mux"
)

// FDMStatus returns an fdm response for a certain rcrs number
func FDMStatus(w http.ResponseWriter, r *http.Request) {
	rcrs := mux.Vars(r)["rcrs"]
	conn, err := fdm.Connect(rcrs)
	if err != nil {
		log.Println(err)
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer conn.Close()
	resp, err := fdm.CheckStatus(conn, rcrs)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, resp)
}
