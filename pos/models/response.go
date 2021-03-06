package models

import (
	"pos-proxy/config"
	"strconv"
	"time"
)

// FDMResponse swagger:model fdmResponse
// defines the body of a response coming from FDM
type FDMResponse struct {
	Identifier         string    `json:"identifier"`
	Sequence           int       `json:"sequence"`
	Retry              int       `json:"retry"`
	Error1             string    `json:"error1"`
	Error2             string    `json:"error2"`
	Error3             string    `json:"error3"`
	ProductionNumber   string    `json:"production_number" bson:"production_number"`
	VSC                string    `json:"vsc"`
	Date               time.Time `json:"date"`
	TimePeriod         time.Time `json:"time_period" bson:"time_period"`
	EventLabel         string    `json:"event_label" bson:"event_label"`
	TicketCounter      string    `json:"ticket_counter" bson:"ticket_counter"`
	TotalTicketCounter string    `json:"total_ticket_counter" bson:"total_ticket_counter"`
	Signature          string    `json:"signature"`
	TicketNumber       string    `json:"ticket_number" bson:"ticket_number"`
	TicketActionTime   string    `json:"ticket_datetime" bson:"ticket_datetime"`
	SoftwareVersion    string    `json:"software_version" bson:"software_version"`
	// Attached attributes from ticket
	PLUHash    string                `json:"plu_hash" bson:"plu_hash"`
	VATSummary map[string]VATSummary `json:"vat_summary" bson:"vat_summary"`
	// FDM Warning
	HasWarning bool   `json:"has_warning" bson:"has_warning"`
	Warning    string `json:"warning" bson:"warning"`
}

func (r *FDMResponse) ProcessStatus(fdm_response []byte) {
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

func (r *FDMResponse) Process(fdm_response []byte, ticket FDMTicket) map[string]interface{} {
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
	r.TicketNumber = ticket.TicketNumber
	r.TicketActionTime = ticket.ActionTime
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
	res["ticket_number"] = r.TicketNumber
	res["ticket_datetime"] = r.TicketActionTime
	res["vat_summary"] = r.VATSummary
	res["software_version"] = config.Version
	return res
}
