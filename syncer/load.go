package syncer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pos-proxy/config"
	"pos-proxy/helpers"
	"pos-proxy/pos/models"
	"pos-proxy/db"
	"os"
	"io/ioutil"
	"time"
	"gopkg.in/mgo.v2/bson"
)

// FetchConfiguration asks CloudInn servers if the conf were updated,
// if yes update the current configurations and write them to the conf file
func FetchConfiguration() {
	uri := fmt.Sprintf("%s/api/pos/proxy/settings/", config.Config.BackendURI)
	netClient := &http.Client{
		Timeout: time.Second * 5,
	}
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
	uri = fmt.Sprintf("%s/api/pos/fdm/", config.Config.BackendURI)
	req, err = http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	req = helpers.PrepareRequestHeaders(req)
	fdmResponse, err := netClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		return
	}
	// open configurations file
	f, err := os.Open("/etc/cloudinn/pos_config.json")
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer f.Close()
	data, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Println(err.Error())
		return
	}
	type ProxySettings struct {
		UpdatedAt string      `json:"updated_at"`
		FDMs      []config.FDMConfig `json:"fdms"`
	}
	dataStr := ProxySettings{}
	err = json.Unmarshal(data, &dataStr)
	if err != nil {
		log.Println(err.Error())
		return
	}
	t, err := time.Parse(time.RFC3339, fmt.Sprintf("%s", dataStr.UpdatedAt))
	if err != nil {
		log.Println(err.Error())
		return
	}
	// Check the configurations coming from the backend are newer than
	// the current configuration
	if (config.Config.UpdatedAt != time.Time{}) && !t.After(config.Config.UpdatedAt) {
		return
	}
	log.Println("New configurations found")
	config.Config.FDMs = dataStr.FDMs
	if len(config.Config.FDMs) > 0 {
		config.Config.IsFDMEnabled = true
	}
	config.Config.UpdatedAt = t
	type FDMSettingsResp struct {
		Data bool `json:"data"`
	}
	fdmSettingsResp := FDMSettingsResp{}
	json.NewDecoder(fdmResponse.Body).Decode(&fdmSettingsResp)
	config.Config.IsFDMEnabled = fdmSettingsResp.Data
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

	for collection, api := range backendApis {
		go func(collection string, api string) {
			var netClient = &http.Client{
				Timeout: time.Second * 10,
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
				return
			}
			if response.StatusCode != 200 {
				log.Printf("Failed to load api from backend: %s\n", api)
				return
			}
			defer response.Body.Close()
			log.Printf("-- syncing %s from %s\n", collection, api)
			if api == "api/pos/posinvoices/?is_settled=false" {
				type Body struct {
					Results []models.Invoice `json:"results"`
				}
				res := Body{}
				json.NewDecoder(response.Body).Decode(&res)
				for _, item := range res.Results {
					_, err = db.DB.C(collection).Upsert(bson.M{"invoice_number": item.InvoiceNumber}, item)
					if err != nil {
						log.Println(err.Error())
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
			} else {
				var res []map[string]interface{}
				json.NewDecoder(response.Body).Decode(&res)
				for _, item := range res {
					delete(item, "_id")
					_, err = db.DB.C(collection).Upsert(bson.M{"id": item["id"]}, item)
					if err != nil {
						log.Println(err.Error())
					}
				}
			}
		}(collection, api)
	}
}
