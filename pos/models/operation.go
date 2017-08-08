package models

type FDMOperation struct {
	RCRS          string
	Invoice       map[string]interface{}
	Events        []map[string]interface{}
	LineItems     []POSLineItem
	TerminalName  string
	CashierName   string
	CashierNumber string
}
