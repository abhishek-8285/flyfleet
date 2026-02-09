package config

import (
	"os"
)

type Config struct {
	Port           string
	DatabaseURL    string
	AuthServiceURL string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8085"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/flyfleet?sslmode=disable"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
