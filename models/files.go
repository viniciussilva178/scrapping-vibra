package models

type File struct {
	ID      string `json:"id"`
	Tipo    int    `json:"tipo"`
	Arquivo []byte `json:"arquivo"` // Todos os documentos aqui
}

//
