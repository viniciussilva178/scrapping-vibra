package pkg

import (
	"fmt"
	"strconv"
)

func ParseFloatWithError(value string) (float64, error) {
	cleanedValue := CleanMonetaryValue(value)
	if cleanedValue == "" {
		return 0.0, fmt.Errorf("valor vazio")
	}
	result, err := strconv.ParseFloat(cleanedValue, 64)
	if err != nil {
		return 0.0, fmt.Errorf("erro ao converter '%s': %v", cleanedValue, err)
	}
	return result, nil
}
