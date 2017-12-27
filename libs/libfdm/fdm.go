package libfdm

import (
	"errors"
	"fmt"
	"log"

	"github.com/tarm/serial"
)

// FDM is a structure the defines the configuration and the port to the fdm
// connection.
type FDM struct {
	c *serial.Config
	s *serial.Port
}

// New Creates a new fdm connection and returns FDM struct.
func New(config *serial.Config) (*FDM, error) {
	fdm := &FDM{}
	fdm.c = config

	log.Printf("Trying to stablish connection with FDM serial port: %s ->\n", fdm.c.Name)
	s, err := serial.OpenPort(fdm.c)
	if err != nil {
		log.Printf("Failed to stablish connection with FDM serial port: %s\n", fdm.c.Name)
		return nil, err
	} else if s == nil {
		return nil, errors.New("Failed to stablish connection with FDM serial port for unknown reason, Kindly check fdm configuration.")
	}
	fdm.s = s

	log.Println("Connection to FDM serial port has been stablished successfully.")
	return fdm, nil
}

// SendAndWaitForACK sends a message to the FDM and retries until it recievs ACK.
func (fdm *FDM) sendAndWaitForACK(packet []byte) (bool, error) {
	// if the response is not valid we try to retry reading the answer again
	ack := 0x00
	maxRetries := byte('2')
	for packet[4] < maxRetries && ack != 0x06 {
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
		incrementRetryCounter(&packet)
		if res[0] == 0x06 {
			log.Println("ACK received.")
			ack = 0x06
		} else {
			log.Println("ACK wasn't received, retrying...")
		}
	}
	if ack == 0x06 {
		return true, nil
	}

	return false, errors.New("Didn't receive ACK")
}

// Write writes a message to the fdm, if just_wait_for_ACK is true, then it won't
// wait for the response. IF its false, then the process goes on and process the
// response.
func (fdm *FDM) Write(message string, just_wait_for_ACK bool, response_size int) ([]byte, error) {
	packet := generateLowLevelMessage(message)
	gotResponse := false
	maxNACKs := 2
	sentNACKs := 0
	response := []byte{}

	ok, err := fdm.sendAndWaitForACK(packet)
	if ok == false {
		log.Println(err)
		return response, errors.New("Didn't recieve ACK")
	}
	if just_wait_for_ACK {
		fdm.sendACK()
		return response, nil
	}
	for gotResponse == false && sentNACKs < maxNACKs {
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
			gotResponse = true
			response = msg
			log.Println("Received a valid response")
			break
		} else {
			log.Println("Not a valid response, sending NACK....")
			response = msg
			sentNACKs++
			fdm.sendNACK()
		}
	}

	if gotResponse == false {
		err := fmt.Errorf("sent %d NACKS without receiving response, giving up", sentNACKs)
		return response, err
	}

	fdm.sendACK()
	return response, nil
}

// Close closes port connection with FDM
func (fdm *FDM) Close() {
	if fdm.s != nil {
		fdm.s.Close()
	}
}

func (fdm *FDM) sendACK() {
	//fdm.s.Write([]byte("0x06"))
	msg := []byte{0x06}
	fdm.s.Write(msg)
	// fdm.Write("0x06", true, 1)
}

func (fdm *FDM) sendNACK() {
	fdm.s.Write([]byte{0x015})
}
