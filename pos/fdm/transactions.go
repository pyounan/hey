package fdm

import (
	"fmt"
	"log"
	"math"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/libs/libfdm"
	"pos-proxy/pos/locks"
	"pos-proxy/pos/models"
	"strconv"

	"gopkg.in/mgo.v2/bson"
)

// CheckStatus sends S000 to the FDM and check if its ready.
func CheckStatus(fdm *libfdm.FDM, RCRS string) (models.FDMResponse, error) {
	l, err := locks.LockFDM(RCRS)
	if err != nil {
		return models.FDMResponse{}, err
	}
	defer l.Unlock()
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

// SendHashAndSignMessage send a message to FDM and creates new ticket number
func SendHashAndSignMessage(fdm *libfdm.FDM, eventLabel string,
	req models.InvoicePOSTRequest, items []models.EJEvent) (models.FDMResponse, error) {
	// If invoice is not void and there is no new items to add, then return
	if req.Invoice.VoidReason == "" && len(items) == 0 {
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
	if req.Invoice.TableNumber != nil {
		t.TableNumber = strconv.Itoa(int(*req.Invoice.TableNumber))
	}
	t.UserID = req.EmployeeID
	t.RCRS = req.RCRS
	t.InvoiceNumber = req.Invoice.InvoiceNumber
	t.Items = items
	t.TotalAmount = totalAmount
	t.PLUHash = generatePLUHash(t.Items)
	t.Postings = req.Postings
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
		return models.FDMResponse{}, err
	}
	pfResponse := models.FDMResponse{}
	pfResponse.SoftwareVersion = config.Version
	pfResponse.Process(res, t)
	err = CheckFDMError(pfResponse)
	log.Println("Checking FDM Errors")
	if err != nil {
		log.Println(err)
		return pfResponse, err
	}

	return pfResponse, nil
}

// Submit loops over the events of the invoice, condiments and discounts of unsubmitted items and
// sends them to FDM.
func Submit(fdm *libfdm.FDM, data models.InvoicePOSTRequest) ([]models.FDMResponse, error) {
	// check status
	resp, err := CheckStatus(fdm, data.RCRS)
	if err != nil {
		return []models.FDMResponse{resp}, err
	}
	l, err := locks.LockFDM(data.RCRS)
	if err != nil {
		return []models.FDMResponse{}, err
	}
	defer l.Unlock()
	items := data.Invoice.Events
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
	if len(positiveItems) > 0 {
		res, err := SendHashAndSignMessage(fdm, "PS", data, positiveItems)
		if err != nil {
			return responses, err
		}
		responses = append(responses, res)
	}
	// send negative msg
	negativeItems := splitItemsByVATRates(items, negativeVATs)
	if len(negativeItems) > 0 {
		res, err := SendHashAndSignMessage(fdm, "PR", data, negativeItems)
		if err != nil {
			return responses, err
		}
		responses = append(responses, res)
	}

	return responses, nil
}

// Folio sends PS or PR with the whole invoice to the FDM
func Folio(fdm *libfdm.FDM, data models.InvoicePOSTRequest) ([]models.FDMResponse, error) {
	resp, err := CheckStatus(fdm, data.RCRS)
	if err != nil {
		return []models.FDMResponse{resp}, err
	}

	l, err := locks.LockFDM(data.RCRS)
	if err != nil {
		return []models.FDMResponse{}, err
	}
	defer l.Unlock()

	responses := []models.FDMResponse{}

	// now send the whole invoice
	items := data.Invoice.GroupedLineItems
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
	if len(positiveItems) > 0 {
		res, err := SendHashAndSignMessage(fdm, "PS", data, positiveItems)
		if err != nil {
			return responses, err
		}
		responses = append(responses, res)
	}
	// send negative msg
	negativeItems := splitItemsByVATRates(items, negativeVATs)
	if len(negativeItems) > 0 {
		res, err := SendHashAndSignMessage(fdm, "PR", data, negativeItems)
		if err != nil {
			return responses, err
		}
		responses = append(responses, res)
	}

	return responses, nil
}

//Payment adds NS or NR to fdm
func Payment(fdm *libfdm.FDM, data models.InvoicePOSTRequest) ([]models.FDMResponse, error) {
	resp, err := CheckStatus(fdm, data.RCRS)
	if err != nil {
		return []models.FDMResponse{resp}, err
	}

	l, err := locks.LockFDM(data.RCRS)
	if err != nil {
		return []models.FDMResponse{}, err
	}
	defer l.Unlock()

	responses := []models.FDMResponse{}

	items := data.Invoice.GroupedLineItems
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
	if len(positiveItems) > 0 {
		eventLabel := "NS"
		if data.Invoice.HouseUse == true {
			eventLabel = "PS"
		}
		res, err := SendHashAndSignMessage(fdm, eventLabel, data, positiveItems)
		if err != nil {
			return responses, err
		}
		responses = append(responses, res)
	}
	// send negative msg
	negativeItems := splitItemsByVATRates(items, negativeVATs)
	if len(negativeItems) > 0 {
		eventLabel := "NR"
		if data.Invoice.HouseUse == true {
			eventLabel = "PR"
		}
		res, err := SendHashAndSignMessage(fdm, eventLabel, data, negativeItems)
		if err != nil {
			return responses, err
		}
		responses = append(responses, res)
	}

	return responses, nil
}

// EmptyPLUHash sends NS ticket with an empty plu hash to fdm
func EmptyPLUHash(fdm *libfdm.FDM, data models.InvoicePOSTRequest) ([]models.FDMResponse, error) {
	l, err := locks.LockFDM(data.RCRS)
	if err != nil {
		return []models.FDMResponse{}, err
	}
	defer l.Unlock()

	responses := []models.FDMResponse{}

	res, err := SendHashAndSignMessage(fdm, "NS", data, []models.EJEvent{})
	if err != nil {
		return responses, err
	}
	responses = append(responses, res)

	return responses, nil
}
