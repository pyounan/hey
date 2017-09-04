package helpers

import (
	"net/http"
	"pos-proxy/config"
)

func PrepareRequestHeaders(req *http.Request) *http.Request {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Del("Authorization")
	req.SetBasicAuth(config.AuthUsername, config.AuthPassword)
	return req
}
