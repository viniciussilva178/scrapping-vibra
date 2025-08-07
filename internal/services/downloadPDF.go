package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-rod/rod"
)

func DownloadPDF(page *rod.Page, pdfURL string, filename string) error {
	cookies := page.MustCookies()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", pdfURL, nil)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %v", err)
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
		return fmt.Errorf("erro na requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro HTTP: %s", resp.Status)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo: %v", err)
	}
	defer file.Close()

	bytesWritten, err := io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("erro ao salvar arquivo: %v", err)
	}

	fmt.Printf("PDF salvo com sucesso: %s (%d bytes)\n", filename, bytesWritten)
	return nil
}
