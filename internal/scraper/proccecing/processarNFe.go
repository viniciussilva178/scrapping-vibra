package scraper

import (
	"fmt"
	"log"
	"os"
	"scraper/db"
	"scraper/models"
	"scraper/pkg"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
)

func ProcessarNFe(page *rod.Page, increment, numeroFaturaOriginal string, op *db.Operation) (int, string, error) {
	var id int
	// nfePath := filepath.Join("./", "nfe_"+increment+".pdf")

	html, err := page.HTML()
	if err != nil {
		return 0, "", fmt.Errorf("erro ao obter HTML: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return 0, "", fmt.Errorf("erro ao parsear HTML: %v", err)
	}

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
		log.Printf("Erro ao buscar link da NFe para increment %s: %v", increment, err)
	} else if downloadURL.Value.String() != "" {
		pdfURL = downloadURL.Value.String()
		log.Printf("--Usando link dinâmico para NFe: %s", pdfURL)
	}

	// Baixar NFe
	fileName := "nfe_" + increment + ".pdf"
	err = pkg.DownloadPDFToFile(page, pdfURL, fileName, "https://cn.vibraenergia.com.br/cn/comercio/notafiscaleletronicanova/")
	if err != nil {
		return 0, "", fmt.Errorf("erro ao baixar NFe para increment %s: %v", increment, err)
	}

	// Ler NFe como bytes
	data, err := os.ReadFile(fileName)
	if err != nil {
		return 0, "", fmt.Errorf("erro ao ler NFe para increment %s: %v", increment, err)
	}

	// Selecionar a linha específica baseada no índice (increment)
	index, err := strconv.Atoi(increment)
	if err != nil {
		os.Remove(fileName)
		return 0, "", fmt.Errorf("increment %s inválido: %v", increment, err)
	}
	rowSelector := fmt.Sprintf("table.table-condensed.table-responsive.table-hover.col-md-12 tbody tr:nth-child(%d)", index)
	row := doc.Find(rowSelector)
	if row.Length() == 0 {
		os.Remove(fileName)
		return 0, "", fmt.Errorf("linha %d não encontrada na tabela de NF-es", index)
	}

	cells := row.Find("td")
	if cells.Length() < 5 { // Garantir até Eq(4)
		os.Remove(fileName)
		return 0, "", fmt.Errorf("células insuficientes na linha %d da tabela de NF-es", index)
	}

	str := strings.TrimSpace(cells.Eq(4).Text())
	log.Printf("NumeroNFE extraído para increment %s: %s", increment, str) // Log para debug
	var numeroNFE *string
	if str != "" {
		numeroNFE = &str
	}

	// Salvar documento no banco
	id, err = op.CreateOperationVibraDocumento(&models.Document{
		Tipo:              3,
		ConteudoDocumento: data,
		NumeroNFE:         numeroNFE,
	}, 3)
	if err != nil {
		os.Remove(fileName)
		return 0, "", fmt.Errorf("erro ao salvar documento para increment %s: %v", increment, err)
	}

	// Usar numeroFaturaOriginal para garantir correspondência com vibra.conta_pagar
	numeroFatura := numeroFaturaOriginal
	log.Printf("Usando numeroFaturaOriginal para increment %s: %s", increment, numeroFatura)

	// Limpeza do arquivo temporário
	if err := os.Remove(fileName); err != nil {
		log.Printf("erro ao remover arquivo temporário %s: %v", fileName, err)
	}

	page.Timeout(120 * time.Second)

	return id, numeroFatura, nil
}
