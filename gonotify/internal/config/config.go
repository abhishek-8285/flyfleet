package config

import "os"

type Config struct {
	AppEnv   string
	LogLevel string

	HTTPAddr string

	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       string

	NATSURL string

	StorageDriver   string
	StorageBucket   string
	StorageEndpoint string
	StorageAccessKey string
	StorageSecretKey string
	StorageUseSSL   bool
}

func Load() Config {
	return Config{
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),

		DatabaseURL: getEnv(
			"DATABASE_URL",
			"postgres://gonotify:gonotify@localhost:5432/gonotify?sslmode=disable",
		),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnv("REDIS_DB", "0"),

		NATSURL: getEnv("NATS_URL", "nats://localhost:4222"),

		StorageDriver:   getEnv("STORAGE_DRIVER", "minio"),
		StorageBucket:   getEnv("STORAGE_BUCKET", "gonotify"),
		StorageEndpoint: getEnv("STORAGE_ENDPOINT", "localhost:9000"),
		StorageAccessKey: getEnv("STORAGE_ACCESS_KEY", "gonotify"),
		StorageSecretKey: getEnv("STORAGE_SECRET_KEY", "gonotify123"),
		StorageUseSSL:   getEnv("STORAGE_USE_SSL", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
