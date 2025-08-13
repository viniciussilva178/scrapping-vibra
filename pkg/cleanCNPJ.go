package pkg

import (
	"regexp"
)

func CleanCNPJ(cnpj string) string {
	if cnpj == "" {
		return ""
	}
	re := regexp.MustCompile(`\D`)
	return re.ReplaceAllString(cnpj, "")
}
