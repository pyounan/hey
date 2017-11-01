package proxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"pos-proxy/config"
	"pos-proxy/helpers"
	//"net/http/httputil"
	//"net/url"
)

// AllowIncomingRequests indicates if the proxy allows to receive operations,
// or all the operations should be halted until an intervertion from support.
var AllowIncomingRequests = true

// StatusMiddleware checks the value of AllowIncomingRequests and determines if the
// ongoing request should be rejected or can continue the operation.
// If AllowIncomingRequests is true, then the proxy is healthy and accepting more
// operations. If false, then the request should return an internal error with a
// message to the client to call suport team.
func StatusMiddleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// If proxy should reject all operations, return error message and don't call next middleware
		if AllowIncomingRequests == false {
			err := map[string]string{"message": "Proxy Internal Error, Operations halted. Please contact support."}
			helpers.ReturnErrorMessage(w, err)
			return
		}
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// Status returns a success message if the proxy is working properly.
func Status(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("success")
}

/*func ProxyToBackend(w http.ResponseWriter, r *http.Request) {
	backendURI, _ := url.Parse(config.Config.BackendURI)
	prox := httputil.NewSingleHostReverseProxy(backendURI)
	r.SetBasicAuth(config.AuthUsername, config.AuthPassword)
	r.Header.Del("Access-Control-Allow-Origin")
	r.Header.Del("Origin")
	r.Header.Set("Origin", "https://test.cloudinn.net")
	w.Header().Del("Access-Control-Allow-Origin")
	prox.ServeHTTP(w, r)
}*/

// ProxyToBackend sends the incoming requests to the backend directly
func ProxyToBackend(w http.ResponseWriter, r *http.Request) {
	// backendURI, _ := url.Parse(config.Config.BackendURI)
	netClient := helpers.NewNetClient()

	uri := fmt.Sprintf("%s%s", config.Config.BackendURI, r.RequestURI)
	req, err := http.NewRequest(r.Method, uri, r.Body)
	// req.Host = config.Config.BackendURI
	req = helpers.PrepareRequestHeaders(req)
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
}
