package models

// Printer swagger:model printer
// defines attributes of Printer entity
type Printer struct {
	ID        int    `json:"id" bson:"id"`
	PrinterID string `json:"printer_id" bson:"printer_id"`
	//type : cashier or kitchen
	PrinterType string  `json:"printer_type" bson:"printer_type"`
	PrinterIP   *string `json:"printer_ip" bson:"printer_ip"`
	PaperWidth  int     `json:"paper_width" bson:"paper_width"`
	IsDefault   bool    `json:"is_default" bson:"is_default"`
	TerminalID  int     `json:"terminal" bson:"terminal"`
	IsUSB       bool    `json:"is_usb" bson:"is_usb"`
}

// PrinterSettings swagger:model printerSetting
// defines attributes of PrinterSetting entity
type PrinterSetting struct {
	ID int     `json:"id" bson:"id"`
	IP *string `json:"ip" bson:"ip"`
}

//StoreMenuItemConfig
type StoreMenuItemConfig struct {
	ID                 int `json:"id" bson:"id"`
	AttachedAttributes struct {
		KitchenPrinter    int `json:"kitchen_printer" bson:"kitchen_printer"`
		RevenueDepartment int `json:"revenue_department" bson:"revenue_department"`
	} `json:"attached_attributes" bson:"attached_attributes"`
	Store int `json:"store" bson:"store"`
	Menu  int `json:"menu" bson:"menu"`
	Item  int `json:"item" bson:"item"`
}
