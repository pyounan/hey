package fdm

import (
	"log"
	"errors"

	"pos-proxy/config"
	"github.com/tarm/serial"
)

type FDM struct {
	c *serial.Config
	s *serial.Port
}

func New() *FDM {
	fdm := &FDM{}
	log.Println("Trying to stablish connection with FDM with configuration:")
	log.Printf("Port: %s", config.Config.FDM_Port)
	log.Printf("Baud Speed: %d", config.Config.FDM_Speed)
	fdm.c = &serial.Config{Name: config.Config.FDM_Port, Baud: config.Config.FDM_Speed}
	s, err := serial.OpenPort(fdm.c)
	fdm.s = s
	if err != nil {
		log.Println("Failed to stablish connection:")
		log.Fatal(err)
	}
	if _, err := fdm.Write("S000", false); err != nil {
		log.Fatal(err)
	}

	log.Println("Connection to FDM has beedn stablished successfully.")
	return fdm
}

func(fdm *FDM) SendAndWaitForACK(packet []byte) (bool, error) {
	// if the response is not valid we try to retry reading the answer again
	ack := 0x00
	max_retries := byte('3')
	for packet[4] < max_retries && ack != 0x06 {
		_, err := fdm.s.Write(packet)
		if err != nil {
			return false, err
		}
		res := make([]byte, 1)
		n, err := fdm.s.Read(res)
		log.Println(n)
		if err != nil {
			return false, err
		}
		incrementRetryCounter(packet)
		if res[0] != 0x00 {
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

func(fdm *FDM) Write(message string, just_wait_for_ACK bool) (bool, error) {
	packet := generateLowLevelMessage(message)
	got_response := false
	max_nacks := 1
	sent_nacks := 0
	ok, err := fdm.SendAndWaitForACK(packet)
	if ok == false {
		return false, errors.New("Didn't recieve ACK")
	}
	if just_wait_for_ACK {
		return true, nil
	}
	log.Println("ACK Received")
	for got_response == false && sent_nacks < max_nacks {
		stx := make([]byte, 1)
		_, err = fdm.s.Read(stx)
		if err != nil {
			return false, err
		}
                msg := make([]byte, 62)
                _, err = fdm.s.Read(msg)
		if err != nil {
			return false, err
		}
		etx := make([]byte, 1)
                _, err = fdm.s.Read(etx)
		if err != nil {
			return false, err
		}
		bcc := make([]byte, 1)
                _, err = fdm.s.Read(bcc)
		if err != nil {
			return false, err
		}
		// compare results
		log.Println(string(stx))
		if stx != nil && etx != nil && bcc != nil {
			got_response = true
			log.Println("got response")
			fdm.s.Write([]byte("0x06"))
		} else {
			log.Println("Received ACK but not a valid response, sending NACK....")
			sent_nacks += 1
			fdm.s.Write([]byte("0x015"))
		}
	}

	if got_response == false {
		log.Printf("sent %d NACKS without receiving response, giving up.", sent_nacks)
		return false, nil
	} else {
		return true, nil
	}
}
