package scraper

import (
	"fmt"
	"regexp"
	"scraper/internal/services"
	"scraper/models"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func ScrapeVibra(user, password string) ([]models.Document, error) {

	//  operation := db.NewOperation()

	// Pegando usuario autenticado
	authenticatedPage, browser, err := services.AuthenticateVibra(user, password)
	if err != nil {
		return nil, fmt.Errorf("erro ao realizar autenticação no portal Vibra:%s", err)

	}
	defer browser.MustClose()
	defer authenticatedPage.MustClose()

	// Puxar Html contendo tabela com os dados
	html, err := authenticatedPage.HTML()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter HTML:%s", err)

	}
	fmt.Println(html)
	authenticatedPage.MustWaitStable().MustElement("#dtListaDocumentos2").MustWaitVisible()

	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("erro ao percorrer html: %s", err)

	}

	var documents []models.Document
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
				document.Boleto = `https://cn.vibraenergia.com.br/cn//comercio/extratodoclientenovo/imprimirLoteExtrato?numeroDocumento=;` + document.Documento + `-1;`

				documents = append(documents, document)
			}
		})
	})

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limite de 5 Downloads simultaneos
	for i := range documents {
		wg.Add(1)
		go func(i int) {
			defer wg.Done() // adquirir slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Liberar slot

			pdfBytes, err := services.DownloadPDF(authenticatedPage, documents[i].Boleto)
			if err != nil {
				fmt.Printf("Erro ao fazer donload do PDF para documento %s: %d bytes\n", documents[i].Documento, err)
				return
			}
			documents[i].Conteudo = pdfBytes
			fmt.Printf("PDF baixado para documento:  %v", documents[i])
		}(i)
	}

	wg.Wait()

	fmt.Printf("=== PROCESSAMENTO CONCLUÍDO ===\n")
	return documents, nil

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
