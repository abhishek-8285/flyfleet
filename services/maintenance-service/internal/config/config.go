package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Port      int
	DBHost    string
	DBPort    int
	DBUser    string
	DBPass    string
	DBName    string
	SSLMode   string
	JWTSecret string
	NATSURL   string
	RedisURL  string
}

func Load() (*Config, error) {
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("SSL_MODE", "disable")

	viper.AutomaticEnv()

	cfg := &Config{
		Port:      viper.GetInt("PORT"),
		DBHost:    viper.GetString("DB_HOST"),
		DBPort:    viper.GetInt("DB_PORT"),
		DBUser:    viper.GetString("DB_USER"),
		DBPass:    viper.GetString("DB_PASS"),
		DBName:    viper.GetString("DB_NAME"),
		SSLMode:   viper.GetString("SSL_MODE"),
		JWTSecret: viper.GetString("JWT_SECRET"),
		NATSURL:   viper.GetString("NATS_URL"),
		RedisURL:  viper.GetString("REDIS_URL"),
	}

	if cfg.DBUser == "" || cfg.DBPass == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("missing mandatory database environment variables")
	}

	return cfg, nil
}

