package services

import (
	"os"
)

func ClearPDF(path string) error {
	arquivos, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, arquivo := range arquivos {
		caminhoArquivo := path + string(os.PathSeparator) + arquivo.Name()
		if arquivo.IsDir() {
			err = ClearPDF(caminhoArquivo)
			if err != nil {
				return err
			}

			err = os.Remove(caminhoArquivo)
			if err != nil {
				return err
			}
		} else {

			err = os.Remove(caminhoArquivo)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
