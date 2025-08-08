package services

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

func DownloadPDF(page *rod.Page, pdfURL string) ([]byte, error) {
	cookies := page.MustCookies()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", pdfURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	for _, cookie := range cookies {
		req.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/pdf,application/octet-stream,*/*")
	req.Header.Set("Referer", "https://cn.vibraenergia.com.br/")

	fmt.Printf("Fazendo download do PDF: %s\n", pdfURL)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro HTTP: %s", resp.Status)
	}

	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "application/pdf") {
		return nil, fmt.Errorf("conteúdo retornado não é um PDF: %s", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o conteúdo do PDF: %v", err)
	}

	if len(body) < 4 || string(body[:4]) != "%PDF" {
		return nil, fmt.Errorf("conteúdo baixado não é um PDF válido")
	}

	fmt.Printf("PDF lido com sucesso: %d bytes\n", len(body))
	return body, nil
}
