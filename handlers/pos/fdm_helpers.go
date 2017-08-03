package pos

import (
	"math"
	"log"
	"pos-proxy/config"
	"pos-proxy/handlers"
	"pos-proxy/integrations/fdm"
)

func submitToFDM(req InvoicePOSTRequest) (fdm.Response, error) {
	// if fdm is enabled submit items to fdm
	if config.Config.IsFDMEnabled == true {
		f, err := fdm.New(req.RCRS)
		if err != nil {
			return fdm.Response{}, err
		}
		// Check FDM Status before submit
		fdmResponse, err := f.CheckStatus()
		if err != nil {
			return fdmResponse, err
		}
		err = fdm.CheckError(fdmResponse)
		if err != nil {
			handlers.ReturnJSONError(w, err.Error())
			return
		}
		// submit items
		items := separateCondimentsAndDiscounts(req.LineItems)
		vats := calculateVATs(items)
		positiveVATs := []string{}
		negativeVATs := []string{}
		for rate, amount := range vats {
			if amount >= 0 {
				positiveVATs = append(positiveVATs, rate)
			} else if amount < 0 {
				negativeVATs = append(negativeVATs, rate)
			}
		}

		// send positive msg
		splittedItems := splitItemsByVATRates(items, positiveVATs)
		if len(items) > 0 {
			res, err := f.SendMessage("PS", req, splittedItems)
			if err != nil {
				return fdmResponse, err
			}
		}
		// send negative msg
		splittedItems = splitItemsByVATRates(items, negativeVATs)
		if len(items) > 0 {
			res, err := f.SendMessage("PR", req, splittedItems)
			if err != nil {
				return fdmResponse, err
			}
		}

		return fdmResponse, nil

	} else {
		return fdm.Response{}, nil
	}
}

func calculateVATs(items []fdm.POSLineItem) map[string]float64 {
	VATs := make(map[string]float64)
	VATs["A"] = 0
	VATs["B"] = 0
	VATs["C"] = 0
	VATs["D"] = 0

	for _, i := range items {
		VATs[i.VAT] += i.NetAmount
	}

	return VATs
}

func separateCondimentsAndDiscounts(rawItems []map["string"]interface{}) []fdm.POSLineItem {
	items := []fdm.POSLineItem{}
	for _, ri := range rawItems {
		priceOperator := 1
		item := fdm.POSLineItem{}
		item.IsCondiment = false
		item.IsDiscount = false
		item.Description = ri["description"]
		item.ID = ri["id"]
		item.LineItemType = "sales"
		item.Quantity = ri["qty"]
		item.Price = ri["price"]
		item.NetAmount = ri["net_amount"]
		item.TaxAmount = ri["tax_amount"]
		item.UnitPrice = ri["unit_price"]
		item.VAT = ri["vat_code"]
		item.VATPercentage = ri["vat_percentage"]
		if item.Price < 0 {
			item.LineItemType = "return"
			if item.NetAmount > 0 {
				item.NetAmount = -1 * item.NetAmount
			}
		}
		items = append(items, item)
		// loop on condiments and add them
		for _, c := range ri["condimentlineitem_set"] {
			cond := fdm.POSLineItem{}
			cond.IsCondiment = true
			cond.IsDiscount = false
			cond.Description = c["name"]
			cond.LineItemType = "sales"			
			cond.Quantity = c["qty"]
			cond.Price = c["price"]
			cond.UnitPrice = c["unit_price"]
			cond.NetAmount = c["net_amount"]
			cond.TaxAmount = c["tax_amount"]
			cond.VAT = c["vat_code"]
			cond.VATPercentage = c["vat_percentage"]
			if item.Price < 0 {
				cond.LineItemType = "return"
					cond.Price = -1 * math.Abs(cond.Price)
					cond.NetAmount = -1 * math.Abs(cond.NetAmount)
			}
			items = append(items, cond)
		}

		// loop on discounts and add them
		for _, discount := range ri["grouped_applieddiscounts"] {
			for key, val range discount {
				disc := fdm.POSLineItem{}
				disc.IsCondiment = false
				disc.IsDiscount = true
				disc.Description = fmt.Sprintf("%s Discount %s%", key, val["percentage"])
				disc.Quantity = priceOperator
				disc.LineItemType = "return"				
				disc.Price = -1 * math.Abs(val["amount"])
				disc.UnitPrice = val["price"]
				disc.NetAmount = -1 * val["net_amount"]
				disc.TaxAmount = val["tax_amount"]
				disc.VAT = val["vat_code"]
				disc.VATPercentage = val["vat_percentage"]
				if item.Price < 0 {
					disc.LineItemType = "sales"
					disc.Price = math.Abs(disc.Price)
					disc.NetAmount = math.Abs(disc.Price)
				}
			}
		}
	}
}

func splitItemsByVATRates(items []fdm.POSLineItem, rates []string) []fdm.POSLineItem {

	result := []fdm.POSLineItem{}
	for _, item := range items {
		for _, rate := range rates {
			if item.VAT == rate {
				result = append(result, item)
			}
		}
	}
	return result
}

func fixItemsPrice(items []fdm.POSLineItem) []fdm.POSLineItem {
	for _, item := range items {
		operator := 1.0
		if item.Price < 0 {
			operator = -1.0
		}
		item.Price = operator * item.Quantity * item.UnitPrice
		log.Printf("item price %f", item.Price)
	}
	return items
}

func calculateTotalAmount(items []fdm.POSLineItem) float64 {
	total := 0.0

	for _, i := range items {
		total += i.Price
	}

	return total
}
