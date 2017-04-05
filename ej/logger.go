package ej

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/fdm"
)

func Log(event_label string, ticket fdm.Ticket, response map[string]interface{}) error {
	record := make(map[string]interface{})
	record["action_time"] = ticket.ActionTime
	changes := make(map[string]interface{})
	changes["event_label"] = event_label
	changes["fdm_response"] = response
	changes["cashier_name"] = ticket.CashierName
	changes["cashier_number"] = ticket.CashierNumber
	changes["invoice_number"] = ticket.InvoiceNumber
	changes["terminal_name"] = ticket.TerminalName
	changes["ticket_number"] = ticket.TicketNumber
	changes["ticket_datetime"] = ticket.ActionTime
	changes["total_amount"] = ticket.TotalAmount
	changes["items"] = ticket.Items
	changes["vats"] = ticket.VATs
	changes["rcrs"] = ticket.RCRS
	changes["plu_hash"] = ticket.PLUHash
	changes["change_type"] = "event"
	// calculate totals summary
	type RateSummary map[string]float64
	summary := make(map[string]RateSummary)
	totals := make(map[string]float64)
	totals["net_amount"] = 0
	totals["vat_amount"] = 0
	totals["taxable_amount"] = 0
	rates := []string{"A", "B", "C", "D"}
	for _, r := range rates {
		summary[r] = RateSummary{}
		summary[r]["net_amount"] = 0
		summary[r]["vat_amount"] = 0
		summary[r]["taxable_amount"] = 0
	}
	for _, item := range ticket.Items {
		summary[item.VAT]["net_amount"] += item.NetAmount
		summary[item.VAT]["vat_amount"] += item.NetAmount * item.VATPercentage / 100
		summary[item.VAT]["taxable_amount"] += item.Price
		totals["net_amount"] += item.NetAmount
		totals["vat_amount"] += item.NetAmount * item.VATPercentage / 100
		totals["taxable_amount"] += item.Price
	}
	changes["summary"] = summary
	changes["totals"] = totals
	if ticket.TableNumber != "" {
		changes["table_number"] = ticket.TableNumber
	} else {
		changes["table_number"] = "takeout"
	}

	record["changes"] = changes
	err := db.DB.C("ej").Insert(record)
	if err != nil {
		return err
	}
	return nil
}

func PushToBackend() {
	var records []map[string]interface{}
	_ = db.DB.C("ej").Find(bson.M{}).All(&records)

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	log.Printf("%d ej records found\n", len(records))

	for _, r := range records {
		go func(netClient *http.Client, DB *mgo.Database, record map[string]interface{}) {
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(r)
			uri := fmt.Sprintf("%s/api/pos/ej/?tenant_id=%s", config.Config.BackendURI, config.Config.TenantID)
			req, err := http.NewRequest("POST", uri, b)
			if err != nil {
				log.Println(err)
			}
			req.Header.Set("Content-Type", "application/json")
			response, err := netClient.Do(req)
			if err != nil {
				log.Println(err)
			}
			if response != nil {
				err := DB.C("ej").RemoveId(r["_id"].(bson.ObjectId))
				if err != nil {
					log.Println(err)
				}

				log.Println("success")
			}
		}(netClient, db.DB, r)
	}

}
