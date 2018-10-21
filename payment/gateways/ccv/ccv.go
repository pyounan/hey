package ccv

import (
	"encoding/json"
	"log"
	"pos-proxy/db"
	generalEntity "pos-proxy/entity"
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

	type SaleRequest struct {
		Amount            float64 `json:"amount"`
		TerminalID        int     `json:"terminal_id"`
		TerminalNumber    int     `json:"terminal_number"`
		UseDefaultAccount bool    `json:"use_default_account"`
		CashierID         int     `json:"cashier_id"`
		Currency          string  `json:"currency"`
	}
	payload := SaleRequest{}
	// bytes.NewReader([]byte(data))
	err := json.Unmarshal(data, &payload)
	if err != nil {
		log.Println(err)
		return
	}

	// Retrieve CCV Settings for this terminal
	var settings *generalEntity.CCVSettings
	if payload.UseDefaultAccount {
		settings, err = db.GetCCVDefaultAccountSettings()
	} else {
		settings, err = db.GetCCVSettingsForTerminal(payload.TerminalID)
	}
	if err != nil {
		m := socket.Event{}
		m.Module = "payment"
		m.Type = "error"
		payload := make(map[string]string, 1)
		payload["error"] = "This terminal doesn't have any CCV pinpad configured"
		encodedPayload, _ := json.Marshal(payload)
		m.Payload = encodedPayload
		gateway.ouputChannel <- m
		return
	}

	sender.Connect(*settings)
	err = receiver.Listen(settings, gateway.ouputChannel)
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
	cardServiceReq.RequestID = strconv.Itoa(getNextRequestID())
	cardServiceReq.TotalAmount = &entity.TotalAmount{}
	cardServiceReq.TotalAmount.Amount = entity.FloatToString(payload.Amount)
	cardServiceReq.TotalAmount.Currency = payload.Currency
	cardServiceReq.POSdata.PrinterStatus = "Available"
	cardServiceReq.POSdata.EJournalStatus = "Available"
	cardServiceReq.POSdata.ClerkID = payload.CashierID
	cardServiceReq.WorkstationID = strconv.Itoa(payload.TerminalNumber)
	res, err := sender.Send(gateway.ouputChannel, cardServiceReq, *settings)
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
	log.Println(res)
}

func (gateway CCV) Reprint(data json.RawMessage) {
	log.Println("Starting CCV Reprint request")
	type RePrintRequest struct {
		TerminalID        int  `json:"terminal_id"`
		UseDefaultAccount bool `json:"use_default_account"`
	}
	payload := RePrintRequest{}
	err := json.Unmarshal(data, &payload)
	if err != nil {
		log.Println(err)
		return
	}

	// Retrieve CCV Settings for this terminal
	var settings *generalEntity.CCVSettings
	if payload.UseDefaultAccount {
		settings, err = db.GetCCVDefaultAccountSettings()
	} else {
		settings, err = db.GetCCVSettingsForTerminal(payload.TerminalID)
	}
	if err != nil {
		m := socket.Event{}
		m.Module = "payment"
		m.Type = "error"
		payload := make(map[string]string, 1)
		payload["error"] = "This terminal doesn't have any CCV pinpad configured"
		encodedPayload, _ := json.Marshal(payload)
		m.Payload = encodedPayload
		gateway.ouputChannel <- m
		return
	}
	sender.Connect(*settings)
	err = receiver.Listen(settings, gateway.ouputChannel)
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
	res, err := sender.Send(gateway.ouputChannel, cardServiceReq, *settings)
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

	type RefundPayload struct {
		Amount            float64 `json:"amount"`
		TerminalID        int     `json:"terminal_id"`
		TerminalNumber    int     `json:"terminal_number"`
		UseDefaultAccount bool    `json:"use_default_account"`
		CashierID         int     `json:"cashier_id"`
		Currency          string  `json:"currency"`
	}
	payload := RefundPayload{}
	err := json.Unmarshal(data, &payload)
	if err != nil {
		log.Println(err)
		return
	}

	// Retrieve CCV Settings for this terminal
	var settings *generalEntity.CCVSettings
	if payload.UseDefaultAccount {
		settings, err = db.GetCCVDefaultAccountSettings()
	} else {
		settings, err = db.GetCCVSettingsForTerminal(payload.TerminalID)
	}
	if err != nil {
		m := socket.Event{}
		m.Module = "payment"
		m.Type = "error"
		payload := make(map[string]string, 1)
		payload["error"] = "This terminal doesn't have any CCV pinpad configured"
		encodedPayload, _ := json.Marshal(payload)
		m.Payload = encodedPayload
		gateway.ouputChannel <- m
		return
	}

	sender.Connect(*settings)
	err = receiver.Listen(settings, gateway.ouputChannel)
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

	cardServiceReq := entity.NewRefundRequest()
	cardServiceReq.RequestID = strconv.Itoa(getNextRequestID())
	cardServiceReq.TotalAmount = &entity.TotalAmount{}
	cardServiceReq.TotalAmount.Amount = entity.FloatToString(payload.Amount)
	cardServiceReq.TotalAmount.Currency = payload.Currency
	cardServiceReq.POSdata.PrinterStatus = "Available"
	cardServiceReq.POSdata.EJournalStatus = "Available"
	cardServiceReq.POSdata.ClerkID = payload.CashierID
	cardServiceReq.WorkstationID = strconv.Itoa(payload.TerminalNumber)
	res, err := sender.Send(gateway.ouputChannel, cardServiceReq, *settings)
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
	/*sender.Connect("192.168.100.114", "4100")
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
	log.Println(res)*/
}

type CancelPayload struct {
	AuthCode          string  `json:"auth_code"`
	Amount            float64 `json:"amount"`
	TerminalID        int     `json:"terminal_id"`
	TerminalNumber    int     `json:"terminal_number"`
	UseDefaultAccount bool    `json:"use_default_account"`
	CashierID         int     `json:"cashier_id"`
	Currency          string  `json:"currency"`
}

// Cancel initiates cancel operation for a certain Auth Code
func (gateway CCV) Cancel(data json.RawMessage) {
	log.Println("Starting CCV Cancel request")

	payload := CancelPayload{}
	err := json.Unmarshal(data, &payload)
	if err != nil {
		log.Println(err)
		return
	}

	// Retrieve CCV Settings for this terminal
	var settings *generalEntity.CCVSettings
	if payload.UseDefaultAccount {
		settings, err = db.GetCCVDefaultAccountSettings()
	} else {
		settings, err = db.GetCCVSettingsForTerminal(payload.TerminalID)
	}
	if err != nil {
		m := socket.Event{}
		m.Module = "payment"
		m.Type = "error"
		payload := make(map[string]string, 1)
		payload["error"] = "This terminal doesn't have any CCV pinpad configured"
		encodedPayload, _ := json.Marshal(payload)
		m.Payload = encodedPayload
		gateway.ouputChannel <- m
		return
	}

	sender.Connect(*settings)
	err = receiver.Listen(settings, gateway.ouputChannel)
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

	cardServiceReq := entity.NewCancelRequest(payload.AuthCode, payload.Amount, payload.Currency)
	cardServiceReq.RequestID = strconv.Itoa(getNextRequestID())
	cardServiceReq.POSdata.PrinterStatus = "Available"
	cardServiceReq.POSdata.EJournalStatus = "Available"
	cardServiceReq.POSdata.ClerkID = payload.CashierID
	cardServiceReq.WorkstationID = strconv.Itoa(payload.TerminalNumber)
	res, err := sender.Send(gateway.ouputChannel, cardServiceReq, *settings)
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
