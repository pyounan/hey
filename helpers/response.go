package helpers

import (
	"encoding/json"
	"log"
	"net/http"
)

func ReturnSuccessMessage(w http.ResponseWriter, msg interface{}) {
	json.NewEncoder(w).Encode(msg)
}

func ReturnErrorMessage(w http.ResponseWriter, msg interface{}) {
	log.Println(msg)
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(msg)
}
