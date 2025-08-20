package db

import (
	"fmt"
	"scraper/models"
)

func NewOperationSis() Operation {
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

func (o *Operation) GetDebitoAutorizado(linha_digitavel string) (int, error) {
	var id int

	query, err := o.connection.Prepare(`SELECT id_debito_direto_autorizado
				FROM public.debito_direto_autorizado 
				WHERE linha_digitavel = $1`)

	if err != nil {
		fmt.Println("Erro ao pegar linha digitavel na tabela debito_direto_autorizado")
	}

	err = query.QueryRow(linha_digitavel).Scan(&id)
	if err != nil {
		return 0, err
	}
	defer query.Close()

	fmt.Printf("o id do Debito Autorizado  é: %v ", id)

	return id, nil

}

func (o *Operation) GetEmpresa() (int, error) {
	var empresaId int

	query, err := o.connection.Prepare(`SELECT se.id_empresa
    	FROM pessoa p 
        INNER JOIN vibra.conta_pagar cp on (cp.cnpj_pagador = p.cnpj_cpf)
        INNER JOIN public.sis_empresa se on (se.id_pessoa = p.id_pessoa)`)

	if err != nil {
		fmt.Println("Erro ao pegar linha digitavel na tabela debito_direto_autorizado")
	}

	err = query.QueryRow().Scan(&empresaId)
	if err != nil {
		return 0, err
	}
	defer query.Close()

	fmt.Printf("o id da empresa  é: %vs ", empresaId)

	return int(empresaId), nil

}

func (o *Operation) GetNomeDocumento(linhaDigitavel string) (string, error) {
	var nomeDocumento string
	FormatadoDocumentoNome := nomeDocumento + ".pdf"

	query, err := o.connection.Prepare(`
        select numero_fatura
		from  vibra.conta_pagar p
		where linha_digitavel=$1
    `)
	if err != nil {
		return "", err
	}
	defer query.Close()

	err = query.QueryRow(linhaDigitavel).Scan(&nomeDocumento)
	if err != nil {
		return "", err
	}

	fmt.Printf("O Nome do documento é: %s", FormatadoDocumentoNome)

	return FormatadoDocumentoNome, nil
}

func (o *Operation) GetConteudoDocumento() ([]byte, int, int8, error) {
	var conteudoDocumento []byte
	var lengthConteudo int
	var tipoDocumento int8

	query, err := o.connection.Prepare(`
       SELECT conteudo_documento, tipo_documento
    FROM vibra.documento_abastecimento p 
        INNER JOIN vibra.conta_pagar cp on (cp.id_nota_fiscal = p.id_documento_abastecimento)
        INNER JOIN public.debito_direto_autorizado d on (d.linha_digitavel = cp.linha_digitavel)
    `)
	if err != nil {
		return nil, 0, 0, err
	}
	defer query.Close()

	err = query.QueryRow().Scan(&conteudoDocumento, &tipoDocumento)
	if err != nil {
		return nil, 0, 0, err
	}

	lengthConteudo = len(conteudoDocumento)

	fmt.Printf("O Nome do documento é: %s", conteudoDocumento)
	fmt.Printf("E o tamanho é: %v", lengthConteudo)
	fmt.Printf("O Tipo de documento é: %v", tipoDocumento)

	return conteudoDocumento, lengthConteudo, tipoDocumento, nil
}

func (o *Operation) CreateSisDocumento(sisDocumento models.SisDocumentos) (int, error) {
	var id int

	query, err := o.connection.Prepare(`
		INSERT INTO public.sis_documento
    	(id_empresa, nome_documento, tamanho, id_usuario_inclusao, id_tipo_documento) 
    	VALUES ($1, $2, $3, $4, $5) RETURNING id_documento;
	`)
	if err != nil {
		fmt.Println("Erro ao Preparar conexão de criar um sisDocumento", err)
		fmt.Println(sisDocumento)
		return 0, err
	}

	err = query.QueryRow(
		sisDocumento.IdEmpresa,
		sisDocumento.NomeDocumento,
		sisDocumento.Tamanho,
		sisDocumento.Id_usuario,
		sisDocumento.Tipo).Scan(&id)
	if err != nil {
		fmt.Println("erro executar Criação de im sisDocumento", err)
		return 0, err
	}

	return id, nil
}

func (o *Operation) CreateSisrelaçãoRegistro(sisRelacaoRegistro models.SisRelacaoRegistro) (string, error) {

	sucessful := "Sis registro Criado com sucesso"

	query, err := o.connection.Prepare(`
		INSERT into public.sis_relacao_registro 
		(id_chave_registro_a, id_chave_registro_b, tipo_relacao, id_usuario_inclusao) 
		VALUES ($1, $2, $3, $4)  
	`)
	if err != nil {
		fmt.Println("Erro ao Preparar conexão de criar um sisDocumento", err)
		return "", err
	}

	_, err = query.Exec(
		sisRelacaoRegistro.IdChaveregistroA,
		sisRelacaoRegistro.IdChaveRegistroB,
		sisRelacaoRegistro.TipoRelação,
		sisRelacaoRegistro.IdUsuarioInclusao,
	)
	if err != nil {
		fmt.Println("erro executar Criação de im sisRelacaoRegistro", err)
		return "", err
	}

	return sucessful, nil
}
