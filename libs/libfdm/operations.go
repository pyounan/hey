package libfdm

import "errors"

// SetPinResponse defines the data types of the P message
type SetPinResponse struct {
	BaseResponse
	VSCID string
}

// Parse parses the SetPin response and set the object values
func (r *SetPinResponse) Parse(data []byte) {
	str := string(data[:])
	r.Identifier = str[:1]
	r.Sequence = str[1:3]

	r.Retry = str[3:4]

	r.Error1 = str[4:5]
	r.Error2 = str[5:7]
	r.Error3 = str[7:10]

	r.FDMProductionNumber = str[10:21]
	r.VSCID = str[21:35]
}

// CheckErrors reads the errors part of an FDM response and
// returns error mesasge if any.
func (r *SetPinResponse) CheckErrors() error {
	if r.Error1 != "0" {
		return errors.New(fdmErrors[r.Error1][r.Error2])
	}
	return nil
}

// SetPin sends a P request to the FDM to set pin for the VSC
func SetPin(conn *FDM, sequence int, pin string) (SetPinResponse, error) {
	resp := SetPinResponse{}
	identifier := "P"
	retry := "0"
	seq := FormatSequenceNumber(sequence)
	msg := identifier + seq + retry + pin
	data, err := conn.Write(msg, false, 35)
	if err != nil {
		return resp, err
	}
	resp.Parse(data)
	err = resp.CheckErrors()
	return resp, err
}

// IdentificationResponse defines the data types of the I message
type IdentificationResponse struct {
	BaseResponse
	FDMFirmware              string `json:"fdm_firmware"`
	FDMCommunicationProtocol string `json:"fdm_communication_protocol"`
	VSCID                    string `json:"vsc_id"`
	VSCVersion               string `json:"vsc_version"`
}

// Parse parses the Identification response and set the object values
func (r *IdentificationResponse) Parse(data []byte) {
	str := string(data[:])
	r.Identifier = str[:1]
	r.Sequence = str[1:3]

	r.Retry = str[3:4]

	r.Error1 = str[4:5]
	r.Error2 = str[5:7]
	r.Error3 = str[7:10]

	r.FDMProductionNumber = str[10:21]
	r.FDMFirmware = str[21:41]
	r.FDMCommunicationProtocol = str[41:42]
	r.VSCID = str[42:56]
	r.VSCVersion = str[56:59]
}

// CheckErrors reads the errors part of an FDM response and
// returns error mesasge if any.
func (r *IdentificationResponse) CheckErrors() error {
	if r.Error1 != "0" {
		return errors.New(fdmErrors[r.Error1][r.Error2])
	}
	return nil
}

// Identification sends an I request to the FDM to get FDM Identification information
func Identification(conn *FDM, sequence int) (IdentificationResponse, error) {
	resp := IdentificationResponse{}
	identifier := "I"
	retry := "0"
	seq := FormatSequenceNumber(sequence)
	msg := identifier + seq + retry
	data, err := conn.Write(msg, false, 59)
	if err != nil {
		return resp, err
	}
	resp.Parse(data)
	err = resp.CheckErrors()
	return resp, err
}
