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
}

func Load() *Config {
	v := viper.New()

	v.SetDefault("server.port", "8082")
	v.SetDefault("database.url", "")
	v.SetDefault("nats.url", "nats://localhost:4222")
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
	v.BindEnv("nats.url", "NATS_URL")
	v.BindEnv("server.log_level", "LOG_LEVEL")
	v.BindEnv("security.jwt_secret", "JWT_SECRET")

	return &Config{
		Port:        v.GetString("server.port"),
		DatabaseURL: v.GetString("database.url"),
		NatsURL:     v.GetString("nats.url"),
		LogLevel:    v.GetString("server.log_level"),
		JWTSecret:   v.GetString("security.jwt_secret"),
	}
}

func (c *Config) GetJWTSecret() string {
	return c.JWTSecret
}
