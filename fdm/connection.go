package fdm

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
	"pos-proxy/config"
	"pos-proxy/db"
)

// FDM is a structure the defines the configuration and the port to the fdm
// connection.
type FDM struct {
	c *serial.Config
	s *serial.Port
}

// New Creates a new fdm connection and returns FDM struct.
func New() (*FDM, error) {
	fdm := &FDM{}
	log.Println("Trying to stablish connection with FDM with configuration:")
	log.Printf("Port: %s", config.Config.FDM_Port)
	log.Printf("Baud Speed: %d", config.Config.FDM_Speed)
	fdm.c = &serial.Config{Name: config.Config.FDM_Port, Baud: config.Config.FDM_Speed, ReadTimeout: time.Second * 5}
	s, err := serial.OpenPort(fdm.c)
	fdm.s = s
	if err != nil {
		log.Println("Failed to stablish connection:")
		return nil, err
	}

	log.Println("Connection to FDM has been stablished successfully.")
	return fdm, nil
}

// CheckStatus sends S000 to the FDM and check if its ready.
func (fdm *FDM) CheckStatus() (bool, error) {
	n, err := db.GetNextSequence()
	db.UpdateLastSequence(n)
	if err != nil {
		return false, err
	}
	msg := fmt.Sprintf("S%s0", FormatSequence(n))
	log.Println(msg)
	if _, err := fdm.Write(msg, true, 21); err != nil {
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
	fmt.Sprintf("%s", packet)
	for packet[4] < max_retries && ack != 0x06 {
		log.Println(packet)
		_, err := fdm.s.Write(packet)
		if err != nil {
			return false, err
		}
		log.Println("packet sent")
		res := make([]byte, 1)
		_, err = fdm.s.Read(res)
		if err != nil {
			log.Println("Couldn't read")
			log.Println(err)
			return false, err
		}
		incrementRetryCounter(packet)
		if res[0] != 0x00 {
			log.Println("ACK received.")
			log.Println(res[0])
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
func (fdm *FDM) Write(message string, just_wait_for_ACK bool, response_size int) (string, error) {
	packet := generateLowLevelMessage(message)
	got_response := false
	max_nacks := 2
	sent_nacks := 0
	response := ""

	ok, err := fdm.SendAndWaitForACK(packet)
	if ok == false {
		log.Println(err)
		return "", errors.New("Didn't recieve ACK")
	}
	log.Println("ACK Received")
	if just_wait_for_ACK {
		return "", nil
	}
	for got_response == false && sent_nacks < max_nacks {
		stx := make([]byte, 1)
		_, err = fdm.s.Read(stx)
		if err != nil {
			return "", err
		}
		time.Sleep(time.Second * 5)
		msg := make([]byte, response_size)
		msg_len, err := fdm.s.Read(msg)
		if err != nil {
			return "", err
		}
		etx := make([]byte, 1)
		_, err = fdm.s.Read(etx)
		if err != nil {
			return "", err
		}
		bcc := make([]byte, 1)
		_, err = fdm.s.Read(bcc)
		if err != nil {
			return "", err
		}
		// compare results
		fmt.Printf("%v, %s, %v, %v\n", stx[0], msg[:msg_len], etx[0], bcc[0])
		if fmt.Sprintf("%v", stx) != fmt.Sprintf("%v", 0x02) && fmt.Sprintf("%v", etx) != fmt.Sprintf("%v", 0x03) && bcc != nil && calculateLRC(msg) == bcc[0] {
			got_response = true
			log.Println("got response")
			response = string(msg)
			fdm.s.Write([]byte("0x06"))
		} else {
			log.Println("Received ACK but not a valid response, sending NACK....")
			response = string(msg)
			sent_nacks += 1
			fdm.s.Write([]byte("0x015"))
		}
	}

	if got_response == false {
		err := errors.New(fmt.Sprintf("sent %d NACKS without receiving response, giving up.", sent_nacks))
		return response, err
	} else {
		return response, nil
	}
}

func (fdm *FDM) Close() {
	fdm.s.Close()
}
