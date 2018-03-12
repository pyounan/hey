package syncer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pos-proxy/callaccounting"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/models"
	"sync"
	"time"

	pb "gopkg.in/cheggaaa/pb.v1"
	"gopkg.in/mgo.v2/bson"
)

// FetchConfiguration asks CloudInn servers if the conf were updated,
// if yes update the current configurations and write them to the conf file
func FetchConfiguration() {
	fmt.Println("Checking for new configuration")
	uri := fmt.Sprintf("%s/api/pos/proxy/settings/", config.Config.BackendURI)
	netClient := helpers.NewNetClient()
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	req = helpers.PrepareRequestHeaders(req)
	response, err := netClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer response.Body.Close()
	incoming := config.ConfigHolder{}
	err = json.NewDecoder(response.Body).Decode(&incoming)
	if err != nil {
		log.Println(err.Error())
		return
	}
	if (config.Config.UpdatedAt != time.Time{}) && !incoming.UpdatedAt.After(config.Config.UpdatedAt) {
		return
	}
	fmt.Println("New configurations found")
	config.Config = &incoming
	// Write conf to file
	if err := config.Config.WriteToFile(); err != nil {
		log.Println(err.Error())
		return
	}
}

// Load data from the backend and insert to mongodb
func Load(apis map[string]string) {
	// Check if there are items in the requests queue,
	// if there is don't load new items until the requests queue is empty
	c, _ := db.DB.C("requests_queue").With(db.Session.Copy()).Find(nil).Count()
	if c > 0 {
		return
	}

	// create console progress bar
	fmt.Println("Pulling data from CloudInn services..")
	bar := pb.StartNew(len(apis))

	// create wait group
	wg := new(sync.WaitGroup)

	netClient := helpers.NewNetClient()
	for collection, api := range apis {
		wg.Add(1)
		go func(netClient *http.Client, collection string, api string) {
			if collection == "terminals" {
				models.TerminalsOperationsMutex.Lock()
				defer models.TerminalsOperationsMutex.Unlock()
			}
			uri := fmt.Sprintf("%s/%s", config.Config.BackendURI, api)
			req, err := http.NewRequest("GET", uri, nil)
			if err != nil {
				log.Println(err.Error())
			}
			req = helpers.PrepareRequestHeaders(req)
			response, err := netClient.Do(req)
			if err != nil {
				log.Println(err.Error())
				bar.Finish()
				wg.Done()
				return
			}
			defer response.Body.Close()
			bar.Increment()
			if response.StatusCode != 200 {
				log.Printf("Failed to load api from backend: %s\n", api)
				bar.Finish()
				wg.Done()
				return
			}
			// bar.Prefix(fmt.Sprintf("syncing %s from %s\n", collection, api))
			session := db.Session.Copy()
			defer session.Close()
			if collection == "sunexportdate" {
				db.DB.C(collection).With(session).Remove(nil)
				type BodyRequest struct {
					Dt string `json:"dt" bson:"dt"`
				}
				res := []BodyRequest{}
				json.NewDecoder(response.Body).Decode(&res)
				if len(res) > 0 {
					db.DB.C(collection).Insert(bson.M{"dt": res[0].Dt})
				}
			} else if collection == "terminals" {
				terminals := []models.Terminal{}
				json.NewDecoder(response.Body).Decode(&terminals)
				retrievedIDs := []int{}
				for _, terminal := range terminals {
					retrievedIDs = append(retrievedIDs, terminal.ID)
					// get old terminal
					t := models.Terminal{}
					err := db.DB.C(collection).With(session).Find(bson.M{"id": terminal.ID}).One(&t)
					if err != nil {
						// terminal not found, create a new one
						db.DB.C(collection).With(session).Insert(terminal)
					} else {
						// terminal already exists, check that the incoming last_invoice_id is larger
						// than the current one, update if true, otherwise continue to the next terminal.
						if t.LastInvoiceID < terminal.LastInvoiceID {
							db.DB.C(collection).With(session).Upsert(bson.M{"id": terminal.ID}, terminal)
						}
					}
				}
				// delete orphan terminals (terminals that have been deleted from backend
				err := db.DB.C(collection).With(session).Remove(bson.M{"id": bson.M{"$nin": retrievedIDs}})
				if err != nil {
					log.Println(err)
				}
			} else if api == "api/pos/posinvoices/?is_settled=false" {
				type Links struct {
					Next *string `json:"next"`
				}
				type Body struct {
					Results []models.Invoice `json:"results"`
					Links   Links            `json:"links"`
				}
				res := Body{}
				json.NewDecoder(response.Body).Decode(&res)
				for _, item := range res.Results {
					// Check if invoice is already settled, don't update it.
					oldInvoice := models.Invoice{}
					err := db.DB.C(collection).With(session).Find(bson.M{"invoice_number": item.InvoiceNumber}).One(&oldInvoice)
					// if older invoice was found, check if it is settled, then don't update it.
					if err == nil {
						if oldInvoice.IsSettled == true {
							continue
						}
					}

					_, err = db.DB.C(collection).With(session).Upsert(bson.M{"invoice_number": item.InvoiceNumber}, item)
					if err != nil {
						log.Println(err.Error())
					}
				}
				for res.Links.Next != nil {
					req, err = http.NewRequest("GET", *res.Links.Next, nil)
					if err != nil {
						log.Println(err.Error())
					}
					req = helpers.PrepareRequestHeaders(req)
					paginationresponse, err := netClient.Do(req)
					if err != nil {
						log.Println(err.Error())
						wg.Done()
						return
					}
					defer paginationresponse.Body.Close()
					if paginationresponse.StatusCode != 200 {
						log.Printf("Failed to load api from backend: %s\n", api)
						wg.Done()
						return
					}
					err = json.NewDecoder(paginationresponse.Body).Decode(&res)
					if err != nil {
						log.Println(err.Error())
					} else {
						for _, item := range res.Results {
							_, err = db.DB.C(collection).With(session).Upsert(bson.M{"invoice_number": item.InvoiceNumber}, item)
							if err != nil {
								log.Println(err.Error())
							}
						}
					}
				}
			} else if api == "shadowinn/api/auditdate/" {
				var res map[string]interface{}
				json.NewDecoder(response.Body).Decode(&res)
				oldAuditDate := make(map[string]interface{})
				err := db.DB.C(collection).With(session).Find(nil).One(&oldAuditDate)
				if err != nil {
					db.DB.C(collection).With(session).Insert(res)
				} else {
					err = db.DB.C(collection).With(session).Update(bson.M{"_id": oldAuditDate["_id"]}, bson.M{"$set": bson.M{"audit_date": res["audit_date"]}})
					if err != nil {
						log.Println(err.Error())
					}
				}
			} else if api == "shadowinn/api/company/" {
				var res map[string]interface{}
				err := json.NewDecoder(response.Body).Decode(&res)
				if err != nil {
					log.Println(err)
				} else {
					db.DB.C(collection).With(session).RemoveAll(nil)
					db.DB.C(collection).With(session).Upsert(bson.M{"name": res["name"].(string)}, res)
				}
			} else if api == "income/api/poscashierpermissions/" {
				var res []map[string]interface{}
				json.NewDecoder(response.Body).Decode(&res)
				for _, item := range res {
					delete(item, "_id")
					_, err = db.DB.C(collection).With(session).Upsert(bson.M{"poscashier_id": item["poscashier_id"]}, item)
					if err != nil {
						log.Println(err.Error())
					}
				}
			} else if api == "core/getallusergroups/" {
				var res []map[string]interface{}
				json.NewDecoder(response.Body).Decode(&res)
				retrievedIDs := []int{}
				for _, item := range res {
					id := int(item["id"].(float64))
					retrievedIDs = append(retrievedIDs, id)
					_, err := db.DB.C(collection).With(session).Upsert(bson.M{"id": id}, item)
					if err != nil {
						log.Println(err.Error())
					}
				}
				// remove extra records from the db (records that has been deleted from the backend,
				// should be deleted here too).
				err := db.DB.C(collection).With(session).Remove(bson.M{"id": bson.M{"$nin": retrievedIDs}})
				if err != nil {
					log.Println(err.Error())
				}
			} else {
				res := []map[string]interface{}{}
				err := json.NewDecoder(response.Body).Decode(&res)
				if err != nil {
					log.Println(err)
				}
				retrievedIDs := []int{}
				for _, item := range res {
					if _, ok := item["_id"]; ok {
						delete(item, "_id")
					}
					id := int(item["id"].(float64))
					retrievedIDs = append(retrievedIDs, id)
					_, err = db.DB.C(collection).With(session).Upsert(bson.M{"id": id}, item)
					if err != nil {
						log.Println(err.Error())
					}
				}
				// remove extra records from the db (records that has been deleted from the backend,
				// should be deleted here too).
				err = db.DB.C(collection).With(session).Remove(bson.M{"id": bson.M{"$nin": retrievedIDs}})
				if err != nil {
					log.Println(err.Error())
				}
			}
			wg.Done()
		}(netClient, collection, api)
	}
	wg.Wait()
	if !bar.IsFinished() {
		bar.FinishPrint("All Data has been loaded successfully")
	}
}

// FetchCallAccountingSettings sends a GET request to the backend to fetch the configuration
// of call accounting for this tenant
func FetchCallAccountingSettings() {
	uri := fmt.Sprintf("%s/api/clients/%d/settings/call_accounting", config.Config.BackendURI, config.Config.InstanceID)
	netClient := helpers.NewNetClient()
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	req = helpers.PrepareRequestHeaders(req)
	response, err := netClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		return
	}
	// TOFIX: check the response code and handle error
	defer response.Body.Close()
	data := callaccounting.Config{}
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		log.Println(err.Error())
		return
	}
	callaccounting.UpdateSettings(data)
}
