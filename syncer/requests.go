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
	"pos-proxy/logging"
	"pos-proxy/pos/models"
	"pos-proxy/proxy"
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

type RequestLog struct {
	ID             bson.ObjectId `json:"id" bson:"_id"`
	Request        RequestRow    `json:"request_row" bson:"request_row"`
	ResponseBody   interface{}   `json:"response_body" bson:"response_body"`
	ResponseStatus int           `json:"response_status" bson:"response_status"`
	CreatedAt      time.Time     `json:"created_at" bson:"created_at"`
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
		req = helpers.PrepareRequestHeaders(req)
		// log.Println("Openning connection for", req.URL)
		// add this request and its response to requests log
		logRecord := RequestLog{}
		logRecord.ID = bson.NewObjectId()
		logRecord.CreatedAt = time.Now()
		logRecord.Request = r
		log.Println("Sending: ", r.Method, req.URL.Path)
		response, err := netClient.Do(req)
		if err != nil {
			log.Println(err.Error())
			return
		}
		defer func() {
			// log.Println("Closing connection of", req.URL)
			response.Body.Close()
		}()
		log.Println(response.Body)
		res, _ := ioutil.ReadAll(response.Body)
		logRecord.ResponseStatus = response.StatusCode
		logRecord.ResponseBody = fmt.Sprintf("%s", res)
		log.Println("Response Content-Type", response.Header["Content-Type"])
		logging.Debug(fmt.Sprintf("response body %s\n", res))
		err = db.DB.C("requests_log").Insert(logRecord)
		if err != nil {
			log.Println("Failed to queue failure to log", err.Error())
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			log.Println(response, "Failed to fetch response from backend")
			if response.StatusCode >= 400 && response.StatusCode <= 500 {
				proxy.AllowIncomingRequests = false
			}
			return
		}
		response.Body = ioutil.NopCloser(bytes.NewBuffer(res))
		proxy.AllowIncomingRequests = true
		if req.URL.Path == "/api/pos/posinvoices/" {
			res := models.Invoice{}
			err := json.NewDecoder(response.Body).Decode(&res)
			if err != nil {
				log.Println("decoding error", err.Error())
			}
			_, err = db.DB.C("posinvoices").Upsert(bson.M{"invoice_number": res.InvoiceNumber}, res)
			if err != nil {
				log.Println(err.Error())
				return
			}
		} else if strings.Contains(req.URL.Path, "createpostings") {
			type RespBody struct {
				Invoice models.Invoice `json:"posinvoice" bson:"posinvoice"`
			}
			res := RespBody{}
			err := json.NewDecoder(response.Body).Decode(&res)
			if err != nil {
				log.Println("decoding error", err.Error())
			}
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
				NewInvoice      models.Invoice `json:"new_invoice" bson:"new_invoice"`
				OriginalInvoice models.Invoice `json:"original_invoice" bson:"original_invoice"`
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
