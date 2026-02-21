package config

import (
	"os"
)

type Config struct {
	GRPCPort    string
	HTTPPort    string
	DatabaseURL string
	JWTSecret   string
	LogLevel    string
	Environment string
}

func Load() *Config {
	return &Config{
		GRPCPort:    getEnv("GRPC_PORT", "50051"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Environment: getEnv("ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
