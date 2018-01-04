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
	defer response.Body.Close()
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
	db.DB.C("currencies").With(db.Session.Copy()).Find(bson.M{"is_default": true}).One(&defaultCurrency)
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
			if transaction.SunMapping.Code != "" {
				number = transaction.SunMapping.Code
			}
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
			} else if exportType == "receiving" {
				journalType = "CIINV"
			} else if exportType == "posinvoice" {
				journalType = "CIPOS"
			}

			transactionRef := strings.Title(exportType)
			description := jv.Description
			if transaction.Description != "" {
				description = transaction.Description
			}
			description = fmt.Sprintf("%-18s", description)
			description = description[0:18]
			sep2 := strings.Repeat(" ", 2)
			sep5 := strings.Repeat(" ", 5)
			sep8 := strings.Repeat(" ", 8)
			sep14 := strings.Repeat(" ", 14)
			sep15 := strings.Repeat(" ", 15)
			sep18 := strings.Repeat(" ", 18)
			sep46 := strings.Repeat(" ", 46)

			line := []string{}
			line = append(line, fmt.Sprintf("%-10s", number))
			line = append(line, sep5)
			line = append(line, glPeriodYear)
			line = append(line, "0")
			line = append(line, fmt.Sprintf("%02d", glPeriodMonth))
			line = append(line, t.Format("20060102"))
			line = append(line, sep2)
			line = append(line, "L")
			line = append(line, sep14)
			line = append(line, fmt.Sprintf("%018d", amountInt))
			line = append(line, trType)
			line = append(line, " ")
			line = append(line, journalType)
			line = append(line, sep5) // Journal Source
			line = append(line, fmt.Sprintf("%-15s", transactionRef))
			line = append(line, fmt.Sprintf("%06d-", jv.OperationId))
			line = append(line, description)
			line = append(line, sep15)                                        //Space 15
			line = append(line, sep8)                                         // Due date
			line = append(line, sep46)                                        // Space 46
			line = append(line, fmt.Sprintf("%-5s", defaultCurrency["code"])) // Currency
			line = append(line, sep18)                                        // Conversion Rate
			line = append(line, sep18)                                        // Amount FC
			line = append(line, sep14)                                        // Space 14
			line = append(line, fmt.Sprintf("%-15s", transaction.SunMapping.TCode1))
			line = append(line, fmt.Sprintf("%-15s", transaction.SunMapping.TCode2))
			line = append(line, fmt.Sprintf("%-15s", transaction.SunMapping.TCode3))
			line = append(line, fmt.Sprintf("%-15s", transaction.SunMapping.TCode4))
			line = append(line, fmt.Sprintf("%-15s", transaction.SunMapping.TCode5))
			line = append(line, fmt.Sprintf("%-15s", transaction.SunMapping.TCode6))
			line = append(line, fmt.Sprintf("%-15s", transaction.SunMapping.TCode7))
			line = append(line, fmt.Sprintf("%-15s", transaction.SunMapping.TCode8))
			line = append(line, fmt.Sprintf("%-15s", transaction.SunMapping.TCode9))
			line = append(line, fmt.Sprintf("%-15s", transaction.SunMapping.TCode10))
			line = append(line, "\r\n")
			lineStr := strings.Join(line, "")
			_, err = f.WriteString(lineStr)
			if err != nil {
				log.Println("Failed to write to file", fileName)
				return err
			}
		}
	}
	return nil
}

func ImportJournalVouchers(w http.ResponseWriter, r *http.Request) {
	lastDate := make(map[string]string)
	db.DB.C("sunexportdate").With(db.Session.Copy()).Find(nil).One(&lastDate)
	log.Println("Before return from post", lastDate["dt"])
	layout := "2006-01-02"
	if r.Method == "GET" {
		templateexport.ExportedTemplates.ExecuteTemplate(w, "export_to_sun",
			bson.M{"lastDate": lastDate["dt"]})
	} else {

		dt := r.PostFormValue("dt")
		dtParsed, err := time.Parse(layout, dt)
		if err != nil {
			templateexport.ExportedTemplates.ExecuteTemplate(w, "export_to_sun",
				bson.M{"message": "Invalid date",
					"lastDate": lastDate["dt"]})
			return
		}
		lastDateParsed, _ := time.Parse(layout, lastDate["dt"])
		if dtParsed.Before(lastDateParsed) {
			templateexport.ExportedTemplates.ExecuteTemplate(w, "export_to_sun",
				bson.M{"message": "Export date cannot be less than last export date",
					"lastDate": lastDate["dt"]})
			return
		}

		fetchedJVS, err := FetchJournalVouchers(dt)
		if err != nil {
			templateexport.ExportedTemplates.ExecuteTemplate(w, "export_to_sun",
				bson.M{"message": fmt.Sprintf("Error: %s", err)})
			return
		}
		if len(fetchedJVS) == 0 {
			templateexport.ExportedTemplates.ExecuteTemplate(w, "export_to_sun",
				bson.M{"message": fmt.Sprintf("Error: %s", "No JVs were fetched")})
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
		db.DB.C("sunexportdate").With(db.Session.Copy()).Remove(nil)
		db.DB.C("sunexportdate").With(db.Session.Copy()).Insert(bson.M{"dt": dt})
		syncer.QueueRequest("/api/inventory/sunexportdate/", "POST",
			r.Header, bson.M{"dt": dt})
		log.Println("After after return from post", dt)
		templateexport.ExportedTemplates.ExecuteTemplate(w, "export_to_sun",
			bson.M{"message": "Success", "lastDate": dt})
	}
}
