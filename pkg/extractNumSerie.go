package pkg

import (
	"fmt"
	"regexp"

	"github.com/ledongthuc/pdf"
)

func ExtractNumSerie(caminho string, pageNum int) (string, error) {
	f, r, err := pdf.Open(caminho)
	if err != nil {
		return "", fmt.Errorf("erro ao abrir PDF: %w", err)
	}
	defer f.Close()

	if pageNum < 1 || pageNum > r.NumPage() {
		return "", fmt.Errorf("página %d não existe no PDF", pageNum)
	}

	text, err := r.Page(pageNum).GetPlainText(nil)
	if err != nil {
		return "", fmt.Errorf("erro ao extrair texto da página %d: %w", pageNum, err)
	}

	re := regexp.MustCompile(``)
	codigo := re.FindString(text)

	cleanLinhaDigitavel := CleanCNPJ(codigo)
	if codigo == "" {
		return "", fmt.Errorf("nenhuma linha digitável encontrada na página %d", pageNum)
	}

	return cleanLinhaDigitavel, nil
}
