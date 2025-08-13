package models

type ContasAPagar struct {
	NumeroDocumento  string  `json:"numero_documento"`
	NumeroFatura     string  `json:"numero_fatura"`
	DataEmissao      string  `json:"data_emissao"`
	DataVencimento   string  `json:"data_vencimento"`
	ValorDocumento   float64 `json:"valor_documento"`
	ValorJuros       float64 `json:"valor_juros"`
	ValorMulta       float64 `json:"valor_multa"`
	ValorDeducao     float64 `json:"valor_deducao"`
	ValorTotal       float64 `json:"valor_total"`
	LinhaDigitavel   string  `json:"linha_digitavel"`
	CNPJBeneficiario string  `json:"cnpj_beneficiario"`
	CNPJPagador      string  `json:"cnpj_pagador"`
	NumeroSerie      string  `json:"numero_serie"`
	IDBoleto         *int    `json:"id_boleto"`
	IDNotaFiscal     *int    `json:"id_nota_fiscal"`
	IDXML            *int    `json:"id_xml"`
}
