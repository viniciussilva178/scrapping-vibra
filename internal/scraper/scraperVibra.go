package scraper

import (
	"fmt"
	"regexp"
	"scraper/internal/db"
	"scraper/internal/services"
	"scraper/models"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var documents []models.Document

func ScrapeVibra(user, password string) {

	operation := db.NewOperation()

	authenticatedPage, browser, err := services.AuthenticateVibra(user, password)
	if err != nil {
		fmt.Println("Erro ao realizar autenticação no portal Vibra: ", err)
		return
	}
	defer browser.MustClose()
	defer authenticatedPage.MustClose()

	// Puxar Html contendo tabela com os dados
	html, err := authenticatedPage.HTML()
	if err != nil {
		fmt.Println("Erro ao obter HTML:", err)
		return
	}
	fmt.Println(html)

	// Extrair URL do PDF do HTML
	pdfURL := "https://cn.vibraenergia.com.br/cn/comercio/extratodoclientenovo/imprimirLoteExtrato?numeroDocumento=;0092225755-1"

	time.Sleep(20 * time.Second)

	// Gerar nome do arquivo com timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("vibra_documentos_%s.pdf", timestamp)

	// Fazer download do PDF
	err = services.DownloadPDF(authenticatedPage, pdfURL, filename)
	if err != nil {
		fmt.Printf("Erro ao fazer download do PDF: %v\n", err)
	} else {
		fmt.Printf("=== DOWNLOAD CONCLUÍDO: %s ===\n", filename)
	}

	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		fmt.Printf("Erro ao percorrer html: %s", err)
		return
	}

	doc.Find("table#dtListaDocumentos2").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {

			var document models.Document

			cells := row.Find("td")
			if cells.Length() >= 9 { // Verificar se tem células suficientes

				re := regexp.MustCompile(`\d+`)
				match := re.FindAllString(cells.Eq(1).Text(), -1)
				if len(match) > 0 {
					document.Documento = match[len(match)-1]
				}

				document.NF = strings.TrimSpace(cells.Eq(2).Text())
				document.Emissao = strings.TrimSpace(cells.Eq(3).Text())
				document.Vencimento = strings.TrimSpace(cells.Eq(4).Text())

				if valor, err := parseFloatWithError(cells.Eq(5).Text()); err != nil {
					fmt.Printf("Erro ao converter Valor: %v\n", err)
				} else {
					document.Valor = valor
				}

				if juros, err := parseFloatWithError(cells.Eq(6).Text()); err != nil {
					fmt.Printf("Erro ao converter Juros: %v\n", err)
				} else {
					document.Juros = juros
				}

				if multas, err := parseFloatWithError(cells.Eq(7).Text()); err != nil {
					fmt.Printf("Erro ao converter Multas: %v\n", err)
				} else {
					document.Multas = multas
				}

				if deducoes, err := parseFloatWithError(cells.Eq(8).Text()); err != nil {
					fmt.Printf("Erro ao converter Deduções: %v\n", err)
				} else {
					document.Deducoes = deducoes
				}

				if total, err := parseFloatWithError(cells.Eq(9).Text()); err != nil {
					fmt.Printf("Erro ao converter Total: %v\n", err)
				} else {
					document.Total = total
				}
				document.Boleto = `https://cn.vibraenergia.com.br/cn/comercio/extratodoclientenovo/imprimirLoteExtrato?numeroDocumento=;` + document.Documento + `-1`

				documents = append(documents, document)
			}
		})
	})

	for _, d := range documents {

		message, err := operation.CreateOperationVibra(&d)
		if err != nil {
			fmt.Printf("Erro ao salvar documento %s: %v\n", d.Documento, err)
		} else {
			fmt.Printf("Documento %s: %s\n", d.Documento, message)
		}
	}

	fmt.Printf("=== PROCESSAMENTO CONCLUÍDO ===\n")

}

func cleanMonetaryValue(value string) string {
	// Remove espaços
	value = strings.TrimSpace(value)

	// Remove símbolos de moeda (R$, $, etc.)
	value = regexp.MustCompile(`[R$\s]`).ReplaceAllString(value, "")

	// Remove caracteres especiais como hífens, parênteses, etc.
	value = regexp.MustCompile(`[\(\)\-]`).ReplaceAllString(value, "")

	// Remove pontos de milhares (mantém apenas o último ponto como separador decimal)
	// Primeiro, substitui vírgulas por pontos se houver
	value = strings.ReplaceAll(value, ",", ".")

	// Se há mais de um ponto, mantém apenas o último como separador decimal
	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		// Reconstrói com apenas o último ponto como decimal
		integerPart := strings.Join(parts[:len(parts)-1], "")
		decimalPart := parts[len(parts)-1]
		value = integerPart + "." + decimalPart
	}

	// Se o valor estiver vazio após limpeza, retorna "0"
	if value == "" {
		return "0"
	}

	return strings.TrimSpace(value)
}

// Função para converter string para float com tratamento de erro
func parseFloatWithError(value string) (float64, error) {
	cleanedValue := cleanMonetaryValue(value)
	fmt.Printf("Valor original: '%s' -> Limpo: '%s'\n", value, cleanedValue)

	if cleanedValue == "" {
		return 0.0, fmt.Errorf("valor vazio")
	}

	result, err := strconv.ParseFloat(cleanedValue, 64)
	if err != nil {
		return 0.0, fmt.Errorf("erro ao converter '%s': %v", cleanedValue, err)
	}

	return result, nil
}
