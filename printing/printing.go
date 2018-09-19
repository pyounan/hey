package printing

//Printing interface that implemnts PrintFolio and PrintKitchen
type Printing interface {
	PrintFolio(req FolioPrint) error
	PrintKitchen(req KitchenPrint) error
}
