package scraper

import (
	"fmt"
	"log"
	"os"
	"scraper/db"
	"scraper/models"
	"scraper/pkg"

	"github.com/go-rod/rod"
)

func ProcessarNFe(page *rod.Page, increment string, op *db.Operation) (int, error) {
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
