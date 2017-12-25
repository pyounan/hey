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
