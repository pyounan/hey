package printing

type Printing interface {
	PrintFolio(req FolioPrint) error
	PrintKitchen(req KitchenPrint) error
}
