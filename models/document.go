package models

type Document struct {
	ID                string  `json:"id_documento_abastecimento"`
	Tipo              int     `json:"tipo_documento"`      // Boleto=1, NF=2, XML=3
	ConteudoDocumento []byte  `json:"conteudo_documento "` // Todos os documentos aqui
	NumeroNFE         *string `json:"numero_nfe"`
}
