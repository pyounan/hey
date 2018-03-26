/*Package payment handles events or singals related to payment operations, like Sale
or Refund
Sale example:
{"module": "payment", "payload": {"gateway": "ccv", "action": "sale", "data": {"amount": 50.00}}}
*/
package payment

import (
	"encoding/json"
	"log"
	"pos-proxy/socket"
)

var inputSignal = make(chan socket.Event)
var outputSignal = make(chan socket.Event)

func init() {
	socket.Register("payment", inputSignal)
	go handleInputSignals()
	go handleOutputSignals()
}

type PaymentPayload struct {
	Gateway string      `json:"gateway"`
	Action  string      `json:"action"`
	Data    json.RawMessage `json:"data"`
}

func handleInputSignals() {
	log.Println("waiting for incoming payment events")
	for event := range inputSignal {
		log.Println("i should make payment now !", event.Payload)
		var payload PaymentPayload
		err := json.Unmarshal(event.Payload, &payload)
		if err != nil {
			log.Println(err, "Failed to cast the payload to a payment payload")
			e := socket.Event{
				Module: "payment",
				Type:   "error",
			}
			payload := map[string]string{"error": err.Error()}
			encodedPayload, _ := json.Marshal(payload)
			e.Payload = encodedPayload
			outputSignal <- e
			continue
		}
		if v, ok := gateways[payload.Gateway]; ok {
			log.Println("handling incoming message for ", payload.Gateway)
			// ccvChan := make(chan ccv.Command)
			// go handleCCVSignals(ccvChan)
			if payload.Action == "sale" {
				v.Sale(payload.Data)
			} else if payload.Action == "reprint" {
				v.Reprint()
			} else if event.Type == "output" {
				v.Output(payload.Data)
			}
		}
	}
}

func handleOutputSignals() {
	for v := range outputSignal {
		log.Println("should send this from ccv to socket")
		socket.Send(v)
	}
}

/*
func handleCCVSignals(c chan ccv.Command) {
	for com := range c {
		log.Printf("received command, %#v\n", com)

		m := socket.Message{
			Type:    com.GetType(),
			Payload: com.GetPayload(),
		}
		socket.Output <- m
	}
}
*/
