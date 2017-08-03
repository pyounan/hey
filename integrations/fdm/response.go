package fdm

import (
	"encoding/json"
	"net/http"
	"pos-proxy/integrations/fdm"
	"strconv"
	"time"
	"errors"
)

type VATSummary map[string]float64

type Response struct {
	Identifier         string    `json:"identifier"`
	Sequence           int       `json:"sequence"`
	Retry              int       `json:"retry"`
	Error1             string    `json:"error1"`
	Error2             string    `json:"error2"`
	Error3             string    `json:"error3"`
	ProductionNumber   string    `json:"production_number"`
	VSC                string    `json:"vsc"`
	Date               time.Time `json:"date"`
	TimePeriod         time.Time `json:"time_period"`
	EventLabel         string    `json:"event_label"`
	TicketCounter      string    `json:"ticket_counter"`
	TotalTicketCounter string    `json:"total_ticket_counter"`
	Signature          string    `json:"signature"`
	// Attached attributes from ticket
	PLUHash    string                `json:"plu_hash"`
	VATSummary map[string]VATSummary `json:"vat_summary"`
}

func (r *Response) ProcessStatus(fdm_response []byte) {
	str := string(fdm_response[:])
	r.Identifier = str[:1]
	n, _ := strconv.Atoi(str[1:3])
	r.Sequence = n

	n, _ = strconv.Atoi(str[3:4])
	r.Retry = n

	r.Error1 = str[4:5]
	r.Error2 = str[5:7]
	r.Error3 = str[7:10]

	r.ProductionNumber = str[10:21]
}

func (r *Response) Process(fdm_response []byte, ticket Ticket) map[string]interface{} {
	str := string(fdm_response[:])

	r.Identifier = str[:1]
	n, _ := strconv.Atoi(str[1:3])
	r.Sequence = n

	n, _ = strconv.Atoi(str[3:4])
	r.Retry = n

	r.Error1 = str[4:5]
	r.Error2 = str[5:7]
	r.Error3 = str[7:10]

	r.ProductionNumber = str[10:21]

	r.VSC = str[21:35]

	t, _ := time.Parse("20060102", str[35:43])
	r.Date = t
	t, _ = time.Parse("150405", str[43:49])
	r.TimePeriod = t

	r.EventLabel = str[49:51]
	r.TicketCounter = str[51:60]
	r.TotalTicketCounter = str[60:69]

	r.Signature = str[69:109]
	// Attaching other attributes from ticket and summarized data
	r.PLUHash = ticket.PLUHash
	r.VATSummary = SummarizeVAT(&ticket.Items)
	// make map
	res := make(map[string]interface{})
	res["identifier"] = r.Identifier
	res["sequence"] = r.Sequence
	res["retry"] = r.Retry
	res["error1"] = r.Error1
	res["error2"] = r.Error2
	res["error3"] = r.Error3
	res["production_number"] = r.ProductionNumber
	res["vsc"] = r.VSC
	res["date"] = r.Date
	res["time_period"] = r.TimePeriod
	res["ticket_counter"] = r.TicketCounter
	res["total_ticket_counter"] = r.TotalTicketCounter
	res["event_label"] = r.EventLabel
	res["signature"] = r.Signature
	return res
}

type FDMErrors map[string]map[string]string
var fdmErrors = FDMErrors{}
fdmErrors["01"] = make(map[string]string)
fdmErrors["01"]["01"] = "FDM Data Storage 90% Full"
fdmErrors["01"]["02"] = "Request Already Answered"
fdmErrors["01"]["03"] = "No Record"
fdmErrors["02"]["01"] = "No VSC or faulty VSC"
fdmErrors["02"]["02"] = "VSC not initialized with pin"
fdmErrors["02"]["03"] = "VSC locked"
fdmErrors["02"]["04"] = "PIN not valid"
fdmErrors["02"]["05"] = "FDM Data Storage Full"
fdmErrors["02"]["06"] = "Unkown message identifier"
fdmErrors["02"]["07"] = "Invalid data in message"
fdmErrors["02"]["08"] = "FDM not operational"
fdmErrors["02"]["09"] = "FDM realtime clock corrupted"
fdmErrors["02"]["10"] = "VSC version not supported by FDM"
fdmErrors["02"]["11"] = "Port 4 not ready"

func CheckError(r Response) error {
	if res.Error1 != "0" {
		return errors.New(fdmErrors[res.Error1][res.Error2])
	}
	return nil
}
