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
	JWTSecret   string
	LogLevel    string

	// Service specific settings
	Auth     ServiceConfig
	Alerts   ServiceConfig
	Tracking ServiceConfig
}

type ServiceConfig struct {
	Enabled bool
	Port    string
}

func Load() *Config {
	v := viper.New()

	// Global Defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.log_level", "info")
	v.SetDefault("database.url", "")
	v.SetDefault("security.jwt_secret", "")
	v.SetDefault("nats.url", "nats://localhost:4222")

	// Module Defaults
	v.SetDefault("modules.auth.enabled", true)
	v.SetDefault("modules.alerts.enabled", true)
	v.SetDefault("modules.tracking.enabled", true)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Environment variable bindings
	v.BindEnv("server.port", "PORT")
	v.BindEnv("database.url", "DATABASE_URL")
	v.BindEnv("security.jwt_secret", "JWT_SECRET")
	v.BindEnv("nats.url", "NATS_URL")
	v.BindEnv("server.log_level", "LOG_LEVEL")

	return &Config{
		Port:        v.GetString("server.port"),
		DatabaseURL: v.GetString("database.url"),
		NatsURL:     v.GetString("nats.url"),
		JWTSecret:   v.GetString("security.jwt_secret"),
		LogLevel:    v.GetString("server.log_level"),
		Auth: ServiceConfig{
			Enabled: v.GetBool("modules.auth.enabled"),
		},
		Alerts: ServiceConfig{
			Enabled: v.GetBool("modules.alerts.enabled"),
		},
		Tracking: ServiceConfig{
			Enabled: v.GetBool("modules.tracking.enabled"),
		},
	}
}
