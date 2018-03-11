package callaccounting

import (
	"fmt"
	"log"
	"plugin"
	"pos-proxy/config"
)

// CallAccounting interface should be implemented by the plugin that is going to be laoded
type CallAccounting interface {
	UpdateSettings(map[string]string)
	Start()
}

var callAccounting CallAccounting

// LoadPlugin checks if the callaccounting flag is enabled then loads the plugin
// into memory.
// NOTICE: it should be the first thing to call before any other methods
func LoadPlugin() {
	p, err := plugin.Open("plugins/call_accounting.so")
	if err != nil {
		log.Println(err)
		return
	}

	s, err := p.Lookup("CallAccounting")
	if err != nil {
		log.Println(err)
		return
	}

	var ok bool
	callAccounting, ok = s.(CallAccounting)
	if !ok {
		log.Println("Plugin CallAccounting doesn't implement CallAccouting interface")
		return
	}

}

// Start initiates the loading process of call_accounting plugin
func Start() {
	callAccounting.Start()
}

// UpdateSettings updates the Configuration for the plugin
func UpdateSettings(cfg Config) {
	configMap := make(map[string]string)
	configMap["tenant_id"] = fmt.Sprintf("%d", config.Config.InstanceID)
	configMap["realm"] = config.Config.InstanceName
	configMap["username"] = cfg.Username
	configMap["password"] = cfg.Password
	configMap["ip"] = cfg.IP
	callAccounting.UpdateSettings(configMap)
}
