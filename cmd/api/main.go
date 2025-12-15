package main

import (
	"fmt"
	"log"
	"net/http"
	"profit-ecommerce/internal/api"
	"profit-ecommerce/internal/config"
	"profit-ecommerce/internal/db"
)

func main() {
	cfg := config.Load()

	dbConn, err := db.ConnectPostgres(cfg.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	router := api.NewRouter(dbConn)

	fmt.Printf("API Gateway listo en http://localhost:%s\n", cfg.Port)
	fmt.Printf("Prueba: http://localhost:%s/v1/products\n", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
