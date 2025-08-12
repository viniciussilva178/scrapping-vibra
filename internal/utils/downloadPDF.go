package utils

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

func DownloadPDFToFile(page *rod.Page, pdfURL, filePath, referer string) error {
	pdfURL = strings.TrimSpace(pdfURL)
	if pdfURL == "" {
		return fmt.Errorf("pdfURL vazio")
	}
	if strings.HasPrefix(pdfURL, "/") {
		pdfURL = "https://cn.vibraenergia.com.br" + pdfURL
	}

	if strings.HasPrefix(pdfURL, "blob:") {
		return fmt.Errorf("pdfURL é um blob URL; é necessário baixar via browser (JS) no contexto da página")
	}

	cookies := page.MustCookies()

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	req, err := http.NewRequest("GET", pdfURL, nil)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %v", err)
	}

	for _, c := range cookies {
		if c.Domain == "" || strings.Contains(c.Domain, "vibraenergia") {
			req.AddCookie(&http.Cookie{
				Name:  c.Name,
				Value: c.Value,
				Path:  c.Path,
			})
		}
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116 Safari/537.36")
	req.Header.Set("Accept", "application/pdf,application/octet-stream,*/*")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", referer)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro na requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("DownloadPDFToFile: status=%s content-type=%s\n", resp.Status, resp.Header.Get("Content-Type"))

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo PDF: %v", err)
	}
	defer out.Close()

	br := bufio.NewReader(resp.Body)
	peek, _ := br.Peek(4)

	_, err = io.Copy(out, br)
	if err != nil {
		return fmt.Errorf("erro ao salvar PDF no disco: %v", err)
	}

	// Verifica header mágico
	if len(peek) < 4 || string(peek) != "%PDF" {
		debugPath := filePath + ".debug.html"
		_ = os.Rename(filePath, debugPath)
		return fmt.Errorf("conteúdo retornado não é um PDF (status=%s content-type=%s). salvo: %s",
			resp.Status, resp.Header.Get("Content-Type"), debugPath)
	}

	fmt.Printf("PDF salvo com sucesso: %s\n", filePath)
	return nil
}
