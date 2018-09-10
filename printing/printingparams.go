package printing

//PrintingParams creates a mapof paper width and required param and returns it's value
func PrintingParams(paperWidth int, param string) int {

	printingParams := make(map[int]map[string]int)

	printingParams[80] = make(map[string]int)

	printingParams[80]["width"] = 800
	printingParams[80]["company_name_width"] = 2
	printingParams[80]["company_name_height"] = 2
	printingParams[80]["item_padding"] = 24
	printingParams[80]["qty_padding"] = 5
	printingParams[80]["price_padding"] = 8
	printingParams[80]["subtotal_padding"] = 40
	printingParams[80]["total_padding"] = 21
	printingParams[80]["fdm_rate_padding"] = 10
	printingParams[80]["fdm_taxable_padding"] = 10
	printingParams[80]["fdm_vat_padding"] = 10
	printingParams[80]["fdm_net_padding"] = 10
	printingParams[80]["tax_padding"] = 5
	printingParams[80]["char_per_line"] = 40
	printingParams[80]["store_unit"] = 2
	printingParams[80]["item_kitchen"] = 27
	printingParams[80]["qty_kitchen"] = 7
	printingParams[80]["unit"] = 5

	printingParams[76] = make(map[string]int)

	printingParams[76]["width"] = 760
	printingParams[76]["company_name_width"] = 2
	printingParams[76]["company_name_height"] = 2
	printingParams[76]["item_padding"] = 24
	printingParams[76]["qty_padding"] = 5
	printingParams[76]["price_padding"] = 8
	printingParams[76]["subtotal_padding"] = 32
	printingParams[76]["total_padding"] = 19
	printingParams[76]["fdm_rate_padding"] = 8
	printingParams[76]["fdm_taxable_padding"] = 8
	printingParams[76]["fdm_vat_padding"] = 8
	printingParams[76]["fdm_net_padding"] = 8
	printingParams[76]["tax_padding"] = 5
	printingParams[76]["char_per_line"] = 32
	printingParams[76]["store_unit"] = 2
	printingParams[76]["item_kitchen"] = 27
	printingParams[76]["qty_kitchen"] = 7
	printingParams[76]["unit"] = 5

	return printingParams[paperWidth][param]

}
