package main

import (
	"fmt"
	"log"
	"net/http"
	"profit-ecommerce/internal/api"
	"profit-ecommerce/internal/api/handlers"
	"profit-ecommerce/internal/catalog"
	"profit-ecommerce/internal/config"
	"profit-ecommerce/pkg/database"
)

func main() {
	cfg := config.Load()

	// 1. CONEXIÓN
	dbConn, err := database.ConnectPostgres(cfg.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	// 2. INYECCIÓN DE DEPENDENCIAS (wiring)
	catRepo := catalog.NewRepository(dbConn)
	catHandler := handlers.NewCatalogHandler(catRepo)

	// 3. ROUTER (solo recibe handlers, no sabe de BD)
	router := api.NewRouter(catHandler)

	fmt.Printf("API Gateway listo en http://localhost:%s\n", cfg.Port)
	fmt.Printf("Prueba: http://localhost:%s/v1/products\n", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
