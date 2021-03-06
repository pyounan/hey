package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pos-proxy/config"
	"pos-proxy/helpers"
	"sync"
	"syscall"
	"time"

	"github.com/TV4/graceful"
)

var netClient = helpers.NewNetClient()
var c = make(chan os.Signal, 1)

// KeypairReloader holds info required to use the certificate certificate. Thread safe
// SO https://stackoverflow.com/questions/37473201/is-there-a-way-to-update-the-tls-certificates-in-a-net-http-server-without-any-d
type KeypairReloader struct {
	certMu   sync.Mutex
	cert     *tls.Certificate
	certPath string
	keyPath  string
}

// NewKeyPairReloader return an intance of the reloader to load the new certificate on the fly
func NewKeyPairReloader(certPath, keyPath string) (*KeypairReloader, error) {
	result := &KeypairReloader{
		certPath: certPath,
		keyPath:  keyPath,
	}
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	go func() {
		signal.Notify(c, syscall.SIGHUP)
		for range c {
			log.Printf("Received SIGHUP, reloading TLS certificate")
			fetchCertificate()
			if err := result.maybeReload(); err != nil {
				log.Printf("Keeping old TLS certificate because the new one could not be loaded: %v", err)
			}
		}
	}()
	result.cert = &cert
	return result, nil
}

func (kpr *KeypairReloader) maybeReload() error {
	newCert, err := tls.LoadX509KeyPair(kpr.certPath, kpr.keyPath)
	if err != nil {
		return err
	}
	kpr.certMu.Lock()
	defer kpr.certMu.Unlock()
	kpr.cert = &newCert
	return nil
}

// GetCertificateFunc returns a new certificate to the server's TlsConfig
func (kpr *KeypairReloader) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		log.Println("mutex", kpr.certMu)
		kpr.certMu.Lock()
		defer kpr.certMu.Unlock()
		return kpr.cert, nil
	}
}

func startTLS(handler http.Handler) {
	go fetchCertificate()
	crtFile := "/etc/cloudinn/tls.crt"
	keyFile := "/etc/cloudinn/tls.key"
	var kpr *KeypairReloader
	for {
		sleepTime := 1 * time.Minute
		_, err := os.Stat(crtFile)
		if os.IsNotExist(err) {

			log.Println("crt file not found .. sleeping")
			time.Sleep(sleepTime)
			continue
		}
		_, err = os.Stat(keyFile)
		if os.IsNotExist(err) {
			log.Println("key file not found .. sleeping")
			time.Sleep(sleepTime)
			continue
		}

		log.Println("Found files, now creating kpr")
		kpr, err = NewKeyPairReloader(crtFile, keyFile)
		if err != nil {
			log.Println(err.Error())
			time.Sleep(sleepTime)
			continue
		}
		log.Println("kpr cretated, breaking", kpr)
		break
	}

	log.Println("handler", handler)
	srv := &http.Server{
		Addr:    ":443",
		Handler: handler,
	}
	log.Println("tlsconfig", srv.TLSConfig)
	srv.TLSConfig = &tls.Config{
		NextProtos: []string{"http/1.1", "http/2"},
	}
	srv.TLSConfig.GetCertificate = kpr.GetCertificateFunc()
	graceful.ListenAndServeTLS(srv, "", "")
}

func fetchCertificate() {
	type RequestBody struct {
		ClientID int    `json:"client_id"`
		EnvName  string `json:"env_name"`
	}
	type ResponseBody struct {
		TLSCrt string `json:"tls.crt"`
		TLSKey string `json:"tls.key"`
	}
	for {
		rBody := RequestBody{EnvName: *config.Config.VirtualHost, ClientID: int(config.Config.InstanceID)}
		uri := fmt.Sprintf("%s%s", config.Config.BackendURI, "/api/pos/proxy/cert/")
		requestBody, err := json.Marshal(rBody)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		req, err := http.NewRequest("POST", uri, bytes.NewBuffer(requestBody))
		req = helpers.PrepareRequestHeaders(req)
		resp, err := netClient.Do(req)
		if err != nil {
			log.Println("Failed to get update data", err.Error())
			time.Sleep(5 * time.Minute)
			continue
		}
		defer resp.Body.Close()
		var respBody ResponseBody
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		if err != nil {
			log.Println("Failed to read response body", err.Error())
			time.Sleep(5 * time.Minute)
			continue
		}
		crt := []byte(respBody.TLSCrt)
		err = ioutil.WriteFile("/etc/cloudinn/tls.crt", crt, 0644)
		if err != nil {
			log.Println("Failed to write certificate secret to file", err.Error())
			time.Sleep(5 * time.Minute)
			continue
		}
		key := []byte(respBody.TLSKey)
		err = ioutil.WriteFile("/etc/cloudinn/tls.key", key, 0644)
		if err != nil {
			log.Println("Failed to write certificate key to file", err.Error())
			time.Sleep(5 * time.Minute)
			continue
		}
		c <- syscall.SIGHUP
		time.Sleep(24 * time.Hour)
	}
}
