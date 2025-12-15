package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresURL string
	ProfitDBURL string
	Port        string
}

func Load() *Config {
	// Intentar cargar .env, pero no fallar si no existe (producción)
	err := godotenv.Load()
	if err != nil {
		cwd, _ := os.Getwd()
		log.Printf("Info: No se pudo cargar el archivo .env: %v (CWD: %s)", err, cwd)
	}

	return &Config{
		PostgresURL: getEnv("POSTGRES_URL", ""),
		ProfitDBURL: getEnv("PROFIT_DB_URL", ""),
		Port:        getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	if fallback == "" {
		// Opcional: Loguear advertencia o fallar si es una variable crítica
		log.Printf("Advertencia: Variable de entorno %s no definida", key)
	}
	return fallback
}
