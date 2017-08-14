package fdm

import (
	"fmt"
	"log"
	"math"
	"pos-proxy/db"
	"pos-proxy/ej"
	"pos-proxy/libs/libfdm"
	"pos-proxy/pos/models"
	"strconv"

	"gopkg.in/mgo.v2/bson"
)

// CheckStatus sends S000 to the FDM and check if its ready.
func CheckStatus(fdm *libfdm.FDM, RCRS string) (models.FDMResponse, error) {
	n, err := db.GetNextSequence(RCRS)
	if err != nil {
		return models.FDMResponse{}, err
	}
	msg := fmt.Sprintf("S%s0", formatSequence(n))
	res, err := fdm.Write(msg, false, 21)
	if err != nil {
		log.Println("Error: ", err)
		return models.FDMResponse{}, err
	}

	log.Println("FDM is ready.")
	response := models.FDMResponse{}
	response.ProcessStatus(res)
	return response, nil
}

func sendHashAndSignMessage(fdm *libfdm.FDM, eventLabel string,
	req models.InvoicePOSTRequest, items []models.POSLineItem) (models.FDMResponse, error) {
	if len(items) == 0 {
		return models.FDMResponse{}, nil
	}
	VATs := calculateVATs(items)
	totalAmount := calculateTotalAmount(items)
	t := models.FDMTicket{}
	t.ID = bson.NewObjectId()
	tn, err := db.GetNextTicketNumber(req.RCRS)
	if err != nil {
		return models.FDMResponse{}, err
	}
	t.ActionTime = req.ActionTime
	t.TicketNumber = strconv.Itoa(tn)
	t.TerminalName = req.TerminalName
	t.CashierName = req.CashierName
	t.CashierNumber = strconv.Itoa(req.CashierNumber)
	t.TableNumber = strconv.Itoa(req.Invoice.TableNumber)
	t.UserID = strconv.Itoa(req.CashierID)
	t.RCRS = req.RCRS
	t.InvoiceNumber = req.Invoice.InvoiceNumber
	t.Items = items
	t.TotalAmount = totalAmount
	t.PLUHash = generatePLUHash(t.Items)
	t.Payments = req.Payments
	t.ChangeAmount = req.ChangeAmount
	t.IsClosed = req.IsClosed
	t.VATs = make([]models.VAT, 4)
	t.VATs[0].Percentage = 21
	t.VATs[0].FixedAmount = math.Abs(VATs["A"])

	t.VATs[1].Percentage = 12
	t.VATs[1].FixedAmount = math.Abs(VATs["B"])

	t.VATs[2].Percentage = 6
	t.VATs[2].FixedAmount = math.Abs(VATs["C"])

	t.VATs[3].Percentage = 0
	t.VATs[3].FixedAmount = math.Abs(VATs["D"])
	t.VATSummary = summarizeVAT(&t.Items)
	// Don't send aything to FDM if is there is no new items added
	err = db.DB.C("tickets").Insert(&t)
	if err != nil {
		return models.FDMResponse{}, err
	}

	log.Println("========= PLU Items =========")
	for _, i := range items {
		log.Println(i.String())
	}
	log.Println("=============================")
	msg := prepareHashAndSignMsg(req.RCRS, eventLabel, t)
	res, err := fdm.Write(msg, false, 109)
	if err != nil {
		return models.FDMResponse{}, err
	}
	if err := db.UpdateLastTicketNumber(req.RCRS, tn); err != nil {
		log.Println(err)
	}
	pf_response := models.FDMResponse{}
	stringRes := pf_response.Process(res, t)
	err = CheckFDMError(pf_response)
	if err != nil {
		return pf_response, err
	}

	go func(eventLabel string, t models.FDMTicket, stringRes map[string]interface{}) {
		ej.Log(eventLabel, stringRes)
	}(eventLabel, t, stringRes)

	return pf_response, nil
}

