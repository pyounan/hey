package syncer

import (
	"net/http"
	"pos-proxy/db"
)

type RequestRow struct {
	URI     string      `bson:"uri"`
	Method  string      `bson:"method"`
	Headers http.Header `bson:"headers"`
	Payload interface{} `bson:"payload"`
}

// QueueRequest insert a request object to a queue that syncs with the backend
func QueueRequest(uri string, method string, headers http.Header, payload interface{}) error {
	body := &RequestRow{}
	body.URI = uri
	body.Method = method
	body.Headers = headers
	body.Payload = payload

	err := db.DB.C("requests_queue").Insert(body)
	if err != nil {
		return err
	}
	return nil
}

func PushToBackend() error {
	return nil
}
