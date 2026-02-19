package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	LogLevel    string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8087"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "super-secret-key"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
