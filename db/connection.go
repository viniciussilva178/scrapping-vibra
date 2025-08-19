package db

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Credentials struct {
	User     string
	Password string
}

func Connection() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Erro ao carregar .env:", err)
	}

	fmt.Println("Testando conexão com banco de dados...")
	operation := NewOperation()
	if !operation.IsConnected() {
		fmt.Println("ERRO: Não foi possível conectar ao banco de dados")
		fmt.Println("Verifique as configurações de conexão em internal/db/db.go")
		return
	}

	fmt.Println("Conexão com banco de dados estabelecida com sucesso!")

}

func LoadCredentials() []Credentials {
	var creds []Credentials
	for i := 1; ; i++ {
		userKey := fmt.Sprintf("USER%d", i)
		passKey := fmt.Sprintf("PASSWORD%d", i)
		user := os.Getenv(userKey)
		pass := os.Getenv(passKey)

		if user == "" || pass == "" {
			break
		}

		creds = append(creds, Credentials{
			User:     user,
			Password: pass,
		})
		fmt.Printf("Credencial %d carregada: %s\n", i, user)
	}
	return creds
}
