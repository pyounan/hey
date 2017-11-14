package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AuthUsername reflects the proxy user's username for backend credentials
var AuthUsername string

// AuthPassword reflects the proxy user's password for backend credentials
var AuthPassword string

// Version reflects the current build number
var Version string

// ConfigHolder struct of the proxy configuration
type ConfigHolder struct {
	BackendURI     string      `json:"backend_uri"`
	IsFDMEnabled   bool        `json:"is_fdm_enabled"`
	InstanceName   string      `json:"instance_name"`
	FDMs           []FDMConfig `json:"fdms"`
	IsOperaEnabled bool        `json:"is_opera_enabled"`
	OperaIP        string      `json:"opera_ip"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// FDMConfig struct of each proxy configuration
type FDMConfig struct {
	FDM_Port  string `json:"port"`
	BaudSpeed string `json:"baud_speed"`
	RCRS      string `json:"rcrs"`
}

// Config holds the value of the proxy configuration loaded from
// the configuration file or
var Config *ConfigHolder = &ConfigHolder{}

// Load loads configuration from a file
func Load(filePath string) {
	log.Println("Loading configuration...")
	confPath, _ := filepath.Abs(filePath)
	log.Printf("File: %s\n", confPath)
	f, err := os.Open(confPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&Config)
	if err != nil {
		log.Println("Failed to decode configuration file:")
		log.Fatal(err)
	}
	log.Println("Configuration loaded successfully...")
}

// ParseAuthCredentials reads username and password from auth file
func ParseAuthCredentials(encKey string) error {
	f, err := ioutil.ReadFile("/etc/cloudinn/auth_credentials")
	if err != nil {
		log.Fatal(err)
	}
	// decrypt the retrieved data
	/*f = []byte(strings.TrimSpace(string(f[:])))
	log.Println("data:", string(f[:]))
	txt, err := decrypt(f, []byte(encKey))
	if err != nil {
		return err
	}*/
	splitted := strings.Split(string(f), ",")
	AuthUsername = strings.TrimSpace(splitted[0])
	AuthPassword = strings.TrimSpace(splitted[1])
	return nil
}

// WriteToFile updates the configuration file with new values
func (config *ConfigHolder) WriteToFile() error {
	f, err := os.OpenFile("/etc/cloudinn/pos_config.json", os.O_RDWR, os.ModeExclusive)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	f.Truncate(0)
	f.Seek(0, 0)
	defer f.Close()
	str, err := json.Marshal(config)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	_, err = f.Write(str)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	log.Println("New configurations has been written succesfully to the configuration file")
	return nil
}
