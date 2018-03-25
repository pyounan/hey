package payment

import "pos-proxy/payment/gateways/ccv"

var gateways map[string]PaymentGateway

type PaymentGateway interface {
	Sale(data interface{})
	Refund()
	Abort()
	Reprint()
	Output(data interface{})
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
