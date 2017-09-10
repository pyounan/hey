package helpers

import (
	"encoding/json"
	"net/http"
)

// ReturnSuccessMessage encodes the passed msg to json,
// then returns a json response with status_code 200
func ReturnSuccessMessage(w http.ResponseWriter, msg interface{}) {
	json.NewEncoder(w).Encode(msg)
}

// ReturnErrorMessage encodes the passed msg to json,
// then returns a json error response with status_code 500
func ReturnErrorMessage(w http.ResponseWriter, msg interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(msg)
}

// ReturnErrorMessageWithStatus encodes the passed msg to json,
// then returns a json error response with the passed status code
func ReturnErrorMessageWithStatus(w http.ResponseWriter, status int, msg interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(msg)
}
