package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

var ProxyToken string = ""

type ConfigHolder struct {
	BackendURI string      `json:"backend_uri"`
	TenantID   string      `json:"tenant_id"`
	FDMs       []FDMConfig `json:"fdms"`
}

type FDMConfig struct {
	FDM_Port  string   `json:"fdm_port"`
	FDM_Speed int      `json:"fdm_speed"`
	RCRS      []string `json:"rcrs"`
}

var Config *ConfigHolder = &ConfigHolder{}

func Load(file_path string) {
	log.Println("Loading configuration...")
	confPath, _ := filepath.Abs(file_path)
	log.Printf("File: %s\n", confPath)
	f, err := os.Open(confPath)
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&Config)
	if err != nil {
		log.Println("Failed to decode configuration file:")
		log.Fatal(err)
	}
	log.Println("Loading proxy token from /etc/cloudinn/proxy_token.json")
	f, err = os.Open("/etc/cloudinn/proxy_token.json")
	if err != nil {
		log.Fatal(err)
	}
	proxyTokenJson := make(map[string]string)
	decoder = json.NewDecoder(f)
	err = decoder.Decode(&proxyTokenJson)
	if err != nil {
		log.Println("Failed to decode proxy token")
		log.Fatal(err)
	}
	ProxyToken = proxyTokenJson["proxy_token"]
	log.Println("Configuration loaded successfully...")
}
