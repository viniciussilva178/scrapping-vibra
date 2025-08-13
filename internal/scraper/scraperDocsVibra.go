package scraper

import (
	"fmt"
	"scraper/db"

	scraper "scraper/internal/scraper/proccecing"
	util "scraper/internal/utils"
	"scraper/models"
	"scraper/pkg"

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
	contasPagar, err := scraper.ProcessarContas(doc, authenticatedPage, op)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar contas: %v", err)
	}

	fmt.Printf("--Processamento concluído para %s (%d contas)\n", user, len(contasPagar))
	fmt.Printf("--Tempo total: %s\n", time.Since(inicio))

	return contasPagar, nil
}
