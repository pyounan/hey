package models

type TaxDef struct {
	Name          string `json:"name" bson:"name"`
	VatCode       string `json:"vat_code,omitempty" bson:"vat_code,omitempty"`
	VatPercentage string `json:"vat_percentage,omitempty" bson:"vat_percentage,omitempty"`
	POS           string `json:"pos" bson:"pos"`
	Formula       string `json:"formula" bson:"formula"`
	DepartmentID  int    `json:"department_id" bson:"department_id"`
}

type Department struct {
	Code         int                 `json:"code" bson:"code"`
	ExchangeRate float32             `json:"exchange_rate" bson:"exchange_rate"`
	TaxDefs      map[string][]TaxDef `json:"tax_defs" bson:"tax_defs"`
	ID           int                 `json:"id" bson:"id"`
}
