package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"profit-ecommerce/internal/config"
	"profit-ecommerce/internal/syncer"
	"profit-ecommerce/pkg/database"

	"github.com/jmoiron/sqlx"
)

func main() {
	// 1. CONTEXTO CON SIGNAL HANDLING (SIGINT / SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	// 2. CONEXIONES
	profitDB, err := database.ConnectSQLServer(cfg.ProfitDBURL)
	if err != nil {
		log.Fatalf("Error conectando a Profit: %v", err)
	}
	defer profitDB.Close()

	pgDB, err := database.ConnectPostgres(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Error conectando a Postgres: %v", err)
	}
	defer pgDB.Close()

	runMigrations(pgDB)

	// 3. INYECCIÓN DE DEPENDENCIAS
	sourceRepo := syncer.NewSourceRepository(profitDB)
	destRepo := syncer.NewDestRepository(pgDB)
	syncService := syncer.NewService(sourceRepo, destRepo)

	fmt.Println("Worker Iniciado. Esperando ciclos... (Ctrl+C para shutdown limpio)")

	// 4. DEFINIR LOS RELOJES
	fastTicker := time.NewTicker(1 * time.Minute)
	defer fastTicker.Stop()

	slowTicker := time.NewTicker(1 * time.Hour)
	defer slowTicker.Stop()

	// 5. EJECUTAR TODO UNA VEZ AL ARRANCAR (con contexto)
	go func() {
		fmt.Println("Ejecución inicial de arranque...")
		syncService.RunSlowSync(ctx)
		syncService.RunFastSync(ctx)
	}()

	// 6. BUCLE PRINCIPAL CON SHUTDOWN GRACEFUL
	for {
		select {
		case <-ctx.Done():
			// SIGINT o SIGTERM recibido: salimos limpiamente
			fmt.Println("\n[SHUTDOWN] Señal recibida. Deteniendo worker...")
			fmt.Println("[SHUTDOWN] Esperando que los syncs en curso terminen...")
			fmt.Println("[SHUTDOWN] Worker detenido limpiamente.")
			return

		case <-fastTicker.C:
			fmt.Println("\n[TICKER] Iniciando Sync Rápido (Stock/Precios)...")
			syncService.RunFastSync(ctx)

		case <-slowTicker.C:
			fmt.Println("\n[TICKER] Iniciando Sync Lento (Maestros)...")
			syncService.RunSlowSync(ctx)
		}
	}
}

func runMigrations(db *sqlx.DB) {
	migrationDir := "db/migrations"
	fmt.Printf("Buscando migraciones en: %s\n", migrationDir)

	files, err := os.ReadDir(migrationDir)
	if err != nil {
		log.Fatal("Error leyendo carpeta de migraciones: ", err)
	}

	var upMigrations []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			upMigrations = append(upMigrations, f.Name())
		}
	}

	sort.Strings(upMigrations)

	for _, filename := range upMigrations {
		fmt.Printf("Ejecutando: %s... ", filename)

		fullPath := filepath.Join(migrationDir, filename)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			log.Fatalf("\nError leyendo %s: %v", filename, err)
		}

		_, err = db.Exec(string(content))
		if err != nil {
			log.Fatalf("\nError ejecutando %s: %v", filename, err)
		}
		fmt.Println("OK")
	}

	fmt.Println("Todas las migraciones aplicadas.")
}
