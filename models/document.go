package models

type Document struct {
	Documento     string
	NF            string
	Emissao       string
	Vencimento    string
	Valor         float64
	Juros         float64
	Multas        float64
	Deducoes      float64
	Total         float64
	LinhaDigitavel string
	Boleto        string
}
