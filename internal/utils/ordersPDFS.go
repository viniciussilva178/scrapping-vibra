package util

import (
	"log"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func OrdersPDFS() error {
	file := "boletos.pdf"
	dirPath := "./docs"

	err := api.SplitFile(file, dirPath, 1, nil)
	if err != nil {
		log.Fatal("Erro ao Dividir Boletos", err)
		return err
	}

	return nil
}
