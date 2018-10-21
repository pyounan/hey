package payment

import (
	"encoding/json"
	"pos-proxy/payment/gateways/ccv"
)

var gateways map[string]PaymentGateway

type PaymentGateway interface {
	Sale(data json.RawMessage)
	Refund(data json.RawMessage)
	Abort()
	Reprint(data json.RawMessage)
	Cancel(data json.RawMessage)
}

func init() {
	gateways = make(map[string]PaymentGateway)
	// register gateways
	ccvGW := ccv.New(outputSignal)
	RegisterGateway("ccv", ccvGW)
}

func RegisterGateway(gateway string, v PaymentGateway) {
	gateways[gateway] = v
}
