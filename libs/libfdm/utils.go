package libfdm

import (
	"fmt"
	"log"
	"strconv"
)

func generateLowLevelMessage(message string) []byte {
	packet := []byte{0x02}
	msg := []byte(message)
	BCC := calculateLRC(msg)
	packet = append(packet, msg...)
	packet = append(packet, 0x03)
	packet = append(packet, BCC)

	return packet
}

func calculateLRC(message []byte) byte {
	var LRC = byte(0)
	for _, rune := range message {
		LRC = (LRC + rune) & 0xFF
	}
	LRC = ((LRC ^ 0xFF) + 1) & 0xFF
	return LRC
}

func incrementRetryCounter(packet *[]byte) {
	s := fmt.Sprint((*packet)[4])
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	i++
	(*packet)[4] = byte(i)
}
