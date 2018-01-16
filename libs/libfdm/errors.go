package libfdm

// FDMErrors represents a type for FDM errors map
type FDMErrors map[string]map[string]string

var fdmErrors = FDMErrors{}

func init() {
	fdmErrors["1"] = make(map[string]string, 3)
	fdmErrors["1"]["01"] = "FDM Data Storage 90% Full"
	fdmErrors["1"]["02"] = "Request Already Answered"
	fdmErrors["1"]["03"] = "No Record"
	fdmErrors["2"] = make(map[string]string, 11)
	fdmErrors["2"]["01"] = "No VSC or faulty VSC"
	fdmErrors["2"]["02"] = "VSC not initialized with pin"
	fdmErrors["2"]["03"] = "VSC locked"
	fdmErrors["2"]["04"] = "PIN not valid"
	fdmErrors["2"]["05"] = "FDM Data Storage Full"
	fdmErrors["2"]["06"] = "Unkown message identifier"
	fdmErrors["2"]["07"] = "Invalid data in message"
	fdmErrors["2"]["08"] = "FDM not operational"
	fdmErrors["2"]["09"] = "FDM realtime clock corrupted"
	fdmErrors["2"]["10"] = "VSC version not supported by FDM"
	fdmErrors["2"]["11"] = "Port 4 not ready"
}

type FDMResponse interface {
	Parse(data []byte) FDMResponse
	CheckErrors() error
	CheckWarning() error
}

// BaseResponse defines the shared part of the FDM response for all messages
type BaseResponse struct {
	Identifier          string
	Sequence            string
	Retry               string
	Error1              string
	Error2              string
	Error3              string
	FDMProductionNumber string
	HasWarning          bool
	Warning             string
}
