package fdm

import (
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ApplySHA1 convert text to SHA1
func ApplySHA1(text string) string {
	msg := sha1.New()
	msg.Write([]byte(text))
	bs := msg.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func GeneratePLUHash(items []POSLineItem) string {
	text := ""
	for _, i := range items {
		text += fmt.Sprintf("%s", i.String())
	}
	fmt.Printf("plu before hasshing: %s\n", text)
	return ApplySHA1(text)
}

// FormatSequence formats sequence number as string of 2 letters length
func FormatSequence(val int) string {
	str := strconv.Itoa(val)
	if len(str) < 2 {
		str = "0" + str
	}
	return str
}

func FormatAmount(old_val float64) string {
	amount := strconv.FormatFloat(old_val, 'f', 2, 64)
	amount = strings.Replace(amount, ".", "", 1)
	// make sure total amount is 11 length, 9.2
	for len(amount) < 11 {
		amount = " " + amount
	}
	return amount
}

func FormatTicketNumber(old_val string) string {
	tn := old_val
	for len(tn) < 6 {
		tn = " " + tn
	}
	fmt.Printf("Ticket Number: %s\n", tn)
	return tn
}

func FormatDate(old_val time.Time) string {
	return old_val.Format("20060102")
}

func FormatTime(old_val time.Time) string {
	return old_val.Format("150405")
}