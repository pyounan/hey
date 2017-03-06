package fdm

import (
	"strconv"
	"strings"
)

type VAT struct {
	Percentage  float64
	FixedAmount float64
}

// PercentageString formats VAT percentage to an FDM string
func (v *VAT) PercentageString() string {
	// percentage should be 4 numerical letters, format: xxyy where y is numbers after decimal point
	per := strconv.FormatFloat(v.Percentage, 'f', 2, 64)
	per = strings.Replace(per, ".", "", 1)
	for len(per) < 4 {
		per = " " + per
	}

	return per
}

// FixedAmountString formats VAT amount to an FDM string
func (v *VAT) FixedAmountString() string {
	amount := strconv.FormatFloat(v.FixedAmount, 'f', 2, 64)
	amount = strings.Replace(amount, ".", "", 1)
	// amount should be 11 numerical letters, format: xxxxxxxxx.yy where yy are numbers after the decimal point
	if v.FixedAmount == 0 {
		amount = "000"
	}
	for len(amount) < 11 {
		amount = " " + amount
	}

	return amount
}
