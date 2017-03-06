package handlers

import (
	"encoding/json"
	"log"
	"net/http"
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
		json.NewEncoder(w).Encode(err)
		return
	}
	json.NewEncoder(w).Encode(ready)
}

func SubmitInvoice(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	type Request struct {
		InvoiceID     float64           `json:"_id"`
		InvoiceNumber string            `json:"invoice_number"`
		Items         []fdm.POSLineItem `json:"posinvoicelineitem_set"`
		UserID        string            `json:"user_id"`
		RCRS          string            `json:"rcrs"`
		Total         float64           `json:"total"`
	}
	var req Request
	err := decoder.Decode(&req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	defer r.Body.Close()
	FDM, err := fdm.New()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	defer FDM.Close()

	t := fdm.Ticket{}
	t.ID = bson.NewObjectId()
	t.TicketNumber = "123456"
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
	t.VATs[0].FixedAmount = 10

	t.VATs[1].Percentage = 12
	t.VATs[1].FixedAmount = 4.5

	t.VATs[2].Percentage = 6
	t.VATs[2].FixedAmount = 0

	t.VATs[3].Percentage = 0
	t.VATs[3].FixedAmount = 0
	err = db.DB.C("tickets").Insert(&t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	// Prepare the message that should be written to the FDM.
	msg := fdm.HashAndSignMsg("NS", t)
	log.Println(msg)

	res, err := FDM.Write(msg, true, 190)
	log.Println("finished writing to FDM")
	log.Println(res)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	json.NewEncoder(w).Encode("success")
}
