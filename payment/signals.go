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

var InputSignal = make(chan socket.Event)
var outputSignal = make(chan socket.Event)

func init() {
	socket.Register("payment", InputSignal)
	go handleInputSignals()
	go handleOutputSignals()
}

// PaymentPayload define the data container for a payment-related operation.
type PaymentPayload struct {
	Gateway string          `json:"gateway"`
	Action  string          `json:"action"`
	Data    json.RawMessage `json:"data"`
}

func handleInputSignals() {
	log.Println("waiting for incoming payment events")
	for event := range InputSignal {
		log.Println("a new socket event was received, processing...")
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
			switch payload.Action {

			case "sale":
				log.Println("received a sale request from web socket")
				v.Sale(payload.Data)
			case "reprint":
				v.Reprint()
			case "abort":
				v.Abort()
			case "refund":
				v.Refund(payload.Data)
			case "cancel":
				v.Cancel(payload.Data)
			}
		} else {
			e := socket.Event{
				Module: "payment",
				Type:   "error",
			}
			body := map[string]string{"error": "This payment module is not registered"}
			payload, _ := json.Marshal(body)
			e.Payload = payload
			socket.Send(e)
		}
	}
}

func handleOutputSignals() {
	for v := range outputSignal {
		socket.Send(v)
	}
}
