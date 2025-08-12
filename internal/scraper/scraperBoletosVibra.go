package scraper

import (
	"fmt"
	"log"
	"regexp"
	"scraper/internal/utils"
	"scraper/models"
	"scraper/pkg"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func ScrapeBoletosVibra(user, password string) ([]models.Document, error) {

	//// Setar Autenticação
	inicio := time.Now()
	authenticatedPage, browser, err := pkg.AuthenticateVibra(user, password)
	if err != nil {
		return nil, fmt.Errorf("erro ao realizar autenticação no portal Vibra: %v", err)
	}
	defer browser.MustClose()
	defer authenticatedPage.MustClose()

	//// Acessar Contas a Pagar no VIBRA
	authenticatedPage.MustWaitStable().MustElement("#dtListaDocumentos2").MustWaitVisible()
	checkAll := authenticatedPage.MustElement(`.marcarTodos`)
	checkAll.MustClick()

	authenticatedPage.MustElement(`#aplica_`).MustClick()
	authenticatedPage.MustWaitStable().MustElement(`.modal.fade.in .modal-content .modal-body`).MustWaitVisible()
	time.Sleep(5 * time.Second)

	//// Buscar HTML da pagina
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

	//// Baixar e Organizar os PDFs separadamente
	pdfURL := pdfPath
	if strings.HasPrefix(pdfURL, "/") {
		pdfURL = "https://cn.vibraenergia.com.br" + pdfURL
	}

	err = utils.DownloadPDFToFile(authenticatedPage, pdfURL, "boletos.pdf", "https://cn.vibraenergia.com.br/cn/comercio/extratodoclientenovo/")
	if err != nil {
		panic(err)
	}

	err = utils.OrdersPDFS()
	if err != nil {
		fmt.Println("Eror ao Executar Ordenação de PDF")
	}

	documents, increment := runTable(doc)

	ScrapperNF(authenticatedPage, increment)
	pkg.ClearPDF("./docs")

	fmt.Printf("=== PROCESSAMENTO CONCLUÍDO PARA USUÁRIO %s ===\n", user)
	tempoCorrido := time.Since(inicio)
	fmt.Printf("Tempo de execução: %s\n", tempoCorrido)

	return documents, nil

}

// // Percorrer HTML da tabela para fazer os Scrapping dos dados
func runTable(doc *goquery.Document) ([]models.Document, string) {
	var documents []models.Document
	var indexOf int
	increment := strconv.Itoa(indexOf)

	doc.Find("table#dtListaDocumentos2").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			var document models.Document

			indexOf = i + 1

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

				if valor, err := pkg.ParseFloatWithError(cells.Eq(5).Text()); err != nil {
					fmt.Printf("Erro ao converter Valor: %v\n", err)
				} else {
					document.Valor = valor
				}

				if juros, err := pkg.ParseFloatWithError(cells.Eq(6).Text()); err != nil {
					fmt.Printf("Erro ao converter Juros: %v\n", err)
				} else {
					document.Juros = juros
				}

				if multas, err := pkg.ParseFloatWithError(cells.Eq(7).Text()); err != nil {
					fmt.Printf("Erro ao converter Multas: %v\n", err)
				} else {
					document.Multas = multas
				}

				if deducoes, err := pkg.ParseFloatWithError(cells.Eq(8).Text()); err != nil {
					fmt.Printf("Erro ao converter Deduções: %v\n", err)
				} else {
					document.Deducoes = deducoes
				}

				if total, err := pkg.ParseFloatWithError(cells.Eq(9).Text()); err != nil {
					fmt.Printf("Erro ao converter Total: %v\n", err)
				} else {
					document.Total = total
				}

				bytesBoleto, err := pkg.ConvertBoletoToBytes(&document, increment, "./docs/boletos_")
				if err != nil {
					log.Fatal("Erro ao atribuir Boleto em Bytes a coluna conteudo", err)
				}
				pkg.ClearPDF("./docs")
				time.Sleep(2 * time.Second)

				bytesNF, err := pkg.ConvertBoletoToBytes(&document, increment, "./docs/nfe_")
				if err != nil {
					log.Fatal("Erro ao atribuir NFe em Bytes a coluna nfeConteudo", err)
				}
				pkg.ClearPDF("./docs")

				code, err := utils.GetBarcodeFromFile("./docs/boletos_"+increment+".pdf", 1)
				if err != nil {
					log.Fatalf("Erro ao pegar linha digitável: %v", err)
				}
				document.LinhaDigitavel = code

				document.Conteudo = bytesBoleto
				document.NFConteudo = bytesNF
				document.LinhaDigitavel = string(code)

				documents = append(documents, document)
			}
		})
	})

	return documents, increment
}
