package database

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
)

func ConnectPostgres(url string) (*sqlx.DB, error) {
	if url == "" {
		return nil, fmt.Errorf("URL de Postgres vacía")
	}
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("error abriendo conexión a Postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error conectando a Postgres: %w", err)
	}

	log.Println("Conectado a Postgres")
	return db, nil
}