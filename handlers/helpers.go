package handlers

import (
	"math"
	"pos-proxy/fdm"
)

func GroupItemsBySign(items []fdm.POSLineItem) map[string][]fdm.POSLineItem {
	result := make(map[string][]fdm.POSLineItem)
	result["+"] = []fdm.POSLineItem{}
	result["-"] = []fdm.POSLineItem{}
	for _, i := range items {
		price := i.NetAmount + i.TaxAmount
		if price > 0 {
			result["+"] = append(result["+"], i)
		} else {
			i.Price = math.Abs(i.Price)
			i.NetAmount = math.Abs(i.NetAmount)
			i.TaxAmount = math.Abs(i.TaxAmount)
			i.Quantity = math.Abs(i.Quantity)
			result["-"] = append(result["-"], i)
		}
	}

	return result
}

func CalculateVATs(items []fdm.POSLineItem) map[string]float64 {
	VATs := make(map[string]float64)
	VATs["A"] = 0
	VATs["B"] = 0
	VATs["C"] = 0
	VATs["D"] = 0

	for _, i := range items {
		VATs[i.VAT] += math.Abs(i.NetAmount)
	}

	return VATs
}

func CalculateTotalAmount(items []fdm.POSLineItem) float64 {
	total := 0.0

	for _, i := range items {
		if i.Price > 0 {
			total += i.Quantity * i.UnitPrice
		} else {
			total += (i.Quantity * i.UnitPrice) * -1
		}
	}

	total = math.Abs(total)
	return total
}
