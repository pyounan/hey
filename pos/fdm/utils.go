package fdm

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"pos-proxy/libs/libfdm"
	"pos-proxy/pos/models"
)

type FDMErrors map[string]map[string]string

var fdmErrors = FDMErrors{}

func init() {
	fdmErrors["1"] = make(map[string]string)
	fdmErrors["1"]["01"] = "FDM Data Storage 90% Full"
	fdmErrors["1"]["02"] = "Request Already Answered"
	fdmErrors["1"]["03"] = "No Record"
	fdmErrors["2"] = make(map[string]string)
	fdmErrors["2"]["01"] = "No VSC or faulty VSC"
	fdmErrors["2"]["02"] = "VSC not initialized with pin"
	fdmErrors["2"]["03"] = "VSC locked"
	fdmErrors["2"]["04"] = "PIN not valid"
	fdmErrors["2"]["05"] = "FDM Data Storage Full"
	fdmErrors["2"]["06"] = "Unkown message identifier"
	fdmErrors["2"]["07"] = "Invalid data in message"
	fdmErrors["2"]["08"] = "FDM not operational"
	fdmErrors["2"]["09"] = "FDM realtime clock corrupted"
	fdmErrors["2"]["10"] = "VSC version not supported by FDM"
	fdmErrors["2"]["11"] = "Port 4 not ready"
}

// ApplySHA1 convert text to SHA1
func applySHA1(text string) string {
	msg := sha1.New()
	msg.Write([]byte(text))
	bs := msg.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func generatePLUHash(items []models.EJEvent) string {
	text := ""
	for _, i := range items {
		text += fmt.Sprintf("%s", i.String())
	}
	return applySHA1(text)
}

func calculateVATs(items []models.EJEvent) map[string]float64 {
	VATs := make(map[string]float64)
	VATs["A"] = 0
	VATs["B"] = 0
	VATs["C"] = 0
	VATs["D"] = 0

	for _, i := range items {
		VATs[i.VATCode] += i.NetAmount
	}

	return VATs
}

func calculateTotalAmount(items []models.EJEvent) float64 {
	total := 0.0

	for _, i := range items {
		total += i.Price
	}

	return total
}

// prepareHashAndSignMsg is a shortcut function that prepares the string that should be sent to the FDM in case of sales or refund
func prepareHashAndSignMsg(RCRS string, eventLabel string, t models.FDMTicket) string {
	// format: identifier + sequence + retry + ticket_date + ticket_time_period + user_id + RCRS + string(ticket_number) + eventLabel + total_amount + 4 vats + plu
	identifier := "H"
	ns, _ := GetNextSequence(RCRS)
	// db.UpdateLastSequence(ns)
	sequence := libfdm.FormatSequenceNumber(ns)
	retry := "0"
	dt := libfdm.FormatDate(t.ActionTime)
	period := libfdm.FormatTime(t.ActionTime)
	amount := libfdm.FormatAmount(t.TotalAmount)
	tn := libfdm.FormatTicketNumber(t.TicketNumber)

	msg := identifier + sequence + retry + dt + period + t.UserID + t.RCRS + tn + eventLabel + amount
	// add VATs
	for _, v := range t.VATs {
		// make sure that every vat percentage is formatted as 4 numerical letters: yy.xx
		// make sure that every vat amount is formatted as 11 numerical letters: yyyyyyyyy.xx
		per := v.PercentageString()
		amount := v.FixedAmountString()
		msg += per + amount
	}
	msg += t.PLUHash
	fmt.Println("HashAndSignMessage: " + msg)
	return msg
}

// CheckFDMError parses the error values of fdm response and converts them to human readable error, if any
func CheckFDMError(res models.FDMResponse) error {
	if res.Error1 == "2" {
		return errors.New(fdmErrors[res.Error1][res.Error2])
	}
	return nil
}

// CheckFDMWarning parses the error values of fdm response and converts them to human readable warning, if any
func CheckFDMWarning(res models.FDMResponse) error {
	if res.Error1 == "1" {
		return errors.New(fdmErrors[res.Error1][res.Error2])
	}
	return nil
}

/*
func separateCondimentsAndDiscounts(rawItems []models.EJEvent, submitMode bool) []models.EJEvent {
	items := []models.EJEvent{}
	for _, item := range rawItems {
		if submitMode == true && item.Quantity == item.SubmittedQuantity {
			continue
		}
		priceOperator := 1
		item.LineItemType = "sales"
		item.TaxAmount = item.Price - item.NetAmount
		if item.Price < -1 {
			priceOperator = -1
			item.LineItemType = "return"
			if item.NetAmount > 0 {
				item.NetAmount *= -1
			}
		}
		if item.OpenItem {
			item.Description = item.Comment
		}
		if submitMode == false {
			items = append(items, item)
		}

		for _, cond := range item.CondimentLineItems {
			if cond.Price == 0 {
				continue
			}
			c := models.POSLineItem{}
			c.Description = cond.Description
			c.LineItemType = item.LineItemType
			c.IsCondiment = true
			c.UnitPrice = cond.Price
			c.Price = float64(item.Quantity) * float64(priceOperator) * cond.Price
			c.Quantity = item.Quantity
			c.VAT = cond.VAT
			c.VATPercentage = cond.VATPercentage
			c.NetAmount = cond.NetAmount
			if c.Price < 0 && c.NetAmount > 0 {
				c.NetAmount *= -1
			}
			c.TaxAmount = c.Price - c.NetAmount
			items = append(items, c)
		}

		for _, disc := range item.GroupedAppliedDiscounts {
			d := models.POSLineItem{}
			//d.Item = item.Item
			d.LineItemType = "discount"
			if item.Price < 0 {
				d.Price = math.Abs(disc.Amount)
			} else {
				d.Price = -1 * math.Abs(disc.Amount)
			}
			d.UnitPrice = d.Price
			d.Quantity = item.Quantity
			d.Description = fmt.Sprintf("Discount %2f%", disc.Percentage)
			d.VAT = disc.VAT
			d.VATPercentage = disc.VATPercentage
			d.NetAmount = disc.NetAmount
			if item.Price > 0 {
				d.NetAmount = -1 * math.Abs(disc.NetAmount)
			}
			d.TaxAmount = d.Price - d.NetAmount
			// only add discounts if mode is not submit, because we already add discoutns from events
			if submitMode == false {
				items = append(items, d)
			}
		}
	}
	return items
}*/

func splitItemsByVATRates(items []models.EJEvent, rates []string) []models.EJEvent {
	result := []models.EJEvent{}
	for _, item := range items {
		for _, rate := range rates {
			if item.VATCode == rate {
				result = append(result, item)
			}
		}
	}
	return result
}
