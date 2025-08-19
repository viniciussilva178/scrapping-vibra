package models

type SisRelacaoRegistro struct {
	IdChaveregistroA  int    `json:"id_chave_registro_a"`
	IdChaveRegistroB  int    `json:"id_chave_registro_b"`
	TipoRelação       string `json:"tipo_relacao"`
	IdUsuarioInclusao int8   `json:"id_usuario_inclusao"`
}
