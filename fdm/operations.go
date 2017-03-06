package fdm

import (
	"log"
	"strconv"
	"strings"
)

// HashAndSignMsg is a shortcut function that prepares the string that should be sent to the FDM in case of sales or refund
func HashAndSignMsg(event_label string, t Ticket) string {
	// format: identifier + sequence + retry + ticket_date + ticket_time_period + user_id + RCRS + string(ticket_number) + event_label + total_amount + 4 vats + plu
	identifier := "H"
	sequence := "01"
	retry := "2"
	dt := t.CreatedAt.Format("20060102")
	period := t.CreatedAt.Format("150405")
	amount := strconv.FormatFloat(t.TotalAmount, 'f', 2, 64)
	amount = strings.Replace(amount, ".", "", 1)
	// make sure total amount is 11 length, 9.2
	for len(amount) < 11 {
		amount = " " + amount
	}
	tn := t.TicketNumber
	for len(tn) < 6 {
		tn = " " + tn
	}

	msg := identifier + sequence + retry + dt + period + t.UserID + t.RCRS + tn + event_label + amount
	// add VATs
	for _, v := range t.VATs {
		// make sure that every vat percentage is formatted as 4 numerical letters: yy.xx
		// make sure that every vat amount is formatted as 11 numerical letters: yyyyyyyyy.xx
		per := v.PercentageString()
		amount := v.FixedAmountString()
		msg += per + amount
	}
	log.Printf("PLU: %s\n", t.PLUHash)
	msg += t.PLUHash
	log.Println("HashAndSignMessage: " + msg)
	return msg
}
