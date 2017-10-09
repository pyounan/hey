package proxy

import (
	"html/template"
	"log"
	"net/http"
	"pos-proxy/db"
	"pos-proxy/helpers"
	"pos-proxy/syncer"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

func HomeView(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/header.html", "templates/home.html", "templates/footer.html")
	if err != nil {
		log.Panic(err)
	}

	t.ExecuteTemplate(w, "home", nil)
}

func RequestsLogView(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/header.html", "templates/syncer_logs.html",
		"templates/footer.html")
	if err != nil {
		log.Panic(err)
	}

	logs := []syncer.RequestLog{}
	db.DB.C("requests_log").Find(nil).All(&logs)
	for _, i := range logs {
		log.Println(i)
	}

	t.ExecuteTemplate(w, "syncer_logs", logs)
}

func SyncerRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	logRecord := syncer.RequestLog{}
	err := db.DB.C("requests_log").FindId(bson.ObjectIdHex(id)).One(&logRecord)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	helpers.ReturnSuccessMessage(w, logRecord.Request)
}

func SyncerResponse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	logRecord := syncer.RequestLog{}
	err := db.DB.C("requests_log").FindId(bson.ObjectIdHex(id)).One(&logRecord)
	if err != nil {
		helpers.ReturnErrorMessage(w, err.Error())
		return
	}

	helpers.ReturnSuccessMessage(w, logRecord.ResponseBody)
}
