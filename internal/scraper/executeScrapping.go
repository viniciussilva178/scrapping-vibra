package scraper

import (
	"fmt"
	"log"
	"scraper/db"
	scraper "scraper/internal/scraper/proccecing"
	"strconv"
	"time"
)

func ExecuteScrapper() {
	credentials := db.LoadCredentials()
	if len(credentials) == 0 {
		fmt.Println("Nenhuma credencial encontrada no .env")
		return
	}

	operation := db.NewOperation()

	inicio := time.Now()
	for i, cred := range credentials {
		fmt.Printf("\n=== PROCESSANDO USUÁRIO %d: %s ===\n", i+1, cred.User)

		contasPagar, page, browser, err := ScrapeDocsVibra(cred.User, cred.Password, &operation)
		if err != nil {

			if page != nil {
				page.MustClose()
			}
			if browser != nil {
				browser.MustClose()
			}
			continue
		}
		defer page.Close()

		for j, conta := range contasPagar {
			_, err = operation.CreateOperationVibraContasAPagar(&conta)
			if err != nil {
				fmt.Printf("Erro ao salvar conta %d para usuário %s: %v\n", j+1, cred.User, err)
			} else {
				fmt.Printf("Conta %d salva com sucesso! Documento: %s\n", j+1, conta.NumeroDocumento)
			}
		}

		for idx, conta := range contasPagar {
			increment := strconv.Itoa(idx + 1)
			idNF, numeroFatura, err := scraper.ProcessarNFe(page, increment, conta.NumeroFatura, &operation)
			if err != nil {
				log.Printf("Erro ao processar NF linha %d: %v", idx+1, err)
				continue
			}

			if _, _, err := operation.UpdateOperationVibraWithNF(numeroFatura, idNF); err != nil {
				log.Printf("Erro ao atualizar conta com NF linha %d (numeroFatura: %s, idNF: %d): %v", idx+1, numeroFatura, idNF, err)
			} else {
				log.Printf("Conta linha %d atualizada com NF %s (id %d)", idx+1, numeroFatura, idNF)
			}
		}
		// defer os.Remove("*.pdf")
		fim := time.Now()
		duracao := fim.Sub(inicio)

		fmt.Printf("\n=== PROCESSAMENTO DE TODOS OS USUÁRIOS CONCLUÍDO em: %v ===", duracao)
		fmt.Printf("Usuário %s processado com %d contas\n", cred.User, len(contasPagar))
	}
}
