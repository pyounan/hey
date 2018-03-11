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

// Start initiates the loading process of call_accounting plugin
func Start() {
	if !config.Config.CallAccountingEnabled {
		return
	}
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

	callAccounting.Start()
}

// SetSettings updates the Configuration for the plugin
func SetSettings(cfg Config) {
	configMap := make(map[string]string)
	configMap["tenant_id"] = fmt.Sprintf("%d", config.Config.InstanceID)
	configMap["realm"] = config.Config.InstanceName
	configMap["username"] = cfg.Username
	configMap["password"] = cfg.Password
	configMap["ip"] = cfg.IP
	callAccounting.UpdateSettings(configMap)
}
