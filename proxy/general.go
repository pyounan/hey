package proxy

import (
	//"time"
	//"fmt"
	"encoding/json"
	"pos-proxy/config"
	//"pos-proxy/helpers"
	"net/http"
	//"io/ioutil"
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
	r.Header.Del("Origin")
	r.Header.Set("Origin", w.Header().Get("Origin"))
	w.Header().Del("Access-Control-Allow-Origin")
	prox.ServeHTTP(w, r)
}

/*func ProxyToBackend(w http.ResponseWriter, r *http.Request) {
	// backendURI, _ := url.Parse(config.Config.BackendURI)
	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	uri := fmt.Sprintf("%s%s", config.Config.BackendURI, r.RequestURI)
	req, err := http.NewRequest(r.Method, uri, r.Body)
	// req.Host = config.Config.BackendURI
	req = helpers.PrepareRequestHeaders(r)
	resp, err := netClient.Do(req)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		helpers.ReturnErrorMessage(w, err)
		return
	}
	defer resp.Body.Close()
	w.Write(respbody)
}*/
