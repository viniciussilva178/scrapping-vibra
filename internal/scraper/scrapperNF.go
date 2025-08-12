package scraper

import (
	"scraper/internal/utils"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

func ScrapperNF(page *rod.Page, incremental string) {
	newDate := time.Now().AddDate(0, 0, -1).Format("02/01/2006")

	page.MustElement("#btnCloseModalImprimir").MustClick()
	time.Sleep(3 * time.Second)

	page.MustElement(`#menuAcessoRevendedoNFe`).MustClick()
	page.MustWaitStable().MustElement("#downloadNotaFiscalForm").MustWaitVisible()

	page.MustEval(`date => {
		const el = document.querySelector('#dataEmissaoInicial');
		el.value = date;
		el.dispatchEvent(new Event('input', { bubbles: true }));
		el.dispatchEvent(new Event('change', { bubbles: true }));
	}`, newDate)

	page.MustElement(`#btListar`).MustClick()
	page.MustWaitStable().MustElement(`#panelNotaFiscal`).MustWaitVisible()

	pdfURL := string("https://cn.vibraenergia.com.br/cn/comercio/notafiscaleletronicanova/downloadNotaFiscal?tipoArquivo=pdf&idLinha=" + incremental + "&tipoDocumento=danfe")
	if strings.HasPrefix(pdfURL, "/") {
		pdfURL = "https://cn.vibraenergia.com.br/" + pdfURL
	}

	err := utils.DownloadPDFToFile(page, pdfURL, "nfe.pdf", "https://cn.vibraenergia.com.br/cn/comercio/notafiscaleletronicanova/")
	if err != nil {
		panic(err)
	}

	time.Sleep(3 * time.Second)

}
