package fdm

import (
	"log"
	"strconv"
	"time"
)

type FDMResponse interface {
	Process(res []byte)
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

func (r ProformaResponse) Process(fdm_response []byte) ProformaResponse {
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

	t, _ := time.Parse("20060102", str[35:42])
	r.Date = t
	t, _ = time.Parse("20060102", str[42:48])
	r.TimePeriod = t

	r.TicketCounter = str[48:57]
	r.TotalTicketCounter = str[57:66]
	r.EventLabel = str[66:68]

	r.Signature = str[68:108]
	return r
}

type NormalResponse struct {
	Length int
}

func (r NormalResponse) Process(fdm_response []byte) {

}
