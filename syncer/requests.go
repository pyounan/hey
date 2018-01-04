package syncer

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/models"
	"pos-proxy/proxy"
	"strings"
	"time"

	pb "gopkg.in/cheggaaa/pb.v1"
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
	err := db.DB.C("requests_queue").With(db.Session.Copy()).Insert(body)
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
	db.DB.C("requests_queue").With(db.Session.Copy()).Find(nil).Sort("action_time").All(&requests)

	// return if there is no something to do
	if len(requests) == 0 {
		return
	}

	fmt.Printf("Pushing %d requests to CloudInn services...\n", len(requests))
	bar := pb.StartNew(len(requests))

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

		// bar.Postfix(fmt.Sprintf("Sending: %s %s", r.Method, req.URL.Path))
		response, err := netClient.Do(req)
		if err != nil {
			bar.Finish()
			log.Println("Error", err.Error())
			return
		}
		defer response.Body.Close()

		logRecord.ResponseStatus = response.StatusCode
		if response.Header.Get("Content-Encoding") == "gzip" {
			response.Body, err = gzip.NewReader(response.Body)
			if err != nil {
				log.Println("Warning:", err.Error())
			} else {
				defer response.Body.Close()
			}
		}
		// Check status code of the response, continue or abort based on that
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			log.Println("Error: Failed to fetch response from backend")
			if response.StatusCode >= 400 && response.StatusCode <= 500 {
				log.Println("Fatal Error: Proxy operations is going to be halted, please contact support.")
				proxy.AllowIncomingRequests = false
			}
			data, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Println("Warning:", err.Error())
			} else {
				logRecord.ResponseBody = string(data)
			}
			err = db.DB.C("requests_log").With(db.Session.Copy()).Insert(logRecord)
			if err != nil {
				log.Println("Warning: failed to log syncer record", err.Error())
			}
			return
		}
		proxy.AllowIncomingRequests = true
		if req.URL.Path == "/api/pos/posinvoices/" || req.URL.Path == "/api/pos/posinvoices/houseuse/" ||
			strings.Contains(req.URL.Path, "folio") || strings.Contains(req.URL.Path, "wasteandvoid") {
			res := models.Invoice{}
			err := json.NewDecoder(response.Body).Decode(&res)
			if err != nil {
				log.Println("Warning:", err.Error())
			} else {
				_, err = db.DB.C("posinvoices").With(db.Session.Copy()).Upsert(bson.M{"invoice_number": res.InvoiceNumber}, res)
				if err != nil {
					log.Println("Warning:", err.Error())
				}
				logRecord.ResponseBody = res
			}
		} else if strings.Contains(req.URL.Path, "createpostings") {
			type RespBody struct {
				Invoice models.Invoice `json:"posinvoice" bson:"posinvoice"`
			}
			res := RespBody{}
			err := json.NewDecoder(response.Body).Decode(&res)
			if err != nil {
				log.Println("Warning:", err.Error())
			} else {
				_, err = db.DB.C("posinvoices").With(db.Session.Copy()).Upsert(bson.M{"invoice_number": res.Invoice.InvoiceNumber}, res.Invoice)
				if err != nil {
					log.Println("Warning:", err.Error())
				}
				logRecord.ResponseBody = res
			}
		} else if strings.Contains(req.URL.Path, "cancelpostings") {
			res := []models.Posting{}
			err := json.NewDecoder(response.Body).Decode(&res)
			if err != nil {
				log.Println("Warning:", err.Error())
			} else {
				invoiceNumber := strings.Split(req.URL.Path, "/")[3]
				_, err = db.DB.C("posinvoices").With(db.Session.Copy()).Upsert(bson.M{"invoice_number": invoiceNumber},
					bson.M{"$set": bson.M{"pospayment": res}})
				if err != nil {
					log.Println("Warning:", err.Error())
				}
				logRecord.ResponseBody = res
			}
		} else if strings.Contains(req.URL.Path, "changetable") || strings.Contains(req.URL.Path, "split") ||
			strings.Contains(req.URL.Path, "bulksubmit") {
			res := []models.Invoice{}
			err := json.NewDecoder(response.Body).Decode(&res)
			if err != nil {
				log.Println("Warning: ", err.Error())
			} else {
				for _, inv := range res {
					_, err = db.DB.C("posinvoices").With(db.Session.Copy()).Upsert(bson.M{"invoice_number": inv.InvoiceNumber}, inv)
					if err != nil {
						log.Println("Warning:", err.Error())
					}
				}
				logRecord.ResponseBody = res
			}
		} else if strings.Contains(req.URL.Path, "refund") {
			type RespBody struct {
				NewInvoice models.Invoice `json:"new_posinvoice" bson:"new_posinvoice"`
			}
			res := RespBody{}
			err := json.NewDecoder(response.Body).Decode(&res)
			if err != nil {
				log.Println("Warning: ", err.Error())
			} else {
				_, err = db.DB.C("posinvoices").With(db.Session.Copy()).Upsert(bson.M{"invoice_number": res.NewInvoice.InvoiceNumber}, res.NewInvoice)
				if err != nil {
					log.Println("Warning:", err.Error())
				}
				logRecord.ResponseBody = res
			}
		}
		err = db.DB.C("requests_log").With(db.Session.Copy()).Insert(logRecord)
		if err != nil {
			log.Println("Warning: failed to log syncer record", err.Error())
		}
		err = db.DB.C("requests_queue").With(db.Session.Copy()).Remove(bson.M{"_id": r.ID})
		if err != nil {
			bar.Finish()
			log.Println("Error: failed to remove request from queue after success", err.Error())
			return
		}
		bar.Increment()
	}
	bar.Finish()
}
