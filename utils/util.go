package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func ConvertToMB(sizeStr string) (float64, error) {
	conversionFactors := map[string]float64{
		"KB": 1.0 / 1024.0,
		"MB": 1.0,
		"GB": 1024.0,
		"TB": 1024.0 * 1024.0,
	}

	numberStr := ""
	unit := ""

	for _, ch := range sizeStr {
		if ch >= '0' && ch <= '9' || ch == '.' {
			numberStr += string(ch)
		} else {
			unit += string(ch)
		}
	}

	number, err := strconv.ParseFloat(numberStr, 64)
	if err != nil {
		return 0, err
	}

	factor, ok := conversionFactors[strings.ToUpper(unit)]
	if !ok {
		return 0, fmt.Errorf("unsupported unit: %s", unit)
	}

	return number * factor, nil
}

func ConvertMBToString(mbValue float64) string {
	thresholds := map[string]float64{
		"TB": 1024.0 * 1024.0,
		"GB": 1024.0,
		"MB": 1.0,
		"KB": 1.0 / 1024.0,
	}

	for unit, threshold := range thresholds {
		if mbValue >= threshold {
			value := mbValue / threshold
			return fmt.Sprintf("%.2f%s", value, unit)
		}
	}

	return fmt.Sprintf("%.2fMB", mbValue)
}
