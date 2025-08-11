package models

type Document struct {
	Documento      string  `json:"codigo_documento"`
	NF             string  `json:"nf_fatura"`
	Emissao        string  `json:"data_emissao"`
	Vencimento     string  `json:"data_vencimento"`
	Valor          float64 `json:"valor_documento"`
	Juros          float64 `json:"valor_juros"`
	Multas         float64 `json:"valor_multa"`
	Deducoes       float64 `json:"valor_deducao"`
	Total          float64 `json:"valor_total"`
	LinhaDigitavel string  `json:"linha_digitavel"`
	BoletoURL      string  `json:"boleto_url"`
	Conteudo       []byte  `json:"conteudo"` // Campo para armazenar os dados do PDF
}
