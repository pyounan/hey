package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type ConfigHolder struct {
	BackendURI string      `json:"backend_uri"`
	TenantID   string      `json:"tenant_id"`
	FDMs       []FDMConfig `json:"fdms"`
}

type FDMConfig struct {
	FDM_Port  string `json:"fdm_port"`
	FDM_Speed int    `json:"fdm_speed"`
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
	log.Println("Configuration loaded successfully...")
}
