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
		Port:        getEnv("PORT", "8088"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "super-secret-key"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