func Submit(fdm *libfdm.FDM, data models.InvoicePOSTRequest) ([]models.FDMResponse, error) {
	// check status
	resp, err := CheckStatus(fdm, data.RCRS)
	if err != nil {
		return []models.FDMResponse{resp}, err
	}
	// req.Items = fixItemsPrice(req.Items)
	items := []models.POSLineItem{}
	for _, e := range data.Invoice.Events {
		log.Printf("ITEM: %v\n", e.Item)
		items = append(items, e.Item)
	}
	// calculate total amount of each VAT rate
	items = separateCondimentsAndDiscounts(items)
	vats := calculateVATs(items)
	positiveVATs := []string{}
	negativeVATs := []string{}
	for rate, amount := range vats {
		if amount >= 0 {
			positiveVATs = append(positiveVATs, rate)
		} else if amount < 0 {
			negativeVATs = append(negativeVATs, rate)
		}
	}

	responses := []models.FDMResponse{}

	// send positive msg
	positiveItems := splitItemsByVATRates(items, positiveVATs)
	log.Println("positive items", len(positiveItems))
	if len(positiveItems) > 0 {
		res, err := sendHashAndSignMessage(fdm, "PS", data, positiveItems)
		if err != nil {
			return responses, err
		}
		responses = append(responses, res)
	}
	// send negative msg
	negativeItems := splitItemsByVATRates(items, negativeVATs)
	if len(negativeItems) > 0 {
		res, err := sendHashAndSignMessage(fdm, "PR", data, negativeItems)
		if err != nil {
			return responses, err
		}
		responses = append(responses, res)
	}

	return responses, nil
}

func Folio(fdm *libfdm.FDM, data models.InvoicePOSTRequest) ([]models.FDMResponse, error) {
	resp, err := CheckStatus(fdm, data.RCRS)
	if err != nil {
		return []models.FDMResponse{resp}, err
	}

	responses := []models.FDMResponse{}

	// now send the whole invoice
	items := separateCondimentsAndDiscounts(data.Invoice.Items)
	vats := calculateVATs(items)
	positiveVATs := []string{}
	negativeVATs := []string{}
	for rate, amount := range vats {
		if amount >= 0 {
			positiveVATs = append(positiveVATs, rate)
		} else if amount < 0 {
			negativeVATs = append(negativeVATs, rate)
		}
	}

	// send positive msg
	positiveItems := splitItemsByVATRates(items, positiveVATs)
	if len(items) > 0 {
		res, err := sendHashAndSignMessage(fdm, "PS", data, positiveItems)
		if err != nil {
			return responses, err
		}
		responses = append(responses, res)
	}
	// send negative msg
	negativeItems := splitItemsByVATRates(items, negativeVATs)
	if len(items) > 0 {
		res, err := sendHashAndSignMessage(fdm, "PR", data, negativeItems)
		if err != nil {
			return responses, err
		}
		responses = append(responses, res)
	}

	return responses, nil
}

func Payment(fdm *libfdm.FDM, data models.InvoicePOSTRequest) ([]models.FDMResponse, error) {
	resp, err := CheckStatus(fdm, data.RCRS)
	if err != nil {
		return []models.FDMResponse{resp}, err
	}

	responses := []models.FDMResponse{}

	// now send the whole invoice
	items := separateCondimentsAndDiscounts(data.Invoice.Items)
	vats := calculateVATs(items)
	positiveVATs := []string{}
	negativeVATs := []string{}
	for rate, amount := range vats {
		if amount >= 0 {
			positiveVATs = append(positiveVATs, rate)
		} else if amount < 0 {
			negativeVATs = append(negativeVATs, rate)
		}
	}

	// send positive msg
	positiveItems := splitItemsByVATRates(items, positiveVATs)
	if len(items) > 0 {
		res, err := sendHashAndSignMessage(fdm, "PS", data, positiveItems)
		if err != nil {
			return responses, err
		}
		responses = append(responses, res)
	}
	// send negative msg
	negativeItems := splitItemsByVATRates(items, negativeVATs)
	if len(items) > 0 {
		res, err := sendHashAndSignMessage(fdm, "PR", data, negativeItems)
		if err != nil {
			return responses, err
		}
		responses = append(responses, res)
	}

	return responses, nil
}
