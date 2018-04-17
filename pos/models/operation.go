package models

// FDMOperation swagger:model fdmOperation
// defines entity that holds the body of request required to communicate with FDM
type FDMOperation struct {
	RCRS          string
	Invoice       map[string]interface{}
	Events        []map[string]interface{}
	LineItems     []POSLineItem
	TerminalName  string
	CashierName   string
	CashierNumber string
}
