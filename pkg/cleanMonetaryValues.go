package pkg

import (
	"regexp"
	"strings"
)

func CleanMonetaryValue(value string) string {
	value = strings.TrimSpace(value)
	value = regexp.MustCompile(`[R$\s]`).ReplaceAllString(value, "")
	value = regexp.MustCompile(`[\(\)\-]`).ReplaceAllString(value, "")
	value = strings.ReplaceAll(value, ",", ".")
	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		integerPart := strings.Join(parts[:len(parts)-1], "")
		decimalPart := parts[len(parts)-1]
		value = integerPart + "." + decimalPart
	}
	if value == "" {
		return "0"
	}
	return strings.TrimSpace(value)
}
