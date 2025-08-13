package pkg

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

func DownloadXMLToFile(page *rod.Page, xmlURL, filePath, referer string) error {
	xmlURL = strings.TrimSpace(xmlURL)
	if xmlURL == "" {
		return fmt.Errorf("xmlURL vazio")
	}
	if strings.HasPrefix(xmlURL, "/") {
		xmlURL = "https://cn.vibraenergia.com.br" + xmlURL
	}

	if strings.HasPrefix(xmlURL, "blob:") {
		return fmt.Errorf("xmlURL é um blob URL; é necessário baixar via browser (JS) no contexto da página")
	}

	cookies := page.MustCookies()

	client := &http.Client{Timeout: 120 * time.Second}

	req, err := http.NewRequest("GET", xmlURL, nil)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %v", err)
	}

	for _, c := range cookies {
		if c.Domain == "" || strings.Contains(c.Domain, "vibraenergia") {
			req.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value, Path: c.Path})
		}
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116 Safari/537.36")
	req.Header.Set("Accept", "application/xml,text/xml,application/octet-stream,*/*")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", referer)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro na requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("DownloadXMLToFile: status=%s content-type=%s\n", resp.Status, resp.Header.Get("Content-Type"))

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo XML: %v", err)
	}
	defer out.Close()

	br := bufio.NewReader(resp.Body)
	peek, _ := br.Peek(64) // pega um pouco mais para validar

	// Se começar com <!DOCTYPE html ou <html>, provavelmente é erro
	if strings.Contains(strings.ToLower(string(peek)), "<html") {
		debugPath := filePath + ".debug.html"
		io.Copy(out, br)
		_ = os.Rename(filePath, debugPath)
		return fmt.Errorf("conteúdo retornado parece HTML, não XML. salvo: %s", debugPath)
	}

	_, err = io.Copy(out, br)
	if err != nil {
		return fmt.Errorf("erro ao salvar XML no disco: %v", err)
	}

	fmt.Printf("XML salvo com sucesso: %s\n", filePath)
	return nil
}
