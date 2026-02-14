package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	LogLevel     string
	JWTSecret    string
	ServiceName  string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8082"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgresql://postgres:password@localhost:5432/flyfleet"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		JWTSecret:    getEnv("JWT_SECRET", "your-jwt-secret-key"),
		ServiceName:  "support-service",
		ReadTimeout:  parseInt(getEnv("READ_TIMEOUT", "15")),
		WriteTimeout: parseInt(getEnv("WRITE_TIMEOUT", "15")),
		IdleTimeout:  parseInt(getEnv("IDLE_TIMEOUT", "60")),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(value string) int {
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return 0
}