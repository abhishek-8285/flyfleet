package config

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Port        string
	DatabaseURL string
	NatsURL     string
	RedisURL    string
	LogLevel    string
	JWTSecret   string
}

func Load() *Config {
	v := viper.New()

	// Defaults
	v.SetDefault("server.port", "8084") // Fleet Core default port from compose
	// Database URL default for local dev
	v.SetDefault("database.url", "postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable")
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("redis.url", "localhost:6380")
	v.SetDefault("server.log_level", "info")
	v.SetDefault("security.jwt_secret", "") // Should ideally not have a default insecure secret

	// Config file setup
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath("../../") 
	v.AddConfigPath(".")      

	// Env vars
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			zap.S().Warnf("Error reading config file: %v", err)
		}
	}

	// Manual bindings for Docker compatibility
	v.BindEnv("server.port", "PORT")
	v.BindEnv("database.url", "DATABASE_URL")
	v.BindEnv("nats.url", "NATS_URL")
	v.BindEnv("redis.url", "REDIS_URL")
	v.BindEnv("server.log_level", "LOG_LEVEL")
	v.BindEnv("security.jwt_secret", "JWT_SECRET")

	return &Config{
		Port:        v.GetString("server.port"),
		DatabaseURL: v.GetString("database.url"),
		NatsURL:     v.GetString("nats.url"),
		RedisURL:    v.GetString("redis.url"),
		LogLevel:    v.GetString("server.log_level"),
		JWTSecret:   v.GetString("security.jwt_secret"),
	}
}
