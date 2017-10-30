package sun

type JournalVoucher struct {
	ID           int64         `json:"id"`
	GLPeriod     string        `json:"gl_period"`
	Dt           string        `json:"posted_on"`
	Description  string        `json:"description"`
	Transactions []Transaction `json:"transaction_set"`
}

type Transaction struct {
	Account         map[string]interface{} `json:"account"`
	TransactionType string                 `json:"transaction_type"`
	Amount          float64                `json:"amount"`
	Description     string                 `json:"description"`
}
