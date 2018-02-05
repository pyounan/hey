package income

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pos-proxy/config"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/pos/fdm"
	"pos-proxy/pos/locks"
	"pos-proxy/pos/models"
	"pos-proxy/syncer"
	"strconv"

	"gopkg.in/mgo.v2/bson"
)

// Cashier models data of a Cashier
type Cashier struct {
	ID          int    `json:"id" bson:"id"`
	Name        string `json:"name" bson:"name"`
	Number      int    `json:"number" bson:"number"`
	EmployeeID  string `json:"employee_id" bson:"employee_id"`
	FDMLanguage string `json:"fdm_language,omitempty" bson:"-"` // used only to return in the login response
}

// Attendance represents cashier attendance log
type Attendance struct {
	BID                bson.ObjectId `json:"_id" bson:"_id"`
	ID                 int           `json:"id" bson:"id"` // cloudinn's id
	CashierID          int           `json:"cashier_id" bson:"cashier_id"`
	ClockinTime        string        `json:"clockin_time" bson:"clockin_time"`
	ClockinTerminalID  int           `json:"clockin_terminal_id" bson:"clockin_terminal_id"`
	ClockoutTime       *string       `json:"clockout_time" bson:"clockout_time"`
	ClockoutTerminalID *int          `json:"clockout_terminal_id" bson:"clockout_terminal_id"`
}

type clockinRequest struct {
	Description  string               `json:"description" bson:"description"`
	Pin          string               `json:"pin" bson:"pin"`
	ClockinTime  string               `json:"clockin_time" bson:"clockin_time"`
	Action       string               `json:"action" bson:"action"`
	TerminalID   int                  `json:"terminal" bson:"terminal"`
	FDMResponses []models.FDMResponse `json:"fdm_responses" bson:"fdm_responses"`
}

type clockoutRequest struct {
	Description  string               `json:"description" bson:"description"`
	TerminalID   int                  `json:"terminal_id" bson:"terminal_id"`
	CashierID    int                  `json:"poscashier_id" bson:"poscashier_id"`
	ClockoutTime string               `json:"clockout_time" bson:"clockout_time"`
	Action       string               `json:"action" bson:"clockout"`
	FDMResponses []models.FDMResponse `json:"fdm_responses" bson:"fdm_responses"`
}

func clockin(cashier Cashier, terminal models.Terminal, time string) (string, models.FDMResponse, error) {
	description := "Clock In"
	fdmResponse := models.FDMResponse{}
	q := bson.M{"cashier_id": cashier.ID}
	attendance := Attendance{}
	session := db.Session.Copy()
	defer session.Close()
	if err := db.DB.C("attendance").With(session).Find(q).Limit(1).Sort("-_id").One(&attendance); err != nil {
		// just log the error, in most cases it means no record was found, so
		// we will create a new one
		log.Println("WARNING", err)
	}
	if attendance.CashierID == 0 || attendance.ClockoutTime != nil {
		a := Attendance{}
		a.BID = bson.NewObjectId()
		a.CashierID = cashier.ID
		a.ClockinTerminalID = terminal.ID
		a.ClockinTime = time
		if config.Config.IsFDMEnabled {
			// create fdm connection
			conn, err := fdm.Connect(terminal.RCRS)
			if err != nil {
				return description, fdmResponse, err
			}
			defer conn.Close()
			fdmConfig := config.FDMConfig{}
			for _, f := range config.Config.FDMs {
				if f.RCRS == terminal.RCRS {
					fdmConfig = f
					break
				}
			}

			fdmReq := models.InvoicePOSTRequest{}
			fdmReq.ActionTime = time
			fdmReq.RCRS = terminal.RCRS
			fdmReq.TerminalID = terminal.ID
			fdmReq.TerminalNumber = terminal.Number
			fdmReq.TerminalName = terminal.Description
			fdmReq.EmployeeID = cashier.EmployeeID
			fdmReq.CashierName = cashier.Name
			fdmReq.CashierNumber = cashier.Number
			item := models.EJEvent{}
			if fdmConfig.Language == "fr" {
				description = "TRAVAIL IN"
			} else {
				description = "ARBEID IN"
			}
			item.Description = description
			item.VATCode = "D"
			item.Quantity = 1
			fdmReq.Invoice.Items = append(fdmReq.Invoice.Items, models.POSLineItem{})
			fdmResponse, err = fdm.SendHashAndSignMessage(conn, "NS", fdmReq, []models.EJEvent{item})
			if err != nil {
				return description, fdmResponse, err
			}
		}
		err := db.DB.C("attendance").With(session).Insert(a)
		if err != nil {
			return description, fdmResponse, err
		}
		return description, fdmResponse, nil
	}

	return description, fdmResponse, nil
}

