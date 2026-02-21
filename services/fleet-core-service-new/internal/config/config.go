package config

import (
	"os"
)

type Config struct {
	GRPCPort    string
	DatabaseURL string
	NatsURL     string
	Environment string
	LogLevel    string
}

func Load() *Config {
	return &Config{
		GRPCPort:    getEnvOrDefault("GRPC_PORT", "50052"),
		DatabaseURL: getEnvOrDefault("DATABASE_URL", "postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"),
		NatsURL:     getEnvOrDefault("NATS_URL", "nats://localhost:4222"),
		Environment: getEnvOrDefault("ENV", "development"),
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),
	}
}

func getEnvOrDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
