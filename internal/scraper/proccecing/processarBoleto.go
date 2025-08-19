package scraper

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"scraper/db"
	"scraper/models"
	"scraper/pkg"
)

func ProcessarBoleto(conta *models.ContasAPagar, increment string, op *db.Operation) error {
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

	cnpj, err := pkg.GetCNPJS(boletoPath, 1)

	if err != nil {
		log.Printf("Aviso: linha digitável não encontrada: %v", err)
	} else {
		formattedBeneficiario := pkg.CleanCNPJ(cnpj[0])
		formattedPagador := pkg.CleanCNPJ(cnpj[2])

		conta.CNPJBeneficiario = formattedBeneficiario
		conta.CNPJPagador = formattedPagador
	}

	// Salvar documento no banco
	docBoleto := models.Document{
		Tipo:              2,
		ConteudoDocumento: data,
	}
	idBoleto, err := op.CreateOperationVibraDocumento(&docBoleto, 2)
	if err != nil {
		return fmt.Errorf("erro ao salvar boleto: %v", err)
	}
	conta.IDBoleto = &idBoleto

	re := regexp.MustCompile(`\boletos_[0-100]\.pdf`)

	defer os.RemoveAll(re.String())

	return nil
}
