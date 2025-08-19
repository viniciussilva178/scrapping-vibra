package db

import (
	"database/sql"
	"fmt"
	"log"
	"scraper/models"
	"time"
)

type Operation struct {
	connection *sql.DB
}

func ConnectionValidation() (*sql.DB, error) {
	conn, err := DBConnect()
	if err != nil {
		return nil, err
	}

	return conn, err
}

func NewOperation() Operation {
	conn, err := ConnectionValidation()
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
	// Verificar se NumeroNFE já existe
	if documento.NumeroNFE != nil {
		err := o.connection.QueryRow(`
			SELECT id_documento_abastecimento
			FROM vibra.documento_abastecimento
			WHERE numero_nfe = $1 AND tipo_documento = $2`, *documento.NumeroNFE, tipo).Scan(&id)
		if err == nil {
			log.Printf("Documento existente encontrado para NumeroNFE %s: id %d", *documento.NumeroNFE, id)
			return id, nil
		} else if err != sql.ErrNoRows {
			return 0, fmt.Errorf("erro ao verificar documento existente para NumeroNFE %s: %v", *documento.NumeroNFE, err)
		}
	}

	// Inserir novo documento
	query, err := o.connection.Prepare(`
		INSERT INTO vibra.documento_abastecimento (tipo_documento, conteudo_documento, numero_nfe)
		VALUES ($1, $2, $3)
		RETURNING id_documento_abastecimento`)
	if err != nil {
		return 0, fmt.Errorf("erro ao preparar a query de inserção: %v", err)
	}
	defer query.Close()

	err = query.QueryRow(tipo, documento.ConteudoDocumento, documento.NumeroNFE).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("erro ao inserir documento para NumeroNFE %v: %v", documento.NumeroNFE, err)
	}

	log.Printf("Novo documento inserido para NumeroNFE %v: id %d", documento.NumeroNFE, id)
	return id, nil
}

func (o *Operation) UpdateOperationVibraWithNF(numero_fatura string, idNF int) (string, int, error) {
	if o.connection == nil {
		return "", 0, fmt.Errorf("conexão com o banco de dados não está disponível")
	}

	var id int
	query, err := o.connection.Prepare(`
        UPDATE vibra.conta_pagar
		SET id_nota_fiscal = $1
		WHERE numero_fatura = $2
		RETURNING id_conta_pagar;
	`)
	if err != nil {
		return "", 0, fmt.Errorf("erro ao preparar a query de atualização: %v", err)
	}
	defer query.Close()

	err = query.QueryRow(idNF, numero_fatura).Scan(&id)
	if err == sql.ErrNoRows {
		return "", 0, fmt.Errorf("nenhuma conta encontrada para numero_fatura: %s", numero_fatura)
	} else if err != nil {
		return "", 0, fmt.Errorf("erro ao executar a query de atualização para numero_fatura %s: %v", numero_fatura, err)
	}

	return numero_fatura, id, nil
}
