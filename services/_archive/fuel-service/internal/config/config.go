package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort      string
	DatabaseURL     string
	LogLevel        string
	RequestTimeout  time.Duration
	VehicleService  string
	DriverService   string
}

func Load() *Config {
	timeout, _ := strconv.Atoi(getEnv("REQUEST_TIMEOUT", "30"))

	return &Config{
		ServerPort:     getEnv("SERVER_PORT", "8082"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		RequestTimeout: time.Duration(timeout) * time.Second,
		VehicleService: getEnv("VEHICLE_SERVICE_URL", "http://localhost:8080"),
		DriverService:  getEnv("DRIVER_SERVICE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
