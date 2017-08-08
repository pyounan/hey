package fdm

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"math"
	"pos-proxy/db"
	"pos-proxy/pos/models"
	"strconv"
	"strings"
	"time"
)

// ApplySHA1 convert text to SHA1
func applySHA1(text string) string {
	msg := sha1.New()
	msg.Write([]byte(text))
	bs := msg.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func generatePLUHash(items []models.POSLineItem) string {
	text := ""
	for _, i := range items {
		text += fmt.Sprintf("%s", i.String())
	}
	return applySHA1(text)
}

// FormatSequence formats sequence number as string of 2 letters length
func formatSequence(val int) string {
	str := strconv.Itoa(val)
	if len(str) < 2 {
		str = "0" + str
	}
	return str
}

func formatAmount(old_val float64) string {
	old_val = math.Abs(old_val)
	amount := strconv.FormatFloat(old_val, 'f', 2, 64)
	amount = strings.Replace(amount, ".", "", 1)
	// make sure total amount is 11 length, 9.2
	for len(amount) < 11 {
		amount = " " + amount
	}
	log.Println("amount: ", amount)
	return amount
}

func formatTicketNumber(old_val string) string {
	tn := old_val
	for len(tn) < 6 {
		tn = " " + tn
	}
	return tn
}

func formatDate(old_val string) string {
	t, _ := time.Parse("2006-01-02 15:04:05Z07:00", old_val)
	str := t.Format("20060102")
	return str
}

func formatTime(old_val string) string {
	t, _ := time.Parse("2006-01-02 15:04:05Z07:00", old_val)
	str := t.Format("150405")
	return str
}

// summarizeVAT calculates the total net_amount and vat_amount of each
// VAT rate
func summarizeVAT(items *[]models.POSLineItem) map[string]models.VATSummary {
	summary := make(map[string]models.VATSummary)
	rates := []string{"A", "B", "C", "D", "Total"}
	for _, r := range rates {
		summary[r] = models.VATSummary{}
		summary[r]["net_amount"] = 0
		summary[r]["vat_amount"] = 0
		summary[r]["taxable_amount"] = 0
	}
	for _, item := range *items {
		summary[item.VAT]["net_amount"] += item.NetAmount
		summary[item.VAT]["vat_amount"] += item.NetAmount * item.VATPercentage / 100
		summary[item.VAT]["taxable_amount"] += item.Price

		summary["Total"]["net_amount"] += item.NetAmount
		summary["Total"]["vat_amount"] += item.NetAmount * item.VATPercentage / 100
		summary["Total"]["taxable_amount"] += item.Price
	}

	return summary
}

func calculateVATs(items []models.POSLineItem) map[string]float64 {
	VATs := make(map[string]float64)
	VATs["A"] = 0
	VATs["B"] = 0
	VATs["C"] = 0
	VATs["D"] = 0

	for _, i := range items {
		VATs[i.VAT] += i.NetAmount
	}

	return VATs
}

func calculateTotalAmount(items []models.POSLineItem) float64 {
	total := 0.0

	for _, i := range items {
		total += i.Price
	}

	return total
}

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
	var LRC byte = byte(0)
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
	i += 1
	(*packet)[4] = byte(i)
}

// prepareHashAndSignMsg is a shortcut function that prepares the string that should be sent to the FDM in case of sales or refund
func prepareHashAndSignMsg(RCRS string, event_label string, t models.FDMTicket) string {
	// format: identifier + sequence + retry + ticket_date + ticket_time_period + user_id + RCRS + string(ticket_number) + event_label + total_amount + 4 vats + plu
	identifier := "H"
	ns, _ := db.GetNextSequence(RCRS)
	// db.UpdateLastSequence(ns)
	sequence := formatSequence(ns)
	retry := "0"
	dt := formatDate(t.ActionTime)
	period := formatTime(t.ActionTime)
	amount := formatAmount(t.TotalAmount)
	tn := formatTicketNumber(t.TicketNumber)

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

func CheckFDMError(res models.FDMResponse) error {
	if res.Error1 != "0" {
		type FDMErrors map[string]map[string]string
		var fdmErrors = FDMErrors{}
		fdmErrors["01"] = make(map[string]string)
		fdmErrors["01"]["01"] = "FDM Data Storage 90% Full"
		fdmErrors["01"]["02"] = "Request Already Answered"
		fdmErrors["01"]["03"] = "No Record"
		fdmErrors["02"] = make(map[string]string)
		fdmErrors["02"]["01"] = "No VSC or faulty VSC"
		fdmErrors["02"]["02"] = "VSC not initialized with pin"
		fdmErrors["02"]["03"] = "VSC locked"
		fdmErrors["02"]["04"] = "PIN not valid"
		fdmErrors["02"]["05"] = "FDM Data Storage Full"
		fdmErrors["02"]["06"] = "Unkown message identifier"
		fdmErrors["02"]["07"] = "Invalid data in message"
		fdmErrors["02"]["08"] = "FDM not operational"
		fdmErrors["02"]["09"] = "FDM realtime clock corrupted"
		fdmErrors["02"]["10"] = "VSC version not supported by FDM"
		fdmErrors["02"]["11"] = "Port 4 not ready"
		return errors.New(fdmErrors[res.Error1][res.Error2])
	}
	return nil
}

func separateCondimentsAndDiscounts(rawItems []models.POSLineItem) []models.POSLineItem {
	items := []models.POSLineItem{}
	return items
}

func splitItemsByVATRates(items []models.POSLineItem, rates []string) []models.POSLineItem {
	result := []models.POSLineItem{}
	for _, item := range items {
		for _, rate := range rates {
			if item.VAT == rate {
				result = append(result, item)
			}
		}
	}
	return result
}
