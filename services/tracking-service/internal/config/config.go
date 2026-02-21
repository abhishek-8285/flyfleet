package config

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Port        string
	GRPCPort    string
	DatabaseURL  string
	NatsURL      string
	JWTSecret    string
	LogLevel     string
	MLServiceURL string
	RedisAddr    string
	RedisPassword string
	RedisDB      int
}

func Load() *Config {
	v := viper.New()

	v.SetDefault("server.port", "8082")
	v.SetDefault("server.grpc_port", "50053")
	v.SetDefault("database.url", "")
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("server.log_level", "info")
	v.SetDefault("security.jwt_secret", "")
	v.SetDefault("services.ml_url", "http://fleet-ml-service:8095")
	v.SetDefault("redis.addr", "redis:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

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
	v.BindEnv("server.grpc_port", "GRPC_PORT")
	v.BindEnv("database.url", "DATABASE_URL")
	v.BindEnv("nats.url", "NATS_URL")
	v.BindEnv("server.log_level", "LOG_LEVEL")
	v.BindEnv("security.jwt_secret", "JWT_SECRET")
	v.BindEnv("services.ml_url", "ML_SERVICE_URL")
	v.BindEnv("redis.addr", "REDIS_ADDR")
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("redis.db", "REDIS_DB")

	return &Config{
		Port:        v.GetString("server.port"),
		GRPCPort:    v.GetString("server.grpc_port"),
		DatabaseURL: v.GetString("database.url"),
		NatsURL:      v.GetString("nats.url"),
		LogLevel:     v.GetString("server.log_level"),
		JWTSecret:    v.GetString("security.jwt_secret"),
		MLServiceURL: v.GetString("services.ml_url"),
		RedisAddr:    v.GetString("redis.addr"),
		RedisPassword: v.GetString("redis.password"),
		RedisDB:      v.GetInt("redis.db"),
	}
}

func (c *Config) GetJWTSecret() string {
	return c.JWTSecret
}
