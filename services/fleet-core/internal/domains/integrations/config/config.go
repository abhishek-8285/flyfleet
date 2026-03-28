package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyfleet/fleet-core-service/internal/pkg/secrets"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	LogLevel    string
}

func (c *Config) Validate() error {
	var missing []string

	if c.Port == "" {
		missing = append(missing, "server.port")
	}
	if c.DatabaseURL == "" {
		missing = append(missing, "database.url")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "security.jwt_secret")
	} else if len(c.JWTSecret) < 32 {
		return fmt.Errorf("security.jwt_secret must be at least 32 characters for security (got %d)", len(c.JWTSecret))
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}
	return nil
}

func Load() *Config {
	v := viper.New()

	// Defaults - SECURE: No hardcoded secrets
	v.SetDefault("server.port", "8087")
	v.SetDefault("database.url", "") // MUST be provided via env
	v.SetDefault("server.log_level", "info")
	// JWT secret has NO default - must be provided

	// Config file setup
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath("../../../../")
	v.AddConfigPath("../")
	v.AddConfigPath(".")

	// Env vars
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Manual bindings for Docker compatibility
	v.BindEnv("server.port", "PORT")
	v.BindEnv("database.url", "DATABASE_URL")
	v.BindEnv("server.log_level", "LOG_LEVEL")
	v.BindEnv("security.jwt_secret", "JWT_SECRET")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			zap.S().Warnf("Error reading config file: %v", err)
		}
	}

	ctx := context.Background()
	cfg := &Config{
		Port:        v.GetString("server.port"),
		DatabaseURL: v.GetString("database.url"),
		JWTSecret:   secrets.ResolveSecret(ctx, v.GetString("security.jwt_secret")),
		LogLevel:    v.GetString("server.log_level"),
	}

	// Validate configuration (fail fast on startup)
	if err := cfg.Validate(); err != nil {
		zap.S().Fatal("Configuration validation failed", zap.Error(err))
	}

	return cfg
}
