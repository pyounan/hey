package helpers

import (
	"encoding/json"
	"net/http"
)

func ReturnSuccessMessage(w http.ResponseWriter, msg interface{}) {
	json.NewEncoder(w).Encode(msg)
}

func ReturnErrorMessage(w http.ResponseWriter, msg interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(msg)
}
