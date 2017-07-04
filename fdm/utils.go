package fdm

import (
	"crypto/sha1"
	"fmt"
	"log"
	"math"
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

func FormatTicketNumber(old_val string) string {
	tn := old_val
	for len(tn) < 6 {
		tn = " " + tn
	}
	return tn
}

func FormatDate(old_val string) string {
	t, _ := time.Parse("2006-01-02 15:04:05Z07:00", old_val)
	str := t.Format("20060102")
	return str
}

func FormatTime(old_val string) string {
	t, _ := time.Parse("2006-01-02 15:04:05Z07:00", old_val)
	str := t.Format("150405")
	return str
}

// SummerizeVAT calculates the total net_amount and vat_amount of each
// VAT rate
func SummarizeVAT(items *[]POSLineItem) map[string]VATSummary {
	summary := make(map[string]VATSummary)
	rates := []string{"A", "B", "C", "D", "Total"}
	for _, r := range rates {
		summary[r] = VATSummary{}
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
