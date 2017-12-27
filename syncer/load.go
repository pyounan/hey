package syncer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/logging"
	"pos-proxy/pos/models"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// FetchConfiguration asks CloudInn servers if the conf were updated,
// if yes update the current configurations and write them to the conf file
func FetchConfiguration() {
	log.Println("Checking for new configuration")
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
	log.Println("New configurations found")
	config.Config = &incoming
	// Write conf to file
	if err := config.Config.WriteToFile(); err != nil {
		log.Println(err.Error())
		return
	}
}

// Load data from the backend and insert to mongodb
func Load() {
	// Check if there are items int he requests queue,
	// if there is don't load new items until the requests queue is empty
	c, _ := db.DB.C("requests_queue").Find(nil).Count()
	if c > 0 {
		return
	}
	backendApis := make(map[string]string)
	backendApis["stores"] = "api/pos/store/"
	backendApis["fixeddiscounts"] = "api/pos/fixeddiscount/"
	backendApis["storedetails"] = "api/pos/storedetails/"
	backendApis["tables"] = "api/pos/tables/"
	backendApis["posinvoices"] = "api/pos/posinvoices/?is_settled=false"
	backendApis["terminals"] = "api/pos/terminal/"
	backendApis["condiments"] = "api/pos/condiment/"
	backendApis["courses"] = "api/pos/course/"
	backendApis["printers"] = "api/pos/printer/"
	backendApis["printersettings"] = "api/pos/printersettings/"

	backendApis["company"] = "shadowinn/api/company/"
	backendApis["audit_date"] = "shadowinn/api/auditdate/"

	backendApis["departments"] = "income/api/department/"
	backendApis["currencies"] = "income/api/currency/"
	backendApis["permissions"] = "income/api/poscashierpermissions/"
	backendApis["cashiers"] = "income/api/cashier/sync/"
	backendApis["attendance"] = "income/api/attendance/"
	backendApis["usergroups"] = "core/getallusergroups/"
	backendApis["operasettings"] = "api/pos/opera/"

	backendApis["sunexportdate"] = "api/inventory/sunexportdate/"

	netClient := helpers.NewNetClient()
	for collection, api := range backendApis {
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
				logging.Error(err.Error())
				return
			}
			defer response.Body.Close()
			if response.StatusCode != 200 {
				logging.Error(fmt.Sprintf("Failed to load api from backend: %s\n", api))
				return
			}
			logging.Info(fmt.Sprintf("syncing %s from %s", collection, api))
			if collection == "sunexportdate" {
				db.DB.C(collection).Remove(nil)
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
				for _, terminal := range terminals {
					// get old terminal
					t := models.Terminal{}
					err := db.DB.C(collection).Find(bson.M{"id": terminal.ID}).One(&t)
					if err != nil {
						// terminal not found, create a new one
						db.DB.C(collection).Insert(terminal)
					} else {
						// terminal already exists, check that the incoming last_invoice_id is larger
						// than the current one, update if true, otherwise continue to the next terminal.
						if t.LastInvoiceID < terminal.LastInvoiceID {
							db.DB.C(collection).Upsert(bson.M{"id": terminal.ID}, terminal)
						}
					}
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
					err := db.DB.C(collection).Find(bson.M{"invoice_number": item.InvoiceNumber}).One(&oldInvoice)
					// if older invoice was found, check if it is settled, then don't update it.
					if err == nil {
						if oldInvoice.IsSettled == true {
							continue
						}
					}

					_, err = db.DB.C(collection).Upsert(bson.M{"invoice_number": item.InvoiceNumber}, item)
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
						return
					}
					defer paginationresponse.Body.Close()
					if paginationresponse.StatusCode != 200 {
						log.Printf("Failed to load api from backend: %s\n", api)
						return
					}
					json.NewDecoder(paginationresponse.Body).Decode(&res)
					for _, item := range res.Results {
						_, err = db.DB.C(collection).Upsert(bson.M{"invoice_number": item.InvoiceNumber}, item)
						if err != nil {
							log.Println(err.Error())
						}
					}
				}
			} else if api == "shadowinn/api/auditdate/" {
				var res map[string]interface{}
				json.NewDecoder(response.Body).Decode(&res)
				oldAuditDate := make(map[string]interface{})
				err := db.DB.C(collection).Find(nil).One(&oldAuditDate)
				if err != nil {
					db.DB.C(collection).Insert(res)
				} else {
					err = db.DB.C(collection).Update(bson.M{"_id": oldAuditDate["_id"]}, bson.M{"$set": bson.M{"audit_date": res["audit_date"]}})
					if err != nil {
						log.Println(err.Error())
					}
				}
			} else if api == "shadowinn/api/company/" {
				var res map[string]interface{}
				json.NewDecoder(response.Body).Decode(&res)
				_, err := db.DB.C(collection).Upsert(bson.M{"name": res["name"]}, res)
				if err != nil {
					db.DB.C(collection).Insert(res)
				}
			} else if api == "income/api/poscashierpermissions/" {
				var res []map[string]interface{}
				json.NewDecoder(response.Body).Decode(&res)
				for _, item := range res {
					delete(item, "_id")
					_, err = db.DB.C(collection).Upsert(bson.M{"poscashier_id": item["poscashier_id"]}, item)
					if err != nil {
						log.Println(err.Error())
					}
				}
			} else if api == "core/getallusergroups/" {
				var res []map[string]interface{}
				json.NewDecoder(response.Body).Decode(&res)
				for _, item := range res {
					_, err := db.DB.C(collection).Upsert(bson.M{"id": item["id"]}, item)
					if err != nil {
						log.Println(err.Error())
					}
				}
			} else {
				res := []map[string]interface{}{}
				err := json.NewDecoder(response.Body).Decode(&res)
				if err != nil {
					log.Println(err)
				}
				for _, item := range res {
					if _, ok := item["_id"]; ok {
						delete(item, "_id")
					}
					_, err = db.DB.C(collection).Upsert(bson.M{"id": item["id"]}, item)
					if err != nil {
						log.Println(err.Error())
					}
				}
			}
		}(netClient, collection, api)
	}
}
