package models

// Printer swagger:model printer
// defines attributes of Printer entity
type Printer struct {
	ID          int     `json:"id" bson:"id"`
	PrinterID   string  `json:"printer_id" bson:"printer_id"`
	PrinterType string  `json:"printer_type" bson:"printer_type"`
	PrinterIP   *string `json:"printer_ip" bson:"printer_ip"`
	PaperWidth  int     `json:"paper_width" bson:"paper_width"`
	IsDefault   bool    `json:"is_default" bson:"is_default"`
	TerminalID  int     `json:"terminal" bson:"terminal"`
}

// PrinterSettings swagger:model printerSetting
// defines attributes of PrinterSetting entity
type PrinterSetting struct {
	ID int     `json:"id" bson:"id"`
	IP *string `json:"ip" bson:"ip"`
}
