package auth

import (
	"net/http"
	"pos-proxy/handlers"
)

func EnsureTenantSelected(w http.ResponseWriter, req *http.Request) {
	handlers.ReturnJSONMessage(w, true)
}
