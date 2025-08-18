package pkg

import (
	"fmt"
	"os"
	"regexp"

	"github.com/ledongthuc/pdf"
)

// GetNumeroFatura tenta extrair o número da fatura (ex: 002287009) do PDF.
// Estratégia:
// 1) busca direto nos bytes do PDF com várias regex tolerantes;
// 2) se não achar, abre o PDF e tenta extrair texto pela lib e reaplica regex;
// 3) por fim, tenta achar qualquer sequência longa de dígitos.
func GetNumeroFatura(caminho string, pageNum int) (string, error) {
	// 1) ler bytes do arquivo (pesquisa direta no binário é preferível)
	data, err := os.ReadFile(caminho)
	if err != nil {
		return "", fmt.Errorf("erro ao ler PDF: %w", err)
	}

	patterns := []string{
		`(?i)N°\*[:]?\s*([0-9]{5,25})`, // Nº: 002.287.009
	}

	nonDigitRe := regexp.MustCompile(`\D`)

	// função auxiliar para validar e limpar: remove não-dígitos e valida tamanho
	cleanAndValidate := func(raw string) (string, bool) {
		digits := nonDigitRe.ReplaceAllString(raw, "")
		if len(digits) >= 5 { // ajuste mínimo aceitável (>=5 dígitos)
			return digits, true
		}
		return "", false
	}

	// procurar nas bytes
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		if m := re.FindSubmatch(data); len(m) >= 2 {
			if cleaned, ok := cleanAndValidate(string(m[1])); ok {
				return cleaned, nil
			}
		}
	}

	// 2) fallback: tentar extrair texto com a lib (só agora abrimos com o parser)
	f, r, err := pdf.Open(caminho)
	if err == nil {
		// garantir fechamento do arquivo
		defer f.Close()

		// checar páginas
		if pageNum >= 1 && pageNum <= r.NumPage() {
			// tenta extrair texto (GetPlainText pode falhar em alguns PDFs)
			if txt, errTxt := r.Page(pageNum).GetPlainText(nil); errTxt == nil {
				// aplicar regexs sobre o texto extraído
				for _, p := range patterns {
					re := regexp.MustCompile(p)
					if m := re.FindStringSubmatch(txt); len(m) >= 2 {
						if cleaned, ok := cleanAndValidate(m[1]); ok {
							return cleaned, nil
						}
					}
				}
			} else {
				// se GetPlainText falhar, podemos tentar um método alternativo (algumas versões
				// da lib expõem GetText). Se não existe, só logamos o erro e seguimos.
				_ = errTxt // mantemos para possível log/depuração
			}
		}
	}

	// 3) último recurso: encontrar qualquer sequência longa de dígitos nos bytes
	longRe := regexp.MustCompile(`[0-9]{6,}`)
	if m := longRe.Find(data); m != nil {
		// remover zeros à esquerda? aqui retorno o que conseguiu
		fmt.Println(string(m))
		return string(m), nil
	}

	return "", fmt.Errorf("nenhum número de fatura encontrado no PDF %s (page %d)", caminho, pageNum)
}
