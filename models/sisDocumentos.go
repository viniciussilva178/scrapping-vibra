package models

type SisDocumentos struct {
	IdDebitoDireto int    `json:"id_debito_direto_autorizado"`
	IdEmpresa      int    `json:"id_empresa"`
	NomeDocumento  string `json:"nome_documento"`
	Tamanho        int    `json:"tamanho"`
	Id_usuario     int    `json:"id_usuario_inclusao"`
	Arquivo        []byte `json:"arquivo"`
	Tipo           int8   `json:"id_tipo_documento"`
}
