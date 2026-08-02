package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server       ServerConfig       `yaml:"server"`
	NATS         NATSConfig         `yaml:"nats"`
	Kafka        KafkaConfig        `yaml:"kafka"`
	ClickHouse   ClickHouseConfig   `yaml:"clickhouse"`
	Redis        RedisConfig        `yaml:"redis"`
	Temporal     TemporalConfig     `yaml:"temporal"`
	Inference    InferenceConfig    `yaml:"inference"`
	FeatureStore FeatureStoreConfig `yaml:"feature_store"`
	Drift        DriftConfig        `yaml:"drift"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type NATSConfig struct {
	URL      string `yaml:"url"`
	Stream   string `yaml:"stream"`
	Replicas int    `yaml:"replicas"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	GroupID string   `yaml:"group_id"`
}

type ClickHouseConfig struct {
	DSN string `yaml:"dsn"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type TemporalConfig struct {
	HostPort  string `yaml:"host_port"`
	Namespace string `yaml:"namespace"`
	TaskQueue string `yaml:"task_queue"`
}

type InferenceConfig struct {
	ModelDir       string        `yaml:"model_dir"`
	HotSwapEnabled bool          `yaml:"hot_swap_enabled"`
	Timeout        time.Duration `yaml:"timeout"`
}

type FeatureStoreConfig struct {
	OnlineBackend  string `yaml:"online_backend"`  // "redis" | "dynamodb" | "feast"
	OfflineBackend string `yaml:"offline_backend"` // "clickhouse" | "s3"
	FeastURL       string `yaml:"feast_url"`
}

type DriftConfig struct {
	CheckInterval time.Duration `yaml:"check_interval"`
	Threshold     float64       `yaml:"threshold"`
	WindowSize    int           `yaml:"window_size"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	setDefaults(cfg)
	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.NATS.URL == "" {
		cfg.NATS.URL = "nats://localhost:4222"
	}
	if cfg.Redis.Addr == "" {
		cfg.Redis.Addr = "localhost:6379"
	}
	if cfg.Temporal.HostPort == "" {
		cfg.Temporal.HostPort = "localhost:7233"
	}
	if cfg.Inference.Timeout == 0 {
		cfg.Inference.Timeout = 5 * time.Second
	}
	if cfg.Drift.CheckInterval == 0 {
		cfg.Drift.CheckInterval = time.Hour
	}
	if cfg.Drift.Threshold == 0 {
		cfg.Drift.Threshold = 0.1
	}
}
