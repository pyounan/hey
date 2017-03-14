package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"pos-proxy/db"
	"pos-proxy/fdm"
)

func FDMStatus(w http.ResponseWriter, r *http.Request) {
	f, err := fdm.New()
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
	type Request struct {
		InvoiceID     string             `json:"_id"`
		InvoiceNumber string             `json:"invoice_number"`
		Items         []fdm.POSLineItem  `json:"posinvoicelineitem_set"`
		UserID        string             `json:"user_id"`
		RCRS          string             `json:"rcrs"`
		Total         float64            `json:"total"`
		VATs          map[string]float64 `json:"vats"`
	}
	var req Request
	err := decoder.Decode(&req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer r.Body.Close()
	FDM, err := fdm.New()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer FDM.Close()

	// split invoice items by + or - values
	grouped_items := GroupItemsBySign(req.Items)

	for sign, items := range grouped_items {
		log.Println(items)
		VATs := CalculateVATs(items)
		total_amount := CalculateTotalAmount(items)
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
		t.UserID = req.UserID
		t.RCRS = req.RCRS
		t.InvoiceNumber = req.InvoiceNumber
		t.Items = items
		t.TotalAmount = total_amount
		t.CreatedAt = time.Now()
		t.PLUHash = fdm.GeneratePLUHash(t.Items)
		t.IsSubmitted = false
		t.IsPaid = false
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
		event_label := ""
		if sign == "+" {
			event_label = "PS"
		} else {
			event_label = "PR"
		}
		msg := fdm.HashAndSignMsg(event_label, t)
		res, err := FDM.Write(msg, true, 64)
		log.Println("finished writing to FDM")
		log.Println(res)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
			return
		}
		db.UpdateLastTicketNumber(tn)
	}
	json.NewEncoder(w).Encode("success")
}

func VoidItem(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	type Request struct {
		InvoiceID     string             `json:"_id"`
		InvoiceNumber string             `json:"invoice_number"`
		Items         []fdm.POSLineItem  `json:"posinvoicelineitem_set"`
		UserID        string             `json:"user_id"`
		RCRS          string             `json:"rcrs"`
		Total         float64            `json:"total"`
		VATs          map[string]float64 `json:"vats"`
	}
	var req Request
	err := decoder.Decode(&req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer r.Body.Close()
	FDM, err := fdm.New()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer FDM.Close()

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
	t.UserID = req.UserID
	t.RCRS = req.RCRS
	t.InvoiceNumber = req.InvoiceNumber
	t.Items = req.Items
	t.TotalAmount = req.Total
	t.CreatedAt = time.Now()
	t.PLUHash = fdm.GeneratePLUHash(t.Items)
	t.IsSubmitted = false
	t.IsPaid = false
	t.VATs = make([]fdm.VAT, 4)
	t.VATs[0].Percentage = 21
	t.VATs[0].FixedAmount = req.VATs["A"]

	t.VATs[1].Percentage = 12
	t.VATs[1].FixedAmount = req.VATs["B"]

	t.VATs[2].Percentage = 6
	t.VATs[2].FixedAmount = req.VATs["C"]

	t.VATs[3].Percentage = 0
	t.VATs[3].FixedAmount = req.VATs["D"]
	err = db.DB.C("tickets").Insert(&t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v, err"))
		return
	}

	// Prepare the message that should be written to the FDM.
	msg := fdm.HashAndSignMsg("PR", t)

	res, err := FDM.Write(msg, true, 64)
	log.Println("finished writing to FDM")
	log.Println(res)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	db.UpdateLastTicketNumber(tn)
	json.NewEncoder(w).Encode("success")
}

func Folio(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	type Request struct {
		InvoiceID     string             `json:"_id"`
		InvoiceNumber string             `json:"invoice_number"`
		Items         []fdm.POSLineItem  `json:"posinvoicelineitem_set"`
		UserID        string             `json:"user_id"`
		RCRS          string             `json:"rcrs"`
		Total         float64            `json:"total"`
		VATs          map[string]float64 `json:"vats"`
	}
	var req Request
	err := decoder.Decode(&req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer r.Body.Close()
	FDM, err := fdm.New()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer FDM.Close()

	// split invoice items by + or - values
	grouped_items := GroupItemsBySign(req.Items)

	for sign, items := range grouped_items {
		VATs := CalculateVATs(items)
		total_amount := CalculateTotalAmount(items)
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
		t.UserID = req.UserID
		t.RCRS = req.RCRS
		t.InvoiceNumber = req.InvoiceNumber
		t.Items = items
		t.TotalAmount = total_amount
		t.CreatedAt = time.Now()
		t.PLUHash = fdm.GeneratePLUHash(t.Items)
		t.IsSubmitted = false
		t.IsPaid = false
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
		event_label := ""
		if sign == "+" {
			event_label = "PS"
		} else {
			event_label = "PR"
		}
		msg := fdm.HashAndSignMsg(event_label, t)
		res, err := FDM.Write(msg, true, 64)
		log.Println("finished writing to FDM")
		log.Println(res)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
			return
		}
		db.UpdateLastTicketNumber(tn)
	}
	json.NewEncoder(w).Encode("success")
}

func PayInvoice(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	type Request struct {
		InvoiceID     string             `json:"_id"`
		InvoiceNumber string             `json:"invoice_number"`
		Items         []fdm.POSLineItem  `json:"posinvoicelineitem_set"`
		UserID        string             `json:"user_id"`
		RCRS          string             `json:"rcrs"`
		Total         float64            `json:"total"`
		VATs          map[string]float64 `json:"vats"`
	}
	var req Request
	err := decoder.Decode(&req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer r.Body.Close()
	FDM, err := fdm.New()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
		return
	}
	defer FDM.Close()

	// split invoice items by + or - values
	grouped_items := GroupItemsBySign(req.Items)

	for sign, items := range grouped_items {
		VATs := CalculateVATs(items)
		total_amount := CalculateTotalAmount(items)
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
		t.UserID = req.UserID
		t.RCRS = req.RCRS
		t.InvoiceNumber = req.InvoiceNumber
		t.Items = items
		t.TotalAmount = total_amount
		t.CreatedAt = time.Now()
		t.PLUHash = fdm.GeneratePLUHash(t.Items)
		t.IsSubmitted = false
		t.IsPaid = false
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
		event_label := ""
		if sign == "+" {
			event_label = "NS"
		} else {
			event_label = "NR"
		}
		msg := fdm.HashAndSignMsg(event_label, t)
		res, err := FDM.Write(msg, true, 64)
		log.Println("finished writing to FDM")
		log.Println(res)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(fmt.Sprintf("%v", err))
			return
		}
		db.UpdateLastTicketNumber(tn)
	}
	json.NewEncoder(w).Encode("success")
}
