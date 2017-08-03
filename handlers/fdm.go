package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"pos-proxy/integrations/fdm"

	"github.com/gorilla/mux"
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
	// only used for payment
	Payments     []fdm.Payment `json:"payments"`
	ChangeAmount float64       `json:"change_amount"`
	IsClosed     bool          `json:"is_closed,omitempty"`
}

func FDMStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	f, err := fdm.New(vars["rcrs"])
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	defer func() {
		log.Println("closing connection with fdm")
		f.Close()
	}()
	res, err := f.CheckStatus()
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	if res.Error1 != "0" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode([]fdm.Response{res})
		return
	}
	json.NewEncoder(w).Encode([]fdm.Response{res})
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
	resp, err := FDM.CheckStatus()
	if err != nil {
		log.Println("Failed to get response from FDM")
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode([]fdm.Response{resp})
		return
	}
	req.Items = fixItemsPrice(req.Items)
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
		res, err := sendMessage("PS", FDM, req, items)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode([]fdm.Response{res})
			return
		}
	}
	// send negative msg
	items = splitItemsByVATRates(req.Items, negativeVATs)
	if len(items) > 0 {
		res, err := sendMessage("PR", FDM, req, items)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode([]fdm.Response{res})
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
		log.Println(err.Error())
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
	resp, err := FDM.CheckStatus()
	if err != nil {
		log.Println("Failed to get response from FDM")
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode([]fdm.Response{resp})
		return
	}
	req.Items = fixItemsPrice(req.Items)
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

	// prepare the array of FDM responses
	responses := []fdm.Response{}

	// send positive msg
	items := splitItemsByVATRates(req.Items, positiveVATs)
	if len(items) > 0 {
		res, err := sendMessage("PS", FDM, req, items)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode([]fdm.Response{res})
			return
		}

		responses = append(responses, res)
	}
	// send negative msg
	items = splitItemsByVATRates(req.Items, negativeVATs)
	if len(items) > 0 {
		res, err := sendMessage("PR", FDM, req, items)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode([]fdm.Response{res})
			return
		}

		responses = append(responses, res)
	}

	json.NewEncoder(w).Encode(responses)
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

	// check status
	resp, err := FDM.CheckStatus()
	if err != nil {
		log.Println("Failed to get response from FDM")
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode([]fdm.Response{resp})
		return
	}
	req.Items = fixItemsPrice(req.Items)
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

	// prepare the array of FDM responses
	responses := []fdm.Response{}

	// send positive msg
	items := splitItemsByVATRates(req.Items, positiveVATs)
	if len(items) > 0 {
		res, err := sendMessage("NS", FDM, req, items)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode([]fdm.Response{res})
			return
		}

		responses = append(responses, res)
	}
	// send negative msg
	items = splitItemsByVATRates(req.Items, negativeVATs)
	if len(items) > 0 {
		res, err := sendMessage("NR", FDM, req, items)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode([]fdm.Response{res})
			return
		}

		responses = append(responses, res)
	}

	json.NewEncoder(w).Encode(responses)
}
