package db

import (
	"database/sql"
	"fmt"
	"scraper/models"
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

func (o *Operation) CreateOperationVibra(document *models.Document) (string, error) {

	// Verificar se a conexão está disponível
	if o.connection == nil {
		return "", fmt.Errorf("conexão com o banco de dados não está disponível")
	}

	sucessful := "Documento Armazanedo com sucesso no Banco de dados!"

	query, err := o.connection.Prepare("INSERT INTO vibra.documento_abastecimento" +
		"(codigo_documento,nf_fatura,data_emissao,data_vencimento,valor_documento,valor_juros,valor_multa,valor_deducao,valor_total,linha_digitavel)" +
		"VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)")
	if err != nil {
		fmt.Println("Erro ao preparar conexão com o banco de dados", err)
		return "", err
	}
	defer query.Close()

	_, err = query.Exec(
		document.Documento,
		document.NF,
		document.Emissao,
		document.Vencimento,
		document.Valor,
		document.Juros,
		document.Multas,
		document.Deducoes,
		document.Total,
		document.LinhaDigitavel,
	)
	if err != nil {
		fmt.Println("Erro ao Executar a query", err)
		return "", err

	}

	return sucessful, nil
}
