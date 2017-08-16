package pos

import (
	"net/http"
	"pos-proxy/config"
	"pos-proxy/helpers"
)

func IsFDMEnabled(w http.ResponseWriter, req *http.Request) {
	helpers.ReturnSuccessMessage(w, config.Config.IsFDMEnabled)
}
