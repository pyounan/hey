package fdm

import (
	"log"
	"strconv"
	"time"
)

type FDMResponse interface {
	Process(res []byte) map[string]interface{}
}

type ProformaResponse struct {
	Identifier         string
	Sequence           int
	Retry              int
	Error1             string
	Error2             string
	Error3             string
	ProductionNumber   string
	VSC                string
	Date               time.Time
	TimePeriod         time.Time
	EventLabel         string
	TicketCounter      string
	TotalTicketCounter string
	Signature          string
}

func (r *ProformaResponse) Process(fdm_response []byte) map[string]interface{} {
	log.Println("FDM STRING RESPONSE =========>")
	str := string(fdm_response[:])
	log.Println(str)

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

type NormalResponse struct {
	Length int
}

func (r NormalResponse) Process(fdm_response []byte) {

}
