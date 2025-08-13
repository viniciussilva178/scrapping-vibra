package scraper

import (
	"fmt"
	"os"
	"scraper/db"
	"scraper/models"
	"scraper/pkg"

	"github.com/go-rod/rod"
)

func ProcessarXML(page *rod.Page, increment string, op *db.Operation) (int, error) {
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
