package scraper

import (
	"fmt"
	"log"
	"regexp"
	"scraper/internal/services"
	"scraper/models"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func ScrapeVibra(user, password string) ([]models.Document, error) {
	inicio := time.Now()

	authenticatedPage, browser, err := services.AuthenticateVibra(user, password)
	if err != nil {
		return nil, fmt.Errorf("erro ao realizar autenticação no portal Vibra: %v", err)
	}
	defer browser.MustClose()
	defer authenticatedPage.MustClose()

	// Aguardar até a tabela ser carregada
	authenticatedPage.MustWaitStable().MustElement("#dtListaDocumentos2").MustWaitVisible()

	checkAll := authenticatedPage.MustElement(`.marcarTodos`)
	checkAll.MustClick()

	authenticatedPage.MustElement(`#aplica_`).MustClick()
	authenticatedPage.MustWaitStable().MustElement(`.modal.fade.in .modal-content .modal-body`).MustWaitVisible()
	time.Sleep(10 * time.Second)

	// Puxar HTML contendo tabela com os dados
	html, err := authenticatedPage.HTML()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter HTML: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear HTML: %v", err)
	}

	// Capturar link do PDF único
	pdfPath, exists := doc.Find("#divUrlBoleto object").Attr("data")
	if !exists {
		return nil, fmt.Errorf("link do PDF não encontrado no HTML")
	}

	// Garantir URL absoluta
	pdfURL := pdfPath
	if strings.HasPrefix(pdfURL, "/") {
		pdfURL = "https://cn.vibraenergia.com.br" + pdfURL
	}

	err = services.DownloadPDFToFile(authenticatedPage, pdfURL, "boletos.pdf")
	if err != nil {
		panic(err)
	}

	err = services.OrdersPDFS()
	if err != nil {
		fmt.Println("Eror ao Executar Ordenação de PDF")
	}

	fmt.Println("URL do PDF único:", pdfURL)

	var documents []models.Document

	doc.Find("table#dtListaDocumentos2").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			var document models.Document
			var bytesBoleto []byte
			var code string
			var htmlPDF string

			indexOf := i + 1
			increment := strconv.Itoa(indexOf)

			cells := row.Find("td")
			if cells.Length() >= 10 { // Verificar se tem células suficientes
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

				bytesBoleto, err = services.ConvertBoletoToBytes(&document, increment)
				if err != nil {
					log.Fatal("Erro ao atribuir Boleto em Bytes a coluna conteudo", err)
				}

				htmlPDF, code, err = services.PdfToHTML(bytesBoleto)
				if err != nil {
					log.Fatal("Erro ao converter boleto em HTML", err)
					fmt.Println(htmlPDF)
				}

				fmt.Println(htmlPDF)
				document.Conteudo = bytesBoleto
				document.LinhaDigitavel = code

				documents = append(documents, document)
			}
		})
	})

	// Processamento concorrente de PDFs e extração da linha digitável
	/*	var wg sync.WaitGroup
		semaphore := make(chan struct{}, 10) // Limite de 5 requisições simultâneas
		for i := range documents {
			wg.Add(1)

			go func(i int) {
				defer wg.Done()
				semaphore <- struct{}{}        // Adquirir slot no semáforo
				defer func() { <-semaphore }() // Liberar slot

				// Atribuir Boleto a Conteudo
				documents[i].Conteudo =

				// documents[i].Conteudo = pdfBytes
				// fmt.Printf("PDF baixado para documento %s: %d bytes\n", documents[i].Documento, len(pdfBytes))

				// Extrair linha digitável do PDF
				linhaDigitavel, extractedText, err := extractLinhaDigitavelFromPDF(pdfBytes, documents[i].Documento)
				if err != nil {
					fmt.Printf("Erro ao extrair linha digitável do PDF para documento %s: %v\n", documents[i].Documento, err)
					// Salvar texto extraído para debug
					os.WriteFile(fmt.Sprintf("pdf_text_%s.txt", documents[i].Documento), []byte(extractedText), 0644)
					// Logar primeiros 1000 caracteres do texto extraído
					logText := extractedText
					if len(logText) > 1000 {
						logText = logText[:1000]
					}
					fmt.Printf("Texto extraído do PDF (primeiros 1000 caracteres) para documento %s: %s\n", documents[i].Documento, logText)
					return
				}
				documents[i].LinhaDigitavel = linhaDigitavel
				fmt.Printf("Linha digitável extraída do PDF para documento %s: %s\n", documents[i].Documento, linhaDigitavel)

			}(i)
		}

		wg.Wait()
	*/

	fmt.Printf("=== PROCESSAMENTO CONCLUÍDO PARA USUÁRIO %s ===\n", user)
	tempoCorrido := time.Since(inicio)
	fmt.Printf("Tempo de execução: %s\n", tempoCorrido)
	return documents, nil

}
func cleanMonetaryValue(value string) string {
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
