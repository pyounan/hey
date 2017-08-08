package auth

import (
	"net/http"
	"pos-proxy/helpers"
)

func EnsureTenantSelected(w http.ResponseWriter, req *http.Request) {
	helpers.ReturnSuccessMessage(w, true)
}
