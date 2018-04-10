package entity

import "fmt"

func FloatToString(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}
