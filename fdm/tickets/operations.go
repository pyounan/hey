package tickets

import (
	"log"
	"strconv"
	"strings"
	"time"
)

// HashAndSignMsg is a shortcut function that prepares the string that should be sent to the FDM in case of sales or refund
func HashAndSignMsg(event_label string, plu_msg string, ticket_date time.Time, user_id string, ticket_number int, total_amount float64, vats []VAT, RCRS string) string {
	// format: identifier + sequence + retry + ticket_date + ticket_time_period + user_id + RCRS + string(ticket_number) + event_label + total_amount + 4 vats + plu
	identifier := "H"
	sequence := "01"
	retry := "2"
	dt := ticket_date.Format("20060102")
	period := ticket_date.Format("150405")
	amount := strconv.FormatFloat(total_amount, 'f', 2, 64)
	amount = strings.Replace(amount, ".", "", 1)
	// make sure total amount is 11 length, 9.2
	for len(amount) < 11 {
		amount = " " + amount
	}
	tn := strconv.Itoa(ticket_number)
	for len(tn) < 6 {
		tn = " " + tn
	}
	log.Println("Ticket Number: " + tn)

	msg := identifier + sequence + retry + dt + period + user_id + RCRS + tn + event_label + amount
	// add VATs
	for _, v := range vats {
		// make sure that every vat percentage is formatted as 4 numerical letters: yy.xx
		// make sure that every vat amount is formatted as 11 numerical letters: yyyyyyyyy.xx
		per := v.PercentageString()
		amount := v.FixedAmountString()
		log.Printf("Percentage: %v, Amount: %v\n", per, amount)
		msg += per + amount
	}
	msg += plu_msg
	log.Println("HashAndSignMessage: " + msg)
	return msg
}
