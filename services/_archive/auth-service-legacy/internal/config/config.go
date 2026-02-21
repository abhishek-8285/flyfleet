package config

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret         string
	LogLevel          string
	FleetCoreServiceURL string
	AlertServiceURL   string
}

func Load() *Config {
	v := viper.New()

	// Defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.log_level", "info")
	v.SetDefault("security.jwt_secret", "") // Should come from config/env
	v.SetDefault("services.fleet_core_url", "http://fleet-core-service:8084")
	v.SetDefault("services.alert_url", "http://alert-service:8085")
	// Database URL doesn't stick to a nice toml structure in old code, but we can map it
	v.SetDefault("database.url", "")

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
	v.BindEnv("security.jwt_secret", "JWT_SECRET")
	v.BindEnv("server.log_level", "LOG_LEVEL")
	v.BindEnv("services.fleet_core_url", "FLEET_CORE_SERVICE_URL")
	v.BindEnv("services.alert_url", "ALERT_SERVICE_URL")

	return &Config{
		Port:               v.GetString("server.port"),
		DatabaseURL:        v.GetString("database.url"),
		JWTSecret:          v.GetString("security.jwt_secret"),
		LogLevel:           v.GetString("server.log_level"),
		FleetCoreServiceURL: v.GetString("services.fleet_core_url"),
		AlertServiceURL:   v.GetString("services.alert_url"),
	}
}