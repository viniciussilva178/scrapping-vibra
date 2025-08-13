package pkg

import (
	"fmt"
	"regexp"

	"github.com/ledongthuc/pdf"
)

func GetCNPJS(caminho string, pageNum int) (string, error) {

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

	re := regexp.MustCompile(`\d{2}\.\d{3}\.\d{3}\/d{4}\-\d{2}`)
	cnpj := re.FindString(text)
	if cnpj == "" {
		return "", fmt.Errorf("nenhuma linha digitável encontrada na página %d", pageNum)
	}

	fmt.Println(cnpj)

	return cnpj, nil
}
