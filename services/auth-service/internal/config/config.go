package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret         string
	LogLevel          string
	DriverServiceURL  string
	VehicleServiceURL string
	TripServiceURL    string
	AlertServiceURL   string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		DriverServiceURL:  getEnv("DRIVER_SERVICE_URL", "http://localhost:8083/api/v1"),
		VehicleServiceURL: getEnv("VEHICLE_SERVICE_URL", "http://localhost:8084/api/v1"),
		TripServiceURL:    getEnv("TRIP_SERVICE_URL", "http://localhost:8082/api/v1"),
		AlertServiceURL:   getEnv("ALERT_SERVICE_URL", "http://localhost:8085/api/v1"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}