package helpers

import "strconv"
import "fmt"

func Round(x, unit float64) float64 {
	formatted, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", x), 64)
	return formatted
}

func ConvertToInt(x float64) int64 {
	//stringToFormat := fmt.Sprintf("%0f", x*100)
	//log.Println("String to format", stringToFormat)
	//formatted, _ := strconv.ParseInt(stringToFormat, 10, 64)
	//log.Println("Formatted", formatted)

	return int64(x * 100)
}
