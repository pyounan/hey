package opera_test

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"pos-proxy/config"
	"pos-proxy/opera"
	"testing"
)

var (
	server   *httptest.Server
	usersUrl string
)

func init() {
	filePath := flag.String("config", "/etc/cloudinn/pos_config.json", "Configuration for the POS proxy")
	flag.Parse()
	config.Load(*filePath)
	r := mux.NewRouter()
	r.HandleFunc("/api/opera/rooms/", opera.ListOperaRooms).Methods("GET")
	opera.Connect()
	server = httptest.NewServer(r) //Creating new server with the user handlers

	usersUrl = fmt.Sprintf("%s/api/opera/rooms/", server.URL) //Grab the address for the API endpoint
}

func TestListOperaRooms(t *testing.T) {
	userJson := `?store=1&terminal=1&room_number=101`

	//reader = strings.NewReader(userJson) //Convert string to reader

	request, err := http.NewRequest("GET", usersUrl+userJson, nil) //Create request with JSON body

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err) //Something is wrong while sending request
	}

	if res.StatusCode != 201 {
		t.Errorf("Success expected: %d", res.StatusCode) //Uh-oh this means our test failed
	}
}
