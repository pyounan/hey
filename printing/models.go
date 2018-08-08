package printing

import (
	"pos-proxy/income"
	"pos-proxy/pos/models"
)

//FolioPrint defines the objects and requried data for folio printing
type FolioPrint struct {
	Items          []models.EJEvent `bson:"items"`
	Invoice        models.Invoice   `bson:"invoice"`
	Terminal       models.Terminal  `bson:"termianl"`
	Store          models.Store     `bson:"store"`
	Cashier        income.Cashier   `bson:"cashier"`
	Company        income.Company   `bson:"company"`
	Printer        models.Printer   `bson:"printer"`
	TotalDiscounts float64          `bson:"total_discounts"`
	Timezone       string           `bson:"timezone"`
}

//KitchenPrint defines the objects and requried data for kitchen printing
type KitchenPrint struct {
	GropLineItems []models.EJEvent `bson:"group_lineitems`
	Invoice       models.Invoice   `bson:"invoice"`
	Printer       models.Printer   `bson:"printer"`
	Cashier       income.Cashier   `bson:"cashier"`
	Timezone      string           `bson:"timezone"`
}
