package utils

import (
	"fmt"
	"regexp"

	"github.com/ledongthuc/pdf"
)

func GetBarcodeFromFile(caminho string, pageNum int) (string, error) {
	f, r, err := pdf.Open(caminho)
	defer f.Close()
	if err != nil {
		return "", fmt.Errorf("erro ao abrir PDF: %w", err)
	}

	if pageNum < 1 || pageNum > r.NumPage() {
		return "", fmt.Errorf("página %d não existe no PDF", pageNum)
	}

	text, err := r.Page(pageNum).GetPlainText(nil)
	if err != nil {
		return "", fmt.Errorf("erro ao extrair texto da página %d: %w", pageNum, err)
	}

	re := regexp.MustCompile(`\d{5}\.\d{5}\s\d{5}\.\d{6}\s\d{5}\.\d{6}\s\d\s\d{14}`)
	codigo := re.FindString(text)
	if codigo == "" {
		return "", fmt.Errorf("nenhuma linha digitável encontrada na página %d", pageNum)
	}

	return codigo, nil
}
