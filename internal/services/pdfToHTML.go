package services

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ledongthuc/pdf"
)

func PdfToHTML(pdfBytes []byte) (string, string, error) {
	reader := bytes.NewReader(pdfBytes)

	// Abre o PDF direto da memória
	doc, err := pdf.NewReader(reader, int64(len(pdfBytes)))
	if err != nil {
		return "", "", fmt.Errorf("erro ao abrir PDF: %w", err)
	}

	// Extrai o texto de todas as páginas
	var textBuf bytes.Buffer
	numPages := doc.NumPage()
	for i := 1; i <= numPages; i++ {
		p := doc.Page(i)
		if p.V.IsNull() {
			continue
		}
		content := p.V.Text()

		textBuf.WriteString(content)
	}

	text := textBuf.String()

	// Converte texto para HTML simples
	htmlTemplate := `<html><body><pre>{{.}}</pre></body></html>`
	tmpl, _ := template.New("pdf").Parse(htmlTemplate)
	var htmlBuf bytes.Buffer
	_ = tmpl.Execute(&htmlBuf, text)

	// Procura linha digitável no texto
	doct, _ := goquery.NewDocumentFromReader(strings.NewReader(htmlTemplate))

	code, err := doct.Find(".s1").Html()
	if err != nil {
		fmt.Println("Erro")
	}
	fmt.Println(code)

	return htmlBuf.String(), code, nil
}
