package config

import (
	"encoding/json"
	"os"
	"log"
	"path/filepath"
)

type ConfigHolder struct {
	FDM_Port string `json:"fdm_port"`
	FDM_Speed int `json:"fdm_speed"`
}

var Config *ConfigHolder = &ConfigHolder{}

func init() {
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
