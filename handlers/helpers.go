package handlers

import (
	"errors"
	"fmt"
	"log"
	"math"
	"pos-proxy/fdm"
	"strconv"

	"gopkg.in/mgo.v2/bson"

	"pos-proxy/db"
	"pos-proxy/ej"
)

func fixItemsPrice(items []fdm.POSLineItem) []fdm.POSLineItem {
	for _, item := range items {
		operator := 1.0
		if item.Price < 0 {
			operator = -1.0
		}
		item.Price = operator * item.Quantity * item.UnitPrice
		log.Printf("item price %f", item.Price)
	}
	return items
}

func sendMessage(event_label string, FDM *fdm.FDM, req Request, items []fdm.POSLineItem) (fdm.ProformaResponse, error) {
	if len(items) == 0 {
		return fdm.ProformaResponse{}, nil
	}
	VATs := calculateVATs(items)
	totalAmount := calculateTotalAmount(items)
	t := fdm.Ticket{}
	t.ID = bson.NewObjectId()
	tn, err := db.GetNextTicketNumber(req.RCRS)
	if err != nil {
		return fdm.ProformaResponse{}, err
	}
	t.ActionTime = req.ActionTime
	t.TicketNumber = strconv.Itoa(tn)
	t.TerminalName = req.TerminalName
	t.CashierName = req.CashierName
	t.CashierNumber = req.CashierNumber
	t.TableNumber = req.TableNumber
	t.UserID = req.UserID
	t.RCRS = req.RCRS
	t.InvoiceNumber = req.InvoiceNumber
	t.Items = items
	t.TotalAmount = totalAmount
	t.PLUHash = fdm.GeneratePLUHash(t.Items)
	t.Payments = req.Payments
	t.ChangeAmount = req.ChangeAmount
	t.IsClosed = req.IsClosed
	t.VATs = make([]fdm.VAT, 4)
	t.VATs[0].Percentage = 21
	t.VATs[0].FixedAmount = math.Abs(VATs["A"])

	t.VATs[1].Percentage = 12
	t.VATs[1].FixedAmount = math.Abs(VATs["B"])

	t.VATs[2].Percentage = 6
	t.VATs[2].FixedAmount = math.Abs(VATs["C"])

	t.VATs[3].Percentage = 0
	t.VATs[3].FixedAmount = math.Abs(VATs["D"])
	// Don't send aything to FDM is there is no new items added
	err = db.DB.C("tickets").Insert(&t)
	if err != nil {
		return fdm.ProformaResponse{}, err
	}

	log.Println("========= PLU Items =========")
	for _, i := range items {
		log.Println(i.String())
	}
	log.Println("=============================")
	msg := fdm.HashAndSignMsg(event_label, t)
	res, err := FDM.Write(msg, false, 109)
	if err != nil {
		return fdm.ProformaResponse{}, err
	}
	if err := db.UpdateLastTicketNumber(req.RCRS, tn); err != nil {
		log.Println(err)
	}
	pf_response := fdm.ProformaResponse{}
	response := pf_response.Process(res)
	if pf_response.Error2 != "00" && pf_response.Error2 != "01" {
		err := errors.New(fmt.Sprintf("FDM Response error, error2 code: %s, erro3 code: %s", pf_response.Error2, pf_response.Error3))
		return fdm.ProformaResponse{}, err
	}
	// send event to Electrnoic Journal
	go func() {
		err := ej.Log(event_label, t, response)
		if err != nil {
			log.Println(err)
		}
	}()
	return pf_response, nil
}

func splitItemsByVATRates(items []fdm.POSLineItem, rates []string) []fdm.POSLineItem {

	result := []fdm.POSLineItem{}
	for _, item := range items {
		for _, rate := range rates {
			if item.VAT == rate {
				result = append(result, item)
			}
		}
	}
	return result
}

//func groupItemsBySign(items []fdm.POSLineItem) map[string][]fdm.POSLineItem {
//	result := make(map[string][]fdm.POSLineItem)
//	result["+"] = []fdm.POSLineItem{}
//	result["-"] = []fdm.POSLineItem{}
//	for _, i := range items {
//		price := i.NetAmount + i.TaxAmount
//		if price > 0 {
//			result["+"] = append(result["+"], i)
//		} else {
//			i.Price = math.Abs(i.Price)
//			i.NetAmount = math.Abs(i.NetAmount)
//			i.TaxAmount = math.Abs(i.TaxAmount)
//			i.Quantity = math.Abs(i.Quantity)
//			result["-"] = append(result["-"], i)
//		}
//	}
//
//	return result
//}

func calculateVATs(items []fdm.POSLineItem) map[string]float64 {
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

func calculateTotalAmount(items []fdm.POSLineItem) float64 {
	total := 0.0

	for _, i := range items {
		total += i.Price
	}

	return total
}
