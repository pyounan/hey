package syncer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/models"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type RequestRow struct {
	ID         bson.ObjectId `bson:"_id,omitempty"`
	URI        string        `bson:"uri"`
	Method     string        `bson:"method"`
	Headers    http.Header   `bson:"headers"`
	Payload    interface{}   `bson:"payload"`
	ActionTime time.Time     `bson:"action_time"`
}

// QueueRequest insert a request object to a queue that syncs with the backend
func QueueRequest(uri string, method string, headers http.Header, payload interface{}) error {
	body := &RequestRow{}
	body.ID = bson.NewObjectId()
	body.URI = uri
	body.Method = method
	body.Headers = headers
	body.Payload = payload
	body.ActionTime = time.Now()
	log.Println("inserting request to queue")
	err := db.DB.C("requests_queue").Insert(body)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func PushToBackend() {
	requests := []RequestRow{}
	db.DB.C("requests_queue").Find(nil).Sort("action_time").All(&requests)

	netClient := helpers.NewNetClient()
	for _, r := range requests {
		uri := fmt.Sprintf("%s%s", config.Config.BackendURI, r.URI)

		payload := new(bytes.Buffer)
		json.NewEncoder(payload).Encode(r.Payload)
		req, err := http.NewRequest(r.Method, uri, payload)
		if err != nil {
			log.Println(err.Error())
			return
		}
		req.Header = r.Headers
		log.Println(req.Header)
		req = helpers.PrepareRequestHeaders(req)
		response, err := netClient.Do(req)
		if err != nil {
			log.Println(err.Error())
			return
		}
		defer response.Body.Close()
		log.Println("Sending: ", r.Method, req.URL.Path)
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			log.Println(response, "Failed to fetch response from backend")
			res, _ := ioutil.ReadAll(response.Body)
			log.Println("error response body", string(res))
			return
		}
		if req.URL.Path == "/api/pos/posinvoices/" || strings.Contains(req.URL.Path, "createpostings") {
			type RespBody struct {
				Invoice models.Invoice `json:"posinvoice" bson:"posinvoice"`
			}
			res := RespBody{}
			json.NewDecoder(response.Body).Decode(&res)
			_, err = db.DB.C("posinvoices").Upsert(bson.M{"invoice_number": res.Invoice.InvoiceNumber}, res.Invoice)
			if err != nil {
				log.Println(err.Error())
				return
			}
		} else if strings.Contains(req.URL.Path, "changetable") || strings.Contains(req.URL.Path, "split") {
			res := []models.Invoice{}
			json.NewDecoder(response.Body).Decode(&res)
			for _, inv := range res {
				_, err = db.DB.C("posinvoices").Upsert(bson.M{"invoice_number": inv.InvoiceNumber}, inv)
				if err != nil {
					log.Println(err.Error())
					return
				}
			}
		} else if strings.Contains(req.URL.Path, "folio") {
			res := models.Invoice{}
			json.NewDecoder(response.Body).Decode(&res)
			_, err = db.DB.C("posinvoices").Upsert(bson.M{"invoice_number": res.InvoiceNumber}, res)
			if err != nil {
				log.Println(err.Error())
				return
			}
		} else if strings.Contains(req.URL.Path, "refund") {
			type RespBody struct {
				NewInvoice      models.Invoice `json:"new_posinvoice" bson:"new_posinvoice"`
				OriginalInvoice models.Invoice `json:"original_posinvoice" bson:"original_posinvoice"`
			}
			res := RespBody{}
			json.NewDecoder(response.Body).Decode(&res)
			_, err = db.DB.C("posinvoices").Upsert(bson.M{"invoice_number": res.OriginalInvoice.InvoiceNumber}, res.OriginalInvoice)
			if err != nil {
				log.Println(err.Error())
				return
			}
			_, err = db.DB.C("posinvoices").Upsert(bson.M{"invoice_number": res.NewInvoice.InvoiceNumber}, res.NewInvoice)
			if err != nil {
				log.Println(err.Error())
				return
			}
		}
		err = db.DB.C("requests_queue").Remove(bson.M{"_id": r.ID})
		if err != nil {
			log.Println(err.Error())
			return
		}
	}
}
