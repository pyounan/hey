package sun

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"strconv"
	"strings"
	"time"
)

func FetchJournalVouchers(exportType string) []JournalVoucher {
	apiURL := fmt.Sprintf("%s/api/inventory/journalvoucher/?type=%s", config.Config.BackendURI, exportType)
	req, _ := http.NewRequest("GET", apiURL, nil)
	netClient := helpers.NewNetClient()
	req = helpers.PrepareRequestHeaders(req)
	response, _ := netClient.Do(req)
	respBody, _ := ioutil.ReadAll(response.Body)
	jvs := []JournalVoucher{}
	_ = json.Unmarshal(respBody, &jvs)
	return jvs
}

func Serialize(jvs []JournalVoucher, exportType string) string {
	t := time.Now()
	fileName := fmt.Sprintf("/tmp/%s_%s.ndf", exportType, t.Format("20060102150405"))
	f, _ := os.Create(fileName)
	defer f.Close()
	f.WriteString("VERSION                         42601\n")
	layout := "2006-01-02T15:04:05.000000Z"
	defaultCurrency := make(map[string]interface{})
	db.DB.C("currencies").Find(bson.M{"is_default": true}).One(&defaultCurrency)
	log.Println("Currency", defaultCurrency)
	for _, jv := range jvs {
		splitted := strings.Split(jv.GLPeriod, "-")
		glPeriodMonth, _ := strconv.Atoi(splitted[0])
		glPeriodYear := splitted[1]
		t, _ = time.Parse(layout, jv.Dt)
		for _, transaction := range jv.Transactions {
			number := transaction.Account["number"]
			amount := transaction.Amount * 1000
			convertedFloat, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", amount), 64)
			convertedFloat *= 1000
			amountInt := int64(convertedFloat)
			trType := "C"
			if transaction.TransactionType == "debit" {
				trType = "D"
			}
			journalType := "MCCON"
			if exportType == "transfer" {
				journalType = "MCTRS"
			} else if exportType == "invoice" {
				journalType = "MCINV"
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
			sep8 := strings.Repeat(" ", 15)
			line := fmt.Sprintf("%-15s%s0%02d%s%sL%s%018d%s %s%s%-15s%06d-%s%s%s%s%s\n",
				number, glPeriodYear, glPeriodMonth, t.Format("20060102"), sep2,
				sep14, amountInt, trType, journalType, sep5, transactionRef, jv.ID,
				description, sep15, sep8, sep46, defaultCurrency["code"])
			f.WriteString(line)
		}
	}
	return fileName
}

func SerializeReceiving(jvs []JournalVoucher) string {
	t := time.Now()
	fileName := fmt.Sprintf("/tmp/invoice_%s.ndf", t.Format("20060102150405"))
	f, _ := os.Create(fileName)
	defer f.Close()
	for _, jv := range jvs {
		for _, transaction := range jv.Transactions {
			f.WriteString(fmt.Sprintf("%-15s\n", transaction.Account["number"]))
		}
	}
	return fileName
}

func ImportJournalVouchers(w http.ResponseWriter, r *http.Request) {
	path := "/media/share/"
	jvs := FetchJournalVouchers("receiving")
	receivingFileName := Serialize(jvs, "invoice")
	jvs = FetchJournalVouchers("transfer")
	transferFileName := Serialize(jvs, "transfer")
	jvs = FetchJournalVouchers("usage")
	consumptionFileName := Serialize(jvs, "consumption")
	newFile := fmt.Sprintf("%s%s", path, filepath.Base(receivingFileName))
	_ = os.Rename(receivingFileName, newFile)
	newFile = fmt.Sprintf("%s%s", path, filepath.Base(transferFileName))
	_ = os.Rename(transferFileName, newFile)
	newFile = fmt.Sprintf("%s%s", path, filepath.Base(consumptionFileName))
	_ = os.Rename(consumptionFileName, newFile)
}
