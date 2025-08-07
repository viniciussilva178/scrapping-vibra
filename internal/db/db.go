package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func DBConnect() (*sql.DB, error) {
	host := "brapoio-dev.internal"
	port := "5432"
	user := "brapoio_vinicius"
	password := "ckt9tqu.rnb7yxu3XQB"
	dbName := "brapoio_db"

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
