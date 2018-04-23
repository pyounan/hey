package entity

// CCVSettings swagger:model ccvSettings
// defines attributes of CCV pinpads setup
type CCVSettings struct {
	ID          int    `json:"id" bson:"id"`
	Description string `json:"description" bson:"description"`
	IP          string `json:"ip" bson:"ip"`
	PinpadPort  int    `json:"pinpad_port" bson:"pinpad_port"`
	ProxyPort   int    `json:"proxy_port" bson:"proxy_port"`
}

// CCVTerminalIntegration swagger:model ccvTerminalIntegration
// defines attributes that attaches a CCV pinpad to a POS terminal
type CCVTerminalIntegration struct {
	ID            int `json:"id" bson:"id"`
	TerminalID    int `json:"terminal_id" bson:"terminal_id"`
	CCVSettingsID int `json:"ccv_settings_id" bson:"ccv_settings_id"`
}
