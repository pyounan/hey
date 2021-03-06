package models

// TaxDef swagger:model taxDef
// Defines attributes of a Tax Definition
type TaxDef struct {
	Name          string  `json:"name" bson:"name"`
	VatCode       string  `json:"vat_code,omitempty" bson:"vat_code,omitempty"`
	VatPercentage float64 `json:"vat_percentage" bson:"vat_percentage"`
	POS           string  `json:"pos" bson:"pos"`
	Formula       string  `json:"formula" bson:"formula"`
	DepartmentID  int     `json:"department_id" bson:"department_id"`
}

// Department swagger:model department
// defines attributes of a department model
type Department struct {
	ID              int                 `json:"id" bson:"id"`
	Code            int                 `json:"code" bson:"code"`
	Name            string              `json:"name" bson:"name"`
	Type            string              `json:"type" bson:"type"`
	ExchangeRate    string              `json:"exchange_rate" bson:"exchange_rate"`
	TaxDefs         map[string][]TaxDef `json:"tax_defs" bson:"tax_defs"`
	CurrencyID      int                 `json:"currency" bson:"currency"`
	CurrencyDetails string              `json:"currency_details" bson:"currency_details"`
	PaymentGateway  *string             `json:"payment_gateway" bson:"payment_gateway"`
	PaymentType     *string             `json:"payment_type" bson:"payment_type"`
	POSPayment      bool                `json:"pos_payment" bson:"pos_payment"`
	OpenCashDrawer  bool                `json:"open_cash_drawer" bson:"open_cash_drawer"`
}
