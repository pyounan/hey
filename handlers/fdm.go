package handlers

import (
	"encoding/json"
	"net/http"
	"pos-proxy/fdm"
)

func FDMStatus(w http.ResponseWriter, r *http.Request) {
	f, err := fdm.New()
	if err != nil {
		json.NewEncoder(w).Encode(err)
		return
	}
	ready, err := f.CheckStatus()
	if err != nil {
		json.NewEncoder(w).Encode(err)
		return
	}
	json.NewEncoder(w).Encode(ready)
}
