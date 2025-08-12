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

	authenticatedPage.MustWaitStable().MustElement("#dtListaDocumentos2").MustWaitVisible()

	checkAll := authenticatedPage.MustElement(`.marcarTodos`)
	checkAll.MustClick()

	authenticatedPage.MustElement(`#aplica_`).MustClick()
	authenticatedPage.MustWaitStable().MustElement(`.modal.fade.in .modal-content .modal-body`).MustWaitVisible()
	time.Sleep(10 * time.Second)

	html, err := authenticatedPage.HTML()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter HTML: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear HTML: %v", err)
	}

	pdfPath, exists := doc.Find("#divUrlBoleto object").Attr("data")
	if !exists {
		return nil, fmt.Errorf("link do PDF não encontrado no HTML")
	}

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

			indexOf := i + 1
			increment := strconv.Itoa(indexOf)

			cells := row.Find("td")
			if cells.Length() >= 10 {
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

				code, err := services.GetBarcodeFromFile("./docs/boletos_"+increment+".pdf", 1)
				if err != nil {
					log.Fatalf("Erro ao pegar linha digitável: %v", err)
				}
				document.LinhaDigitavel = code

				document.Conteudo = bytesBoleto
				document.LinhaDigitavel = string(code)

				documents = append(documents, document)
			}
		})
	})

	services.ClearPDF("./docs")

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
