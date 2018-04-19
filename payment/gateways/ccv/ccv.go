package ccv

import (
	"encoding/json"
	"log"
	"pos-proxy/payment/gateways/ccv/entity"
	"pos-proxy/payment/gateways/ccv/receiver"
	"pos-proxy/payment/gateways/ccv/sender"
	"pos-proxy/socket"
	"strconv"
	"sync"
)

var requestID int
var requestIDMutex = &sync.Mutex{}

func getNextRequestID() int {
	requestIDMutex.Lock()
	defer requestIDMutex.Unlock()
	requestID++
	return requestID
}

type CCV struct {
	ouputChannel chan socket.Event
}

// New creates a new ccv instance and hooks the ouput channel to it
func New(ch chan socket.Event) CCV {
	return CCV{ouputChannel: ch}
}

/*func OutputResponse(event socket.Event) {
	resp := entity.DeviceResponse{}
	resp.OverallResult = "Success"
	resp.Output.OutResult = "Success"
	receiver.Send(&resp)
}*/

// Sale initiates a CardService Requests to CCV Pinpad
func (gateway CCV) Sale(data json.RawMessage) {
	// go handleInternalSignals(notif)
	log.Println("Starting CCV Sale request")
	sender.Connect("192.168.100.114", "4100")
	err := receiver.Listen(":4102", gateway.ouputChannel)
	if err != nil {
		log.Println(err)
		m := socket.Event{}
		m.Module = "payment"
		m.Type = "error"
		payload := make(map[string]string, 1)
		payload["error"] = err.Error()
		encodedPayload, _ := json.Marshal(payload)
		m.Payload = encodedPayload
		gateway.ouputChannel <- m
		return
	}

	type SaleRequest struct {
		Amount float64 `json:"amount"`
	}
	payload := SaleRequest{}
	// bytes.NewReader([]byte(data))
	err = json.Unmarshal(data, &payload)
	if err != nil {
		log.Println(err)
		return
	}

	cardServiceReq := entity.NewSaleRequest()
	cardServiceReq.RequestID = strconv.Itoa(getNextRequestID())
	cardServiceReq.TotalAmount = &entity.TotalAmount{}
	cardServiceReq.TotalAmount.Amount = entity.FloatToString(payload.Amount)
	log.Println("Amount to be sent to pinpad is", cardServiceReq.TotalAmount.Amount)
	cardServiceReq.TotalAmount.Currency = "EUR"
	cardServiceReq.POSdata.PrinterStatus = "Available"
	cardServiceReq.POSdata.EJournalStatus = "Available"
	cardServiceReq.POSdata.ClerkID = 1
	res, err := sender.Send(gateway.ouputChannel, cardServiceReq)
	if err != nil {
		m := socket.Event{}
		m.Module = "payment"
		m.Type = "error"
		payload := make(map[string]string, 1)
		payload["error"] = err.Error()
		encodedPayload, _ := json.Marshal(payload)
		m.Payload = encodedPayload
		gateway.ouputChannel <- m
		return
	}
	log.Println(res)
}

func (gateway CCV) Reprint() {
	log.Println("Starting CCV Reprint request")
	sender.Connect("192.168.100.114", "4100")
	err := receiver.Listen(":4102", gateway.ouputChannel)
	if err != nil {
		log.Println(err)
		m := socket.Event{}
		m.Module = "payment"
		m.Type = "error"
		payload := make(map[string]string, 1)
		payload["error"] = err.Error()
		encodedPayload, _ := json.Marshal(payload)
		m.Payload = encodedPayload
		gateway.ouputChannel <- m
		return
	}

	cardServiceReq := entity.NewSaleRequest()
	cardServiceReq.RequestType = "TicketReprint"
	cardServiceReq.RequestID = strconv.Itoa(getNextRequestID())
	cardServiceReq.POSdata.PrinterStatus = "Available"
	cardServiceReq.POSdata.EJournalStatus = "Available"
	cardServiceReq.POSdata.ClerkID = 1
	res, err := sender.Send(gateway.ouputChannel, cardServiceReq)
	if err != nil {
		m := socket.Event{}
		m.Module = "payment"
		m.Type = "error"
		payload := make(map[string]string, 1)
		payload["error"] = err.Error()
		encodedPayload, _ := json.Marshal(payload)
		m.Payload = encodedPayload
		gateway.ouputChannel <- m
		return
	}
	log.Println(res)

}

func (gateway CCV) Refund(data json.RawMessage) {
	log.Println("Starting CCV Refund request")
	sender.Connect("192.168.100.114", "4100")
	err := receiver.Listen(":4102", gateway.ouputChannel)
	if err != nil {
		log.Println(err)
		m := socket.Event{}
		m.Module = "payment"
		m.Type = "error"
		payload := make(map[string]string, 1)
		payload["error"] = err.Error()
		encodedPayload, _ := json.Marshal(payload)
		m.Payload = encodedPayload
		gateway.ouputChannel <- m
		return
	}

	cardServiceReq := entity.NewSaleRequest()
	cardServiceReq.RequestType = "TicketReprint"
	cardServiceReq.RequestID = strconv.Itoa(getNextRequestID())
	cardServiceReq.POSdata.PrinterStatus = "Available"
	cardServiceReq.POSdata.EJournalStatus = "Available"
	cardServiceReq.POSdata.ClerkID = 1
	res, err := sender.Send(gateway.ouputChannel, cardServiceReq)
	if err != nil {
		m := socket.Event{}
		m.Module = "payment"
		m.Type = "error"
		payload := make(map[string]string, 1)
		payload["error"] = err.Error()
		encodedPayload, _ := json.Marshal(payload)
		m.Payload = encodedPayload
		gateway.ouputChannel <- m
		return
	}
	log.Println(res)
}
func (gateway CCV) Abort() {
	log.Println("Starting CCV Abort request")
	sender.Connect("192.168.100.114", "4100")
	err := receiver.Listen(":4102", gateway.ouputChannel)
	if err != nil {
		log.Println(err)
		m := socket.Event{}
		m.Module = "payment"
		m.Type = "error"
		payload := make(map[string]string, 1)
		payload["error"] = err.Error()
		encodedPayload, _ := json.Marshal(payload)
		m.Payload = encodedPayload
		gateway.ouputChannel <- m
		return
	}

	cardServiceReq := entity.NewSaleRequest()
	cardServiceReq.RequestType = "AbortRequest"
	cardServiceReq.RequestID = strconv.Itoa(getNextRequestID())
	cardServiceReq.POSdata.PrinterStatus = "Available"
	cardServiceReq.POSdata.EJournalStatus = "Available"
	cardServiceReq.POSdata.ClerkID = 1
	res, err := sender.Send(gateway.ouputChannel, cardServiceReq)
	if err != nil {
		m := socket.Event{}
		m.Module = "payment"
		m.Type = "error"
		payload := make(map[string]string, 1)
		payload["error"] = err.Error()
		encodedPayload, _ := json.Marshal(payload)
		m.Payload = encodedPayload
		gateway.ouputChannel <- m
		return
	}
	log.Println(res)
}

func (gateway CCV) Output(data interface{}) {}
