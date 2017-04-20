package handlers

import (
	"encoding/json"
	"net/http"
)

func ProxyTest(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("success")
}
