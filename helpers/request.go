package helpers

import (
	"net"
	"net/http"
	"pos-proxy/config"
	"time"
)

func PrepareRequestHeaders(req *http.Request) *http.Request {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Del("Authorization")
	req.SetBasicAuth(config.AuthUsername, config.AuthPassword)
	return req
}

func NewNetClient() *http.Client {
	c := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   120 * time.Second,
				KeepAlive: 120 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 15 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	return c
}