func clockout(cashier Cashier, terminal models.Terminal, time string) (string, models.FDMResponse, error) {
	description := "Clock Out"
	fdmResponse := models.FDMResponse{}
	if config.Config.IsFDMEnabled {
		// create fdm connection
		conn, err := fdm.Connect(terminal.RCRS)
		if err != nil {
			return description, fdmResponse, err
		}
		defer conn.Close()
		fdmConfig := config.FDMConfig{}
		for _, f := range config.Config.FDMs {
			if f.RCRS == terminal.RCRS {
				fdmConfig = f
				break
			}
		}

		fdmReq := models.InvoicePOSTRequest{}
		fdmReq.ActionTime = time
		fdmReq.RCRS = terminal.RCRS
		fdmReq.TerminalID = terminal.ID
		fdmReq.TerminalNumber = terminal.Number
		fdmReq.TerminalName = terminal.Description
		fdmReq.EmployeeID = cashier.EmployeeID
		fdmReq.CashierName = cashier.Name
		fdmReq.CashierNumber = cashier.Number
		item := models.EJEvent{}
		if fdmConfig.Language == "fr" {
			description = "TRAVAIL OUT"
		} else {
			description = "ARBEID UIT"
		}
		item.Description = description
		item.Quantity = 1
		item.VATCode = "D"
		fdmResponse, err = fdm.SendHashAndSignMessage(conn, "NS", fdmReq, []models.EJEvent{item})
		if err != nil {
			return description, fdmResponse, err
		}
	}
	q := bson.M{"cashier_id": cashier.ID, "clockout_time": nil}
	updateQ := bson.M{"$set": bson.M{"clockout_time": time, "clockout_terminal_id": terminal.ID}}
	session := db.Session.Copy()
	defer session.Close()
	_, err := db.DB.C("attendance").With(session).UpdateAll(q, updateQ)
	if err != nil {
		return description, fdmResponse, nil
	}
	return description, fdmResponse, nil
}

// GetPosCashier used to clockin a cashier and compare his
// pin against the correct one from the database
func GetPosCashier(w http.ResponseWriter, req *http.Request) {
	cashier := Cashier{}
	q := bson.M{}
	store, _ := strconv.Atoi(req.URL.Query().Get("store"))

	postBody := clockinRequest{}
	err := json.NewDecoder(req.Body).Decode(&postBody)
	if err != nil {
		helpers.ReturnSuccessMessage(w, err.Error())
		return
	}
	defer req.Body.Close()

	q["pin"] = postBody.Pin
	q["store_set"] = store
	session := db.Session.Copy()
	defer session.Close()
	err = db.DB.C("cashiers").With(session).Find(q).One(&cashier)
	if err != nil {
		helpers.ReturnErrorMessageWithStatus(w, 400, "No matching PIN code to selected store.")
		return
	}
	if config.Config.IsFDMEnabled && cashier.EmployeeID == "" {
		helpers.ReturnErrorMessageWithStatus(w, 400, "employee id is not set")
		return
	}
	if config.Config.IsFDMEnabled && len(cashier.EmployeeID) < 11 {
		helpers.ReturnErrorMessageWithStatus(w, 400, "employee id is not valid, must be 11 characters")
		return
	}

	cashierHashExists := true
	_, err = req.Cookie("cashier_hash")
	if err != nil {
		cashierHashExists = false
	}
	otherCashier, err := locks.LockTerminal(postBody.TerminalID, cashier.ID)
	if err != nil && ((cashierHashExists && otherCashier == cashier.ID) || otherCashier != cashier.ID) {
		helpers.ReturnErrorMessageWithStatus(w, 400, "Terminal is locked.")
		return
	}

	resp := cashier
	terminal := models.Terminal{}
	err = db.DB.C("terminals").With(session).Find(bson.M{"id": postBody.TerminalID}).One(&terminal)
	if err != nil {
		helpers.ReturnErrorMessageWithStatus(w, 500, err.Error())
		return
	}
	description, fdmResponse, err := clockin(cashier, terminal, postBody.ClockinTime)
	if err != nil {
		helpers.ReturnErrorMessageWithStatus(w, 500, err.Error())
		return
	}
	if config.Config.IsFDMEnabled {
		postBody.FDMResponses = []models.FDMResponse{fdmResponse}
		postBody.Description = description
	}
	syncer.QueueRequest(req.RequestURI, req.Method, req.Header, postBody)

	if config.Config.IsFDMEnabled {
		for _, fdm := range config.Config.FDMs {
			if fdm.RCRS == terminal.RCRS {
				resp.FDMLanguage = fdm.Language
				break
			}
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "cashier_id",
		Value: fmt.Sprintf("%d", cashier.ID),
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "cashier_hash",
		Value: "1",
	})

	helpers.ReturnSuccessMessage(w, resp)
}

// GetCashierPermissions return a json http response contains list of
// permissions assigned to a certain cashier
func GetCashierPermissions(w http.ResponseWriter, r *http.Request) {
	query := bson.M{}
	queryParams := r.URL.Query()
	for key, val := range queryParams {
		query[key] = val
	}
	permissions := []map[string]interface{}{}
	session := db.Session.Copy()
	defer session.Close()
	err := db.DB.C("permissions").With(session).Find(query).All(&permissions)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	helpers.ReturnSuccessMessage(w, permissions)
}

// Clockout logs out a cashier
func Clockout(w http.ResponseWriter, r *http.Request) {
	body := clockoutRequest{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	defer r.Body.Close()
	cashier := Cashier{}
	q := bson.M{"id": body.CashierID}
	session := db.Session.Copy()
	defer session.Close()
	err = db.DB.C("cashiers").With(session).Find(q).One(&cashier)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	terminal := models.Terminal{}
	q = bson.M{"id": body.TerminalID}
	err = db.DB.C("terminals").With(session).Find(q).One(&terminal)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	description, fdmResp, err := clockout(cashier, terminal, body.ClockoutTime)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}
	locks.UnlockTerminal(terminal.ID)
	if config.Config.IsFDMEnabled {
		body.Description = description
		body.FDMResponses = []models.FDMResponse{fdmResp}
	}
	syncer.QueueRequest(r.RequestURI, r.Method, r.Header, body)
	helpers.ReturnSuccessMessage(w, true)
}
