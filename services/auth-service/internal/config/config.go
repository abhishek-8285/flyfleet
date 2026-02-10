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
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		DriverServiceURL:  getEnv("DRIVER_SERVICE_URL", ""),
		VehicleServiceURL: getEnv("VEHICLE_SERVICE_URL", ""),
		TripServiceURL:    getEnv("TRIP_SERVICE_URL", ""),
		AlertServiceURL:   getEnv("ALERT_SERVICE_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}