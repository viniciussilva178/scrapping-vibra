package scraper

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"scraper/db"
	util "scraper/internal/utils"
	"scraper/models"
	"scraper/pkg"

	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
)

func ScrapeDocsVibra(user, password string, op *db.Operation) ([]models.ContasAPagar, error) {
	inicio := time.Now()

	// 1. Autenticação
	authenticatedPage, browser, err := pkg.AuthenticateVibra(user, password)
	if err != nil {
		return nil, fmt.Errorf("erro ao realizar autenticação: %v", err)
	}
	defer func() {
		if browser != nil {
			browser.MustClose()
		}
	}()
	defer func() {
		if authenticatedPage != nil {
			authenticatedPage.MustClose()
		}
	}()

	// 2. Navegação para contas a pagar
	if err := rod.Try(func() {
		authenticatedPage.MustElement("#dtListaDocumentos2").MustWaitVisible()
		authenticatedPage.MustElement(`.marcarTodos`).MustClick()
		authenticatedPage.MustElement(`#aplica_`).MustClick()
		authenticatedPage.MustWaitStable().MustElement(`.modal.fade.in .modal-content .modal-body`).MustWaitVisible()
	}); err != nil {
		return nil, fmt.Errorf("erro ao navegar: %v", err)
	}
	time.Sleep(5 * time.Second)

	// 3. Obter HTML
	html, err := authenticatedPage.HTML()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter HTML: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear HTML: %v", err)
	}

	// 4. Baixar PDF de boletos
	pdfPath, exists := doc.Find("#divUrlBoleto object").Attr("data")
	if !exists {
		return nil, fmt.Errorf("link do PDF não encontrado")
	}

	pdfURL := pdfPath
	if strings.HasPrefix(pdfURL, "/") {
		pdfURL = "https://cn.vibraenergia.com.br" + pdfURL
	}

	err = pkg.DownloadPDFToFile(authenticatedPage, pdfURL, "boletos.pdf", "https://cn.vibraenergia.com.br/cn/comercio/extratodoclientenovo/")
	if err != nil {
		return nil, fmt.Errorf("erro ao baixar PDF: %v", err)
	}

	// 5. Dividir PDF em páginas individuais
	err = util.OrdersPDFS()
	if err != nil {
		return nil, fmt.Errorf("erro ao dividir PDF: %v", err)
	}

	// 6. Processar contas
	contasPagar, err := processarContas(doc, authenticatedPage, op)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar contas: %v", err)
	}

	fmt.Printf("--Processamento concluído para %s (%d contas)\n", user, len(contasPagar))
	fmt.Printf("--Tempo total: %s\n", time.Since(inicio))

	return contasPagar, nil
}

func processarContas(doc *goquery.Document, page *rod.Page, op *db.Operation) ([]models.ContasAPagar, error) {
	var contasPagar []models.ContasAPagar

	if err := rod.Try(func() {
		page.MustElement("#btnCloseModalImprimir").MustClick()
		time.Sleep(1 * time.Second)
		page.MustElement(`#menuAcessoRevendedoNFe`).MustClick()
		page.MustWaitStable().MustElement("#downloadNotaFiscalForm").MustWaitVisible()

		// Configurar filtro por data
		dataOntem := time.Now().AddDate(0, 0, -1).Format("02/01/2006")
		page.MustEval(`date => {
			const el = document.querySelector('#dataEmissaoInicial');
			el.value = date;
			el.dispatchEvent(new Event('input', { bubbles: true }));
			el.dispatchEvent(new Event('change', { bubbles: true }));
		}`, dataOntem)

		page.MustElement(`#btListar`).MustClick()
		page.MustWaitStable().MustElement(`#panelNotaFiscal`).MustWaitVisible()
	}); err != nil {
		log.Printf("Aviso: navegação NFes falhou: %v", err)
	}

	// Processar cada linha
	doc.Find("table#dtListaDocumentos2 tbody tr").Each(func(idx int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() < 1 {
			log.Printf("Linha %d ignorada - células insuficientes", idx+1)
			return
		}

		conta := models.ContasAPagar{}
		increment := strconv.Itoa(idx + 1)

		// Extração de dados
		conta.CNPJBeneficiario = pkg.CleanCNPJ(strings.TrimSpace(cells.Eq(10).Text()))
		conta.CNPJPagador = pkg.CleanCNPJ(strings.TrimSpace(cells.Eq(11).Text()))

		re := regexp.MustCompile(`\d+`)
		match := re.FindAllString(cells.Eq(1).Text(), -1)
		if len(match) > 0 {
			conta.NumeroDocumento = match[len(match)-1]
		}

		conta.NumeroFatura = strings.TrimSpace(cells.Eq(2).Text())
		conta.DataEmissao = strings.TrimSpace(cells.Eq(3).Text())
		conta.DataVencimento = strings.TrimSpace(cells.Eq(4).Text())

		if valor, err := pkg.ParseFloatWithError(cells.Eq(5).Text()); err == nil {
			conta.ValorDocumento = valor
		}
		if juros, err := pkg.ParseFloatWithError(cells.Eq(6).Text()); err == nil {
			conta.ValorJuros = juros
		}
		if multas, err := pkg.ParseFloatWithError(cells.Eq(7).Text()); err == nil {
			conta.ValorMulta = multas
		}
		if deducoes, err := pkg.ParseFloatWithError(cells.Eq(8).Text()); err == nil {
			conta.ValorDeducao = deducoes
		}
		if total, err := pkg.ParseFloatWithError(cells.Eq(9).Text()); err == nil {
			conta.ValorTotal = total
		}

		// Processar boleto
		if err := processarBoleto(&conta, increment, op); err != nil {
			log.Printf("Erro no boleto linha %d: %v", idx+1, err)
			return
		}

		// Processar NFe
		if idNF, err := processarNFe(page, increment, op); err == nil {
			conta.IDNotaFiscal = &idNF
		} else {
			log.Printf("NF-e falhou linha %d: %v", idx+1, err)

		}

		// Processar XML
		if idXML, err := processarXML(page, increment, op); err == nil {
			conta.IDXML = &idXML
		} else {
			log.Printf("XML falhou linha %d: %v", idx+1, err)
		}

		contasPagar = append(contasPagar, conta)
	})

	if len(contasPagar) == 0 {
		return nil, fmt.Errorf("nenhuma conta válida processada")
	}

	return contasPagar, nil
}

