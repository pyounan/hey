package handlers

import (
	"encoding/json"
	"net/http"
)

// ReturnJSONError return an error in a json response
func ReturnJSONError(w http.ResponseWriter, err string) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(err)
}

// ReturnJSONError return an error in a json response
func ReturnJSONMessage(w http.ResponseWriter, msg interface{}) {
	json.NewEncoder(w).Encode(msg)
}
