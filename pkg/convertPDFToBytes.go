package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"scraper/models"
)

func ConvertPDFToBytes(document *models.Document, increment, path string) ([]byte, error) {
	files, err := filepath.Glob(path + increment + ".pdf")
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar arquivos: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("nenhum PDF encontrado")
	}

	file := files[0]

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo %s: %w", file, err)
	}

	document.ConteudoDocumento = data
	fmt.Printf("Arquivo %s convertido em %d bytes\n", filepath.Base(file), len(data))

	return data, nil
}
