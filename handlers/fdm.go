package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"gopkg.in/mgo.v2/bson"

	"pos-proxy/db"
	"pos-proxy/ej"
	"pos-proxy/fdm"
)

type Request struct {
	ActionTime    string            `json:"action_time"`
	InvoiceNumber string            `json:"invoice_number"`
	TableNumber   string            `json:"table_number"`
	TerminalName  string            `json:"terminal_name"`
	Items         []fdm.POSLineItem `json:"items"`
	UserID        string            `json:"user_id"`
	RCRS          string            `json:"rcrs"`
	CashierName   string            `json:"cashier_name"`
	CashierNumber string            `json:"cashier_number"`
}

func FDMStatus(w http.ResponseWriter, r *http.Request) {
	f, err := fdm.New("")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	defer f.Close()
	ready, err := f.CheckStatus()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	json.NewEncoder(w).Encode(ready)
}

func SubmitInvoice(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req Request
	err := decoder.Decode(&req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer r.Body.Close()
	FDM, err := fdm.New(req.RCRS)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer FDM.Close()

	// check status
	ok, err := FDM.CheckStatus()
	if err != nil || ok == false {
		log.Println("Failed to get response from FDM")
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Failed to get response from FDM")
		return
	}
	req.Items = fixItemsPrice(req.Items)
	for _, item := range req.Items {
		log.Printf("item price: %f", item.Price)
	}
	// calculate total amount of each VAT rate
	vats := calculateVATs(req.Items)
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
	items := splitItemsByVATRates(req.Items, positiveVATs)
	if len(items) > 0 {
		err = sendMessage("PS", FDM, req, items)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
			return
		}
	}
	// send negative msg
	items = splitItemsByVATRates(req.Items, negativeVATs)
	if len(items) > 0 {
		err = sendMessage("PR", FDM, req, items)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
			return
		}
	}

	json.NewEncoder(w).Encode("success")
}

func Folio(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req Request
	err := decoder.Decode(&req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer r.Body.Close()
	FDM, err := fdm.New(req.RCRS)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer FDM.Close()

	VATs := calculateVATs(req.Items)
	total_amount := calculateTotalAmount(req.Items)
	t := fdm.Ticket{}
	t.ID = bson.NewObjectId()
	tn, err := db.GetNextTicketNumber()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	t.TicketNumber = strconv.Itoa(tn)
	t.TerminalName = req.TerminalName
	t.CashierName = req.CashierName
	t.CashierNumber = req.CashierNumber
	t.TableNumber = req.TableNumber
	t.UserID = req.UserID
	t.RCRS = req.RCRS
	t.InvoiceNumber = req.InvoiceNumber
	t.Items = req.Items
	t.TotalAmount = total_amount
	t.PLUHash = fdm.GeneratePLUHash(t.Items)
	t.VATs = make([]fdm.VAT, 4)
	t.VATs[0].Percentage = 21
	t.VATs[0].FixedAmount = VATs["A"]

	t.VATs[1].Percentage = 12
	t.VATs[1].FixedAmount = VATs["B"]

	t.VATs[2].Percentage = 6
	t.VATs[2].FixedAmount = VATs["C"]

	t.VATs[3].Percentage = 0
	t.VATs[3].FixedAmount = VATs["D"]
	// Don't send aything to FDM is there is no new items added
	if len(t.Items) == 0 {
		json.NewEncoder(w).Encode("success")
		return
	}
	err = db.DB.C("tickets").Insert(&t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v, err"))
		return
	}
	event_label := "PS"
	msg := fdm.HashAndSignMsg(event_label, t)
	res, err := FDM.Write(msg, false, 109)
	log.Println("finished writing to FDM")
	log.Println(res)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	db.UpdateLastTicketNumber(tn)
	pf_response := fdm.ProformaResponse{}
	response := pf_response.Process(res)
	go func() {
		err := ej.Log(event_label, t, response)
		if err != nil {
			log.Println(err)
		}
	}()
	json.NewEncoder(w).Encode("success")
}

func PayInvoice(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req Request
	err := decoder.Decode(&req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer r.Body.Close()
	FDM, err := fdm.New(req.RCRS)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer FDM.Close()

	VATs := calculateVATs(req.Items)
	total_amount := calculateTotalAmount(req.Items)
	t := fdm.Ticket{}
	t.ID = bson.NewObjectId()
	tn, err := db.GetNextTicketNumber()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	t.TicketNumber = strconv.Itoa(tn)
	t.TerminalName = req.TerminalName
	t.CashierName = req.CashierName
	t.CashierNumber = req.CashierNumber
	t.TableNumber = req.TableNumber
	t.UserID = req.UserID
	t.RCRS = req.RCRS
	t.InvoiceNumber = req.InvoiceNumber
	t.Items = req.Items
	t.TotalAmount = total_amount
	t.PLUHash = fdm.GeneratePLUHash(t.Items)
	t.VATs = make([]fdm.VAT, 4)
	t.VATs[0].Percentage = 21
	t.VATs[0].FixedAmount = VATs["A"]

	t.VATs[1].Percentage = 12
	t.VATs[1].FixedAmount = VATs["B"]

	t.VATs[2].Percentage = 6
	t.VATs[2].FixedAmount = VATs["C"]

	t.VATs[3].Percentage = 0
	t.VATs[3].FixedAmount = VATs["D"]
	// Don't send aything to FDM is there is no new items added
	if len(t.Items) == 0 {
		json.NewEncoder(w).Encode("success")
		return
	}
	err = db.DB.C("tickets").Insert(&t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v, err"))
		return
	}
	log.Println("Making a Normal Sale")
	event_label := "NS"
	msg := fdm.HashAndSignMsg(event_label, t)
	res, err := FDM.Write(msg, false, 109)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	db.UpdateLastTicketNumber(tn)
	go func(res []byte) {
		pf_response := fdm.ProformaResponse{}
		response := pf_response.Process(res)
		err := ej.Log(event_label, t, response)
		if err != nil {
			log.Println(err)
		}
	}(res)
	json.NewEncoder(w).Encode("success")
}
