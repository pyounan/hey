package syncer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pos-proxy/config"
	"pos-proxy/db"
	"time"
)

// Load data from the backend and insert to mongodb
func Load() {
	backendApis := make(map[string]string)
	backendApis["stores"] = "api/pos/store/"
	backendApis["fdm"] = "api/pos/fdm/"
	backendApis["proxy_settings"] = "api/pos/proxy/settings/"
	backendApis["fixeddiscounts"] = "api/pos/fixeddiscount/"
	backendApis["storedetails"] = "api/pos/storedetails/"
	backendApis["tables"] = "api/pos/tables/"
	backendApis["posinvoices"] = "api/pos/posinvoices/?is_settled=false"
	backendApis["terminals"] = "api/pos/terminal/"
	backendApis["condiments"] = "api/pos/condiment/"
	backendApis["courses"] = "api/pos/course/"
	backendApis["printers"] = "api/pos/printer/"
	backendApis["printersettings"] = "api/pos/printersettings/"

	backendApis["reservations"] = "shadowinn/api/reservations/"
	backendApis["rooms"] = "shadowinn/api/rooms/"
	backendApis["users"] = "shadowinn/api/user/"
	backendApis["company"] = "shadowinn/api/company/"
	backendApis["audit_date"] = "shadowinn/api/auditdate/"

	backendApis["departments"] = "income/api/department/"
	backendApis["currencies"] = "income/api/currency/"
	backendApis["permissions"] = "income/api/poscashierpermissions/"
	backendApis["cashiers"] = "income/api/cashier/sync/"

	for collection, api := range backendApis {
		go func(collection string, api string) {
			var netClient = &http.Client{
				Timeout: time.Second * 5,
			}
			uri := fmt.Sprintf("%s/%s", config.Config.BackendURI, api)
			req, err := http.NewRequest("GET", uri, nil)
			if err != nil {
				log.Println(err.Error())
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("JWT %s", config.ProxyToken))
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
				var res map[string]interface{}
				json.NewDecoder(response.Body).Decode(&res)
				for _, item := range res["results"].([]interface{}) {
					err = db.DB.C(collection).Insert(item)
					if err != nil {
						log.Println(err.Error())
					}
				}
			} else {
				var res []map[string]interface{}
				json.NewDecoder(response.Body).Decode(&res)
				for _, item := range res {
					err = db.DB.C(collection).Insert(item)
					if err != nil {
						log.Println(err.Error())
					}
				}
			}
		}(collection, api)
	}
}
