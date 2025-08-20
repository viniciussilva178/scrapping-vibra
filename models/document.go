package models

type Document struct {
	ID                string  `json:"id_documento_abastecimento"`
	Tipo              int     `json:"tipo_documento"` // Boleto=2, NF=3, XML=1
	ConteudoDocumento []byte  `json:"conteudo_documento "`
	NumeroNFE         *string `json:"numero_nfe"`
}
