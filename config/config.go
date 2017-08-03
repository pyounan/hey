package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ProxyToken holds the value of the token used to authorize the
// proxy requests to CloudInn servers
var ProxyToken string = ""

// ConfigHolder struct of the proxy configuration
type ConfigHolder struct {
	BackendURI   string      `json:"backend_uri"`
	IsFDMEnabled bool        `json:"is_fdm_enabled"`
	FDMs         []FDMConfig `json:"fdms"`
	UpdatedAt    time.Time   `json:"updated_at"`
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
func Load(file_path string) {
	log.Println("Loading configuration...")
	confPath, _ := filepath.Abs(file_path)
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

// FetchConfiguration asks CloudInn servers if the conf were updated,
// if yes update the current configurations and write them to the conf file
func FetchConfiguration() {
	uri := fmt.Sprintf("%s/api/pos/proxy/settings/", Config.BackendURI)
	netClient := &http.Client{
		Timeout: time.Second * 5,
	}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("JWT %s", ProxyToken))
	response, err := netClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println(response)
	// open configurations file
	f, err := os.Open("/etc/cloudinn/pos_config.json")
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer f.Close()
	data, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Printf("%s\n", data)
	type ProxySettings struct {
		UpdatedAt string      `json:"updated_at"`
		FDMs      []FDMConfig `json:"fdms"`
	}
	dataStr := ProxySettings{}
	err = json.Unmarshal(data, &dataStr)
	if err != nil {
		log.Println(err.Error())
		return
	}
	t, err := time.Parse(time.RFC3339, fmt.Sprintf("%s", dataStr.UpdatedAt))
	if err != nil {
		log.Println(err.Error())
		return
	}
	// Check the configurations coming from the backend are newer than
	// the current configuration
	if (Config.UpdatedAt != time.Time{}) && !t.After(Config.UpdatedAt) {
		return
	}
	log.Println("New configurations found")
	Config.FDMs = dataStr.FDMs
	Config.UpdatedAt = t
	if err := Config.WriteToFile(); err != nil {
		log.Println(err.Error())
		return
	}
}

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
