package scraper

import (
	"fmt"
	"log"
	"regexp"
	"scraper/db"
	"scraper/models"
	"scraper/pkg"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
)

func ProcessarContas(doc *goquery.Document, page *rod.Page, op *db.Operation) ([]models.ContasAPagar, error) {
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
		increment := strconv.Itoa(idx)

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

		if err := ProcessarBoleto(&conta, increment, op); err != nil {
			log.Printf("Erro no boleto linha %d: %v", idx+1, err)
			return
		}

		conta.IDXML = nil

		// Processar boleto
		/*

			// Processar NFe
			if idNF, err := ProcessarNFe(page, increment, op); err == nil {
				conta.IDNotaFiscal = nil
			} else {
				log.Printf("NF-e falhou linha %d: %v", idx+1, err)

			}

			// Processar XML
			if idXML, err := ProcessarXML(page, increment, op); err == nil {
				conta.IDXML = &idXML
			} else {
				log.Printf("XML falhou linha %d: %v", idx+1, err)
			}
		*/

		contasPagar = append(contasPagar, conta)
	})

	if len(contasPagar) == 0 {
		return nil, fmt.Errorf("nenhuma conta válida processada")
	}

	return contasPagar, nil
}
