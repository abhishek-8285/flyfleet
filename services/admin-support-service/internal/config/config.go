package config

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	LogLevel    string
}

func Load() *Config {
	v := viper.New()

	v.SetDefault("server.port", "8086")
	v.SetDefault("database.url", "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable")
	v.SetDefault("server.log_level", "info")
	v.SetDefault("security.jwt_secret", "")

	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath("../../") 
	v.AddConfigPath(".")      

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			zap.S().Warnf("Error reading config file: %v", err)
		}
	}

	v.BindEnv("server.port", "PORT")
	v.BindEnv("database.url", "DATABASE_URL")
	v.BindEnv("server.log_level", "LOG_LEVEL")
	v.BindEnv("security.jwt_secret", "JWT_SECRET")

	return &Config{
		Port:        v.GetString("server.port"),
		DatabaseURL: v.GetString("database.url"),
		LogLevel:    v.GetString("server.log_level"),
		JWTSecret:   v.GetString("security.jwt_secret"),
	}
}
