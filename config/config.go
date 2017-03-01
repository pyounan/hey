package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
)

type ConfigHolder struct {
	FDM_Port  string `json:"fdm_port"`
	FDM_Speed int    `json:"fdm_speed"`
}

var Config *ConfigHolder = &ConfigHolder{}

func init() {
	flag.String("config", "/etc/cloudinn/pos_config.json", "Configuration for the POS proxy")
	flag.Parse()
	log.Println("Loading configuration...")
	confPath, _ := filepath.Abs("config/config.json")
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
	log.Println("Configuration loaded successfully...")
}
