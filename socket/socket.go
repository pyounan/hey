package socket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var socket = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}
var conn *websocket.Conn

// Event defines the required attributes to send a message through a websocket
type Event struct {
	Module  string          `json:"module"` // e.g: payment, invoice
	Type    string          `json:"type"`   // e.g: output, error
	Payload json.RawMessage `json:"payload"`
}

// StartSocket creates socket and start reading from it forever
func StartSocket(w http.ResponseWriter, r *http.Request) {
	var err error
	conn, err = socket.Upgrade(w, r, nil)
	if err != nil {
		log.Print(err)
		return
	}
	for {
		event, err := read()
		if err != nil {
			log.Println(err)
			continue
		}
		if v, ok := registry[event.Module]; ok {
			v <- event
		}
	}
}

func Send(msg Event) {
	err := conn.WriteJSON(msg)
	if err != nil {
		log.Println(err)
	}
}

func read() (Event, error) {
	message := Event{}
	err := conn.ReadJSON(&message)
	if err != nil {
		log.Println(err)
		return message, err
	}

	return message, nil
}
