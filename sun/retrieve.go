package sun

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/syncer"
	"pos-proxy/templateexport"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

func FetchJournalVouchers(dt string) ([]JournalVoucher, error) {
	jvs := []JournalVoucher{}
	apiURL := fmt.Sprintf("%s/api/inventory/journalvoucher/", config.Config.BackendURI)
	req, _ := http.NewRequest("GET", apiURL, nil)
	netClient := helpers.NewNetClient()
	req = helpers.PrepareRequestHeaders(req)
	q := req.URL.Query()
	q.Add("dt", dt)
	req.URL.RawQuery = q.Encode()
	response, err := netClient.Do(req)
	if err != nil {
		log.Println("Failed to fetch jvs", err)
		return jvs, err
	}
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Failed to fetch jvs", err)
		return jvs, err
	}
	err = json.Unmarshal(respBody, &jvs)
	if err != nil {
		log.Println("Failed to fetch jvs", err)
		return jvs, err
	}
	return jvs, nil
}

func Serialize(jvs []JournalVoucher, exportType string) error {
	t := time.Now()
	fileName := fmt.Sprintf("/media/share/%s_%s.ndf", exportType, t.Format("20060102150405"))
	f, err := os.Create(fileName)
	if err != nil {
		log.Println("Failed to create file", fileName)
		return err
	}
	defer f.Close()
	f.WriteString("VERSION                         42601\r\n")
	layout := "2006-01-02T15:04:05Z"
	defaultCurrency := make(map[string]interface{})
	db.DB.C("currencies").Find(bson.M{"is_default": true}).One(&defaultCurrency)
	for _, jv := range jvs {
		splitted := strings.Split(jv.GLPeriod, "-")
		glPeriodMonth, err := strconv.Atoi(splitted[0])
		if err != nil {
			log.Println("Failed to parse gl period month", jv.GLPeriod)
			return err
		}
		glPeriodYear := splitted[1]
		t, err = time.Parse(layout, jv.Dt)
		if err != nil {
			log.Println("Failed to parse JV date", jv.Dt)
			return err
		}
		for _, transaction := range jv.Transactions {
			number := transaction.Account["number"].(string)
			number = strings.ToUpper(number)
			amount := transaction.Amount * 1000
			convertedFloat, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", amount), 64)
			convertedFloat *= 1000
			amountInt := int64(convertedFloat)
			trType := "C"
			if transaction.TransactionType == "debit" {
				trType = "D"
			}
			journalType := "CICON"
			if exportType == "transfer" {
				journalType = "CITRS"
			} else if exportType == "invoice" {
				journalType = "CIINV"
			}

			transactionRef := strings.Title(exportType)
			description := jv.Description
			if transaction.Description != "" {
				description = transaction.Description
			}
			description = fmt.Sprintf("%-18s", description)
			description = description[0:18]
			sep14 := strings.Repeat(" ", 14)
			sep46 := strings.Repeat(" ", 46)
			sep2 := strings.Repeat(" ", 2)
			sep5 := strings.Repeat(" ", 5)
			sep15 := strings.Repeat(" ", 15)
			sep8 := strings.Repeat(" ", 8)
			sep201 := strings.Repeat(" ", 201)
			line := fmt.Sprintf("%-15s%s0%02d%s%sL%s%018d%s %s%s%-15s%06d-%s%s%s%s%-5s%s\r\n",
				number, glPeriodYear, glPeriodMonth, t.Format("20060102"), sep2,
				sep14, amountInt, trType, journalType, sep5, transactionRef, jv.ID,
				description, sep15, sep8, sep46, defaultCurrency["code"], sep201)
			_, err = f.WriteString(line)
			if err != nil {
				log.Println("Failed to write to file", fileName)
				return err
			}
		}
	}
	return nil
}

func ImportJournalVouchers(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templateexport.ExportedTemplates.ExecuteTemplate(w, "export_to_sun", nil)
	} else {
		/*err := r.ParseForm()
		if err != nil{
			   panic(err)
		}*/

		dt := r.PostFormValue("dt")
		fetchedJVS, err := FetchJournalVouchers(dt)
		if err != nil {
			templateexport.ExportedTemplates.ExecuteTemplate(w, "export_to_sun",
				bson.M{"message": fmt.Sprintf("Error: %s", err)})
			return
		}

		groupedJVS := make(map[string][]JournalVoucher)
		for _, m := range fetchedJVS {
			groupedJVS[m.OperationType] = append(groupedJVS[m.OperationType], m)
		}

		for k, v := range groupedJVS {
			err = Serialize(v, k)
			if err != nil {
				templateexport.ExportedTemplates.ExecuteTemplate(w, "export_to_sun",
					bson.M{"message": fmt.Sprintf("Error: %s", err)})
				return
			}
		}
		syncer.QueueRequest("/api/inventory/exportsundate/", "POST",
			r.Header, bson.M{"dt": dt})
		templateexport.ExportedTemplates.ExecuteTemplate(w, "export_to_sun",
			bson.M{"message": "Success"})
	}
}
