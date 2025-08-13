package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func DBConnect() (*sql.DB, error) {

	host := os.Getenv("HOST_DB")
	port := os.Getenv("PORT_DB")
	user := os.Getenv("USER_DB")
	password := os.Getenv("PASSWORD_DB")
	dbName := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open the database: %v", err)
	}

	_, err = db.Exec("SET search_path TO public;")
	if err != nil {
		log.Fatalf("Error setting search_path: %v", err)
	}

	fmt.Println("Connected to database:", dbName)

	return db, err

}
