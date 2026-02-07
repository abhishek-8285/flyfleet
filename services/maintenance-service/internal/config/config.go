package config

import (
"os"
"strconv"
"time"
)

type Config struct {
ServerPort     string
DatabaseURL    string
JWTSecret      string
LogLevel       string
RequestTimeout time.Duration
}

func Load() *Config {
timeout, _ := strconv.Atoi(getEnv("REQUEST_TIMEOUT", "30"))

return &Config{
ServerPort:     getEnv("SERVER_PORT", "8080"),
DatabaseURL:    getEnv("DATABASE_URL", "postgres://localhost:5432/maintenance"),
JWTSecret:      getEnv("JWT_SECRET", "secret-key"),
LogLevel:       getEnv("LOG_LEVEL", "info"),
RequestTimeout: time.Duration(timeout) * time.Second,
}
}

func getEnv(key, defaultValue string) string {
if value := os.Getenv(key); value != "" {
return value
}
return defaultValue
}
