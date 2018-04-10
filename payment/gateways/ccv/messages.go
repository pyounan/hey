package ccv

import "encoding/json"

type Command interface {
	GetType() string
	GetPayload() []byte
}

type Message struct {
	Type string
}

func (msg Message) GetType() string {
	return msg.Type
}

type Output struct {
	Message
	Text []string
}

func (msg Output) GetType() string {
	return msg.Type
}

func (msg Output) GetPayload() []byte {
	b, err := json.Marshal(msg.Text)
	if err != nil {
		return b
	}
	return b
}
