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

type SunMapping struct {
	Code    string `json:"account_code"`
	TCode1  string `json:"analysis_code_1"`
	TCode2  string `json:"analysis_code_2"`
	TCode3  string `json:"analysis_code_3"`
	TCode4  string `json:"analysis_code_4"`
	TCode5  string `json:"analysis_code_5"`
	TCode6  string `json:"analysis_code_6"`
	TCode7  string `json:"analysis_code_7"`
	TCode8  string `json:"analysis_code_8"`
	TCode9  string `json:"analysis_code_9"`
	TCode10 string `json:"analysis_code_10`
}

type Transaction struct {
	Account         map[string]interface{} `json:"account"`
	TransactionType string                 `json:"transaction_type"`
	Amount          float64                `json:"amount"`
	Description     string                 `json:"description"`
	SunMapping      SunMapping             `json:"sunmapping"`
}
