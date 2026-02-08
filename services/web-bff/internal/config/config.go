package config

import (
	"os"
)

type Config struct {
	Port              string
	AuthServiceURL    string
	VehicleServiceURL string
	JWTSecret         string
	LogLevel          string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8081"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		VehicleServiceURL: getEnv("VEHICLE_SERVICE_URL", "http://vehicle-service:8084"),
		JWTSecret:      getEnv("JWT_SECRET", "super-secret-key"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
