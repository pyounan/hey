package fdm

import (
	"errors"
	"fmt"
	"log"
	// "time"

	"pos-proxy/config"
	"pos-proxy/db"

	"github.com/tarm/serial"
)

// FDM is a structure the defines the configuration and the port to the fdm
// connection.
type FDM struct {
	RCRS string
	c *serial.Config
	s *serial.Port
}

// New Creates a new fdm connection and returns FDM struct.
func New(RCRS string) (*FDM, error) {
	fdm := &FDM{}
	if RCRS == "" {
		return nil, errors.New("You must specifiy a valid RCRS number")
	}

	// find the FDM that is supposed to receive requests from this RCRS number
	for _, f := range config.Config.FDMs {
		for _, r := range f.RCRS {
			if r == RCRS {
				fdm.c = &serial.Config{Name: f.FDM_Port, Baud: f.FDM_Speed}
				fdm.RCRS = RCRS
				break
			}
		}
	}

	if fdm.c == nil {
		err := errors.New("there is no fdm configuration for this production number")
		return nil, err
	}
	log.Println("Trying to stablish connection with FDM with configuration ->")
	log.Printf("Port: %s", fdm.c.Name)
	s, err := serial.OpenPort(fdm.c)
	fdm.s = s
	if err != nil {
		log.Println("Failed to stablish connection with FDM")
		return nil, err
	}

	log.Println("Connection to Serial Port has been stablished successfully.")
	return fdm, nil
}

// CheckStatus sends S000 to the FDM and check if its ready.
func (fdm *FDM) CheckStatus() (bool, error) {
	n, err := db.GetNextSequence(fdm.RCRS)
	// db.UpdateLastSequence(fdm.RCRS, n)
	if err != nil {
		return false, err
	}
	msg := fmt.Sprintf("S%s0", FormatSequence(n))
	if _, err := fdm.Write(msg, false, 21); err != nil {
		log.Println("Error: ", err)
		return false, err
	}

	log.Println("FDM is ready.")
	return true, nil
}

// SendAndWaitForACK sends a message to the FDM and retries until it recievs ACK.
func (fdm *FDM) SendAndWaitForACK(packet []byte) (bool, error) {
	// if the response is not valid we try to retry reading the answer again
	ack := 0x00
	max_retries := byte('3')
	for packet[4] < max_retries && ack != 0x06 {
		_, err := fdm.s.Write(packet)
		if err != nil {
			return false, err
		}
		res := make([]byte, 1)
		_, err = fdm.s.Read(res)
		if err != nil {
			log.Println(err)
			return false, err
		}
		incrementRetryCounter(packet)
		if res[0] == 0x06 {
			log.Println("ACK received.")
			ack = 0x06
		} else {
			log.Println("ACK wasn't received, retrying...")
		}
	}
	if ack == 0x06 {
		return true, nil
	} else {
		return false, nil
	}
}

// Write writes a message to the fdm, if just_wait_for_ACK is true, then it won't
// wait for the response. IF its false, then the process goes on and process the
// response.
func (fdm *FDM) Write(message string, just_wait_for_ACK bool, response_size int) ([]byte, error) {
	packet := generateLowLevelMessage(message)
	got_response := false
	max_nacks := 2
	sent_nacks := 0
	response := []byte{}

	ok, err := fdm.SendAndWaitForACK(packet)
	if ok == false {
		log.Println(err)
		return response, errors.New("Didn't recieve ACK")
	}
	if just_wait_for_ACK {
		fdm.SendACK()
		return response, nil
	}
	for got_response == false && sent_nacks < max_nacks {
		stx := make([]byte, 1)
		_, err = fdm.s.Read(stx)
		if err != nil {
			log.Println("Error reading stx", stx, err)
			return response, err
		}
		msg := make([]byte, response_size)
		for i := 0; i < response_size; i++ {
			tmp := make([]byte, 1)
			_, err = fdm.s.Read(tmp)
			if err != nil {
				log.Println("Error reading msg", msg, err)
				return response, err
			}
			msg[i] = tmp[0]
		}

		etx := make([]byte, 1)
		_, err = fdm.s.Read(etx)
		if err != nil {
			log.Println("Error reading etx", etx, err)
			return response, err
		}
		bcc := make([]byte, 1)
		_, err = fdm.s.Read(bcc)
		if err != nil {
			log.Println("Error reading bcc", bcc, err)
			return response, err
		}
		// compare results
		// fmt.Printf("FDM RESPONSE: %v, %v, %v\n", stx[0], etx[0], bcc[0])
		if fmt.Sprintf("%v", stx) != fmt.Sprintf("%v", 0x02) && fmt.Sprintf("%v", etx) != fmt.Sprintf("%v", 0x03) && bcc != nil && calculateLRC(msg) == bcc[0] {
			got_response = true
			response = msg
			log.Println("Received a valid response")
			break
		} else {
			log.Println("Not a valid response, sending NACK....")
			response = msg
			sent_nacks += 1
			fdm.SendNACK()
		}
	}

	if got_response == false {
		err := errors.New(fmt.Sprintf("sent %d NACKS without receiving response, giving up.", sent_nacks))
		return response, err
	} else {
		fdm.SendACK()
		return response, nil
	}
}

func (fdm *FDM) Close() {
	fdm.s.Close()
}

func (fdm *FDM) SendACK() {
	//fdm.s.Write([]byte("0x06"))
	msg := []byte{0x06}
	fdm.s.Write(msg)
	// fdm.Write("0x06", true, 1)
}

func (fdm *FDM) SendNACK() {
	fdm.s.Write([]byte{0x015})
}
