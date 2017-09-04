package proxy

import (
	"encoding/json"
	"pos-proxy/config"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func Status(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("success")
}

func ProxyToBackend(w http.ResponseWriter, r *http.Request) {
	backendURI, _ := url.Parse(config.Config.BackendURI)
	prox := httputil.NewSingleHostReverseProxy(backendURI)
	r.SetBasicAuth(config.AuthUsername, config.AuthPassword)
	r.Header.Del("Access-Control-Allow-Origin")
	w.Header().Del("Access-Control-Allow-Origin")
	prox.ServeHTTP(w, r)
}
