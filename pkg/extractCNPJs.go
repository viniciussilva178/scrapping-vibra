package pkg

import (
	"fmt"
	"regexp"

	"github.com/ledongthuc/pdf"
)

func GetCNPJS(caminho string, pageNum int) ([]string, error) {

	f, r, err := pdf.Open(caminho)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir PDF: %w", err)
	}
	defer f.Close()

	if pageNum < 1 || pageNum > r.NumPage() {
		return nil, fmt.Errorf("página %d não existe no PDF", pageNum)
	}

	text, err := r.Page(pageNum).GetPlainText(nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao extrair texto da página %d: %w", pageNum, err)
	}

	re := regexp.MustCompile(`\d{2}\.\d{3}\.\d{3}\/\d{4}\-\d{2}`)
	cnpj := re.FindAllString(text, -1)

	if cnpj == nil {
		return nil, fmt.Errorf("nenhuma linha digitável encontrada na página %d", pageNum)
	}

	return cnpj, nil
}
