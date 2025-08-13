package db

import (
	"database/sql"
	"fmt"
	"scraper/models"
	"time"
)

type Operation struct {
	connection *sql.DB
}

func connectionValidation() (*sql.DB, error) {
	conn, err := DBConnect()
	if err != nil {
		return nil, err
	}

	return conn, err
}

func NewOperation() Operation {
	conn, err := connectionValidation()
	if err != nil {
		fmt.Println("Erro na conexaão com o Banco de dados", err)
	}

	if conn == nil {
		fmt.Println("AVISO: Não foi possível conectar ao banco de dados")
	}

	return Operation{
		connection: conn,
	}
}

func (o *Operation) IsConnected() bool {
	return o.connection != nil
}

func (o *Operation) CreateOperationVibraContasAPagar(contaPagar *models.ContasAPagar) (string, error) {
	dataEmissaoFormatada, err := time.Parse("02/01/2006", contaPagar.DataEmissao)
	if err != nil {
		fmt.Println("Erro ao realizar a formatação da data de emisão: ", err)
	}

	dataVencimentoFormatada, err := time.Parse("02/01/2006", contaPagar.DataVencimento)
	if err != nil {
		fmt.Println("Erro ao realizar a formatação da data de vencimento: ", err)
	}

	// Verificar se a conexão está disponível
	if o.connection == nil {
		return "", fmt.Errorf("conexão com o banco de dados não está disponível")
	}

	sucessful := "Contas A Pagar Armazaneda com sucesso no Banco de dados!"

	query, err := o.connection.Prepare(
		"INSERT INTO vibra.conta_pagar" +
			"(numero_documento,numero_fatura,data_emissao,data_vencimento,valor_documento,valor_juros,valor_multa," +
			"valor_deducao, valor_total, linha_digitavel, cnpj_beneficiario, cnpj_pagador, numero_serie, id_boleto, id_nota_fiscal, id_xml) " +
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)")
	if err != nil {
		fmt.Println("Erro ao preparar conexão com o banco de dados", err)
		return "", err
	}
	defer query.Close()

	_, err = query.Exec(
		contaPagar.NumeroDocumento,
		contaPagar.NumeroFatura,
		dataEmissaoFormatada,
		dataVencimentoFormatada,
		contaPagar.ValorDocumento,
		contaPagar.ValorJuros,
		contaPagar.ValorMulta,
		contaPagar.ValorDeducao,
		contaPagar.ValorTotal,
		contaPagar.LinhaDigitavel,
		contaPagar.CNPJBeneficiario,
		contaPagar.CNPJPagador,
		contaPagar.NumeroSerie,
		contaPagar.IDBoleto,
		contaPagar.IDNotaFiscal,
		contaPagar.IDXML,
	)
	if err != nil {
		fmt.Println("Erro ao Executar a query", err)
		return "", err

	}

	return sucessful, nil
}

func (o *Operation) CreateOperationVibraDocumento(documento *models.Document, tipo int) (int, error) {
	if o.connection == nil {
		return 0, fmt.Errorf("conexão com o banco de dados não está disponível")
	}

	var id int
	query, err := o.connection.Prepare(`
        INSERT INTO vibra.documento_abastecimento (tipo_documento, conteudo_documento)
        VALUES ($1, $2)
        RETURNING id_documento_abastecimento`,
	)
	if err != nil {
		fmt.Println("Erro ao preparar a requisicao")
	}

	err = query.QueryRow(tipo, documento.ConteudoDocumento).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}
