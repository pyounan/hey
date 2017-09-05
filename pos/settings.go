package pos

import (
	"net/http"
	"pos-proxy/config"
	"pos-proxy/helpers"
	"gopkg.in/mgo.v2/bson"
)

func IsFDMEnabled(w http.ResponseWriter, req *http.Request) {
	data := bson.M{"data": config.Config.IsFDMEnabled}
	helpers.ReturnSuccessMessage(w, data)
}