func processarBoleto(conta *models.ContasAPagar, increment string, op *db.Operation) error {
	boletoPath := filepath.Join("docs", "boletos_"+increment+".pdf")

	// Ler boleto como bytes
	data, err := os.ReadFile(boletoPath)
	if err != nil {
		return fmt.Errorf("erro ao ler boleto: %v", err)
	}

	// Extrair linha digitável
	code, err := pkg.GetBarcodeFromFile(boletoPath, 1)
	if err != nil {
		log.Printf("Aviso: linha digitável não encontrada: %v", err)
	} else {
		conta.LinhaDigitavel = code
	}

	// Salvar documento no banco
	docBoleto := models.Document{
		Tipo:              1,
		ConteudoDocumento: data,
	}
	idBoleto, err := op.CreateOperationVibraDocumento(&docBoleto, 1)
	if err != nil {
		return fmt.Errorf("erro ao salvar boleto: %v", err)
	}
	conta.IDBoleto = &idBoleto

	re := regexp.MustCompile(`\boletos[0-100]\.pdf`)

	defer os.RemoveAll(re.String())

	return nil
}

func processarNFe(page *rod.Page, increment string, op *db.Operation) (int, error) {
	// Construir URL da NFe
	pdfURL := "https://cn.vibraenergia.com.br/cn/comercio/notafiscaleletronicanova/downloadNotaFiscal?tipoArquivo=pdf&idLinha=" +
		increment + "&tipoDocumento=danfe"

	// Tentar obter link dinâmico
	jsScript := fmt.Sprintf(
		`() => {
			const link = document.querySelector('a[href*="idLinha=%s"]');
			return link ? link.href : '';
		}`, increment)

	downloadURL, err := page.Eval(jsScript)
	if err != nil {
		log.Printf(" Erro ao buscar link da NFe: %v", err)
	} else if downloadURL.Value.String() != "" {
		pdfURL = downloadURL.Value.String()
		log.Printf("--Usando link dinâmico para NFe: %s", pdfURL)
	}

	// Baixar NFe
	fileName := "nfe_" + increment + ".pdf"
	err = pkg.DownloadPDFToFile(page, pdfURL, fileName, page.MustInfo().URL)
	if err != nil {
		return 0, fmt.Errorf("erro ao baixar NFe: %v", err)
	}

	// Ler NFe como bytes
	data, err := os.ReadFile(fileName)
	if err != nil {
		return 0, fmt.Errorf("erro ao ler NFe: %v", err)
	}

	// Salvar documento no banco
	id, err := op.CreateOperationVibraDocumento(&models.Document{
		Tipo:              2,
		ConteudoDocumento: data,
	}, 2)

	// Limpeza do arquivo temporário
	defer os.Remove(fileName)

	return id, err
}

func processarXML(page *rod.Page, increment string, op *db.Operation) (int, error) {
	// Construir URL do XML
	xmlURL := "https://cn.vibraenergia.com.br/cn/comercio/notafiscaleletronicanova/downloadNotaFiscal?tipoArquivo=xml&idLinha=" +
		increment + "&tipoDocumento=nfe"

	// Baixar XML
	fileName := "nfe_" + increment + ".xml"
	err := pkg.DownloadXMLToFile(page, xmlURL, fileName, page.MustInfo().URL)
	if err != nil {
		return 0, fmt.Errorf("erro ao baixar XML: %v", err)
	}

	// Ler XML como bytes
	data, err := os.ReadFile(fileName)
	if err != nil {
		return 0, fmt.Errorf("erro ao ler XML: %v", err)
	}

	// Salvar documento no banco
	id, err := op.CreateOperationVibraDocumento(&models.Document{
		Tipo:              3,
		ConteudoDocumento: data,
	}, 3)

	// Limpeza do arquivo temporário
	defer os.Remove(fileName)

	return id, err
}
