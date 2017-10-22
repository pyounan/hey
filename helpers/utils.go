package helpers

import "strconv"
import "fmt"

func Round(x, unit float64) float64 {
	var val float64
	if x > 0 {
		val = float64(int64(x/unit+0.5)) * unit
	} else {
		val = float64(int64(x/unit+0.5)) * unit
	}
	formatted, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", val), 64)
	return formatted
}

func ConvertToInt(x float64) int64 {
	//stringToFormat := fmt.Sprintf("%0f", x*100)
	//log.Println("String to format", stringToFormat)
	//formatted, _ := strconv.ParseInt(stringToFormat, 10, 64)
	//log.Println("Formatted", formatted)

	return int64(x * 100)
}
