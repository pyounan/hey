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
	requests := []RequestRow{}
	db.DB.C("requests_queue").Find(nil).All(&requests)

	for _, r := range requests {
		uri := fmt.Sprintf("%s/%s", Config.BackendURI, r.URI)
		netClient := &http.Client{
			Timeout: time.Second * 5,
		}
		req, err := http.NewRequest(r.Method, uri, r.Payload)
		if err != nil {
			log.Println(err.Error())
			return
		}
		req.Header = r.Headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("JWT %s", config.ProxyToken))
		response, err := netClient.Do(req)
		if err != nil {
			log.Println(err.Error())
			return
		}
		defer response.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(response.Body).Decode(&res)
		for _, item := range res {
			err = db.DB.C(collection).Upsert(item)
			if err != nil {
				log.Println(err.Error())
			}
		}
	}
}
