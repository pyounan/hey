package libfdm

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
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
		log.Println(err)
	}
	i++
	(*packet)[4] = byte(i)
}

// FormatSequenceNumber formats sequence number as string of 2 letters length
func FormatSequenceNumber(val int) string {
	str := strconv.Itoa(val)
	if len(str) < 2 {
		str = "0" + str
	}
	return str
}

// FormatAmount formats a price or amount from float to an fdm amount string
// which has two numbers on the right acts as numbers after decimal point
func FormatAmount(oldVal float64) string {
	oldVal = math.Abs(oldVal)
	amount := strconv.FormatFloat(oldVal, 'f', 2, 64)
	amount = strings.Replace(amount, ".", "", 1)
	// make sure total amount is 11 length, 9.2
	for len(amount) < 11 {
		amount = " " + amount
	}
	return amount
}

// FormatTicketNumber formats ticket number as an FDM ticket number
// which consists of 6 letters
func FormatTicketNumber(oldVal string) string {
	tn := oldVal
	for len(tn) < 6 {
		tn = " " + tn
	}
	return tn
}

// FormatDate formats a date to an FDM date string
func FormatDate(oldVal string) string {
	t, _ := time.Parse("2006-01-02 15:04:05Z07:00", oldVal)
	str := t.Format("20060102")
	return str
}

// FormatTime formats a time to and FDM time string
func FormatTime(oldVal string) string {
	t, _ := time.Parse("2006-01-02 15:04:05Z07:00", oldVal)
	str := t.Format("150405")
	return str
}
