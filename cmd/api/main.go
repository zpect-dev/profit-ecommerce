package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"profit-ecommerce/internal/api"
	"profit-ecommerce/internal/api/handlers"
	"profit-ecommerce/internal/auth"
	"profit-ecommerce/internal/cart"
	"profit-ecommerce/internal/catalog"
	"profit-ecommerce/internal/config"
	"profit-ecommerce/pkg/database"

	"github.com/redis/go-redis/v9"
)

// mockCatalogService provee una implementación básica para cumplir la interfaz requerida por el Cart.
type mockCatalogService struct{}

func (m *mockCatalogService) CheckStock(ctx context.Context, productIDs []string) (map[string]int, error) {
	// Retorna un dummy (asumimos que hay suficiente stock "99" para cualquier item)
	res := make(map[string]int)
	for _, id := range productIDs {
		res[id] = 99
	}
	return res, nil
}

func main() {
	cfg := config.Load()

	// 1. CONEXIONES
	dbConn, err := database.ConnectPostgres(cfg.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	// Instanciamos el cliente de Redis usando la configuración distribuida
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Error conectando a Redis: %v", err)
	}
	log.Println("Conectado a Redis")

	// 2. INYECCIÓN DE DEPENDENCIAS (wiring)
	// -- Dominio Catálogo
	catRepo := catalog.NewRepository(dbConn)
	catHandler := handlers.NewCatalogHandler(catRepo)

	// -- Dominio Carrito
	cartCacheRepo := cart.NewRedisCartRepository(redisClient)
	cartDBRepo := cart.NewSQLCartRepository(dbConn)
	mockCatalog := &mockCatalogService{}

	// Orquestador e Inyección HTTP
	cartSvc := cart.NewService(cartCacheRepo, cartDBRepo, mockCatalog)
	cartHandler := cart.NewCartHandler(cartSvc)

	// -- Dominio Autenticación
	authCacheRepo := auth.NewRedisAuthRepository(redisClient)
	clientRepo := auth.NewClientRepository(dbConn)
	authSvc := auth.NewAuthService(authCacheRepo, clientRepo)
	authHandler := auth.NewAuthHandler(authSvc)

	// 3. ROUTER
	router := api.NewRouter(catHandler, cartHandler, authHandler, authSvc)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// 4. GRACEFUL SHUTDOWN STRICTO
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Arrancamos el servidor en una goroutine concurrente
	go func() {
		fmt.Printf("\nAPI Gateway listo en http://localhost:%s\n", cfg.Port)
		fmt.Printf("Prueba: http://localhost:%s/v1/products\n", cfg.Port)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error crítico en el servidor HTTP: %v", err)
		}
	}()

	// Bloqueamos main hasta recibir la señal (Ctrl+C o SIGTERM de Docker/K8s)
	<-ctx.Done()
	log.Println("\nIniciando apagado Graceful del servidor...")

	// Damos un margen a peticiones HTTP en vuelo
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Fallo al apagar servidor HTTP gracefulmente: %v", err)
	}

	// ¡MUY IMPORTANTE! Drenado del buffer del patrón Write-Behind antes de terminar main.
	log.Println("Deteniendo orquestador de Carrito y drenando eventos de base de datos pendientes...")
	if err := cartSvc.Close(); err != nil {
		log.Printf("Error al cerrar el Worker de Carritos: %v", err)
	}

	log.Println("Procesos finalizados exitosamente. Bye!")
}
