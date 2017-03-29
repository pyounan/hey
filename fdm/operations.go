package fdm

import (
	"log"

	"pos-proxy/db"
)

// HashAndSignMsg is a shortcut function that prepares the string that should be sent to the FDM in case of sales or refund
func HashAndSignMsg(event_label string, t Ticket) string {
	// format: identifier + sequence + retry + ticket_date + ticket_time_period + user_id + RCRS + string(ticket_number) + event_label + total_amount + 4 vats + plu
	identifier := "H"
	ns, _ := db.GetNextSequence()
	db.UpdateLastSequence(ns)
	sequence := FormatSequence(ns)
	retry := "0"
	dt := FormatDate(t.ActionTime)
	period := FormatTime(t.ActionTime)
	amount := FormatAmount(t.TotalAmount)
	tn := FormatTicketNumber(t.TicketNumber)

	msg := identifier + sequence + retry + dt + period + t.UserID + t.RCRS + tn + event_label + amount
	// add VATs
	for _, v := range t.VATs {
		// make sure that every vat percentage is formatted as 4 numerical letters: yy.xx
		// make sure that every vat amount is formatted as 11 numerical letters: yyyyyyyyy.xx
		per := v.PercentageString()
		amount := v.FixedAmountString()
		msg += per + amount
	}
	msg += t.PLUHash
	log.Println("HashAndSignMessage: " + msg)
	return msg
}
