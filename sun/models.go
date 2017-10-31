package sun

type JournalVoucher struct {
	ID            int64         `json:"id"`
	GLPeriod      string        `json:"gl_period"`
	Dt            string        `json:"dt"`
	Description   string        `json:"description"`
	Transactions  []Transaction `json:"transaction_set"`
	OperationType string        `json:"operation_type"`
	OperationId   int64         `json:"operation_id"`
}

type Transaction struct {
	Account         map[string]interface{} `json:"account"`
	TransactionType string                 `json:"transaction_type"`
	Amount          float64                `json:"amount"`
	Description     string                 `json:"description"`
}
