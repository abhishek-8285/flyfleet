package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds the fleet-core service configuration.
type Config struct {
	Port        string   `mapstructure:"port"`
	DatabaseURL string   `mapstructure:"database_url"`
	RedisURL    string   `mapstructure:"redis_url"`
	NatsURL     string   `mapstructure:"nats_url"`
	JWTSecret   string   `mapstructure:"jwt_secret"`
	JWKSURL     string   `mapstructure:"jwks_url"`
	LogLevel    string   `mapstructure:"log_level"`
	Environment string   `mapstructure:"environment"`
	Flags       FeatureFlags `mapstructure:"feature_flags"`
	IAM         IAMConfig    `mapstructure:"iam"`
	Billing     BillingConfig `mapstructure:"billing"`
}

// BillingConfig configures the invoice generator and billing services
type BillingConfig struct {
	Enabled               bool `mapstructure:"enabled"`
	ChromeTimeoutSeconds  int  `mapstructure:"chrome_timeout_seconds"`
}

// FeatureFlags controls runtime toggles for the tracking feature.
type FeatureFlags struct {
	VehicleMapEnabled  bool `mapstructure:"vehicle_map_enabled"`
	RealTimeUpdates    bool `mapstructure:"real_time_updates"`
	GeofenceAlerts     bool `mapstructure:"geofence_alerts"`
	HistoryTrail       bool `mapstructure:"history_trail"`
	MaxVehiclesOnMap   int  `mapstructure:"max_vehicles_on_map"`
}

// IAMConfig holds configuration for the IAM (Identity & Access Management) domain.
type IAMConfig struct {
	Enabled             bool          `mapstructure:"enabled"`
	BcryptCost          int           `mapstructure:"bcrypt_cost"`
	MaxLoginAttempts    int           `mapstructure:"max_login_attempts"`
	LockoutDuration     time.Duration `mapstructure:"lockout_duration"`
	PasswordHistoryLen  int           `mapstructure:"password_history_length"`
	JWTAccessTTL        time.Duration `mapstructure:"jwt_access_ttl"`
	JWTRefreshTTL       time.Duration `mapstructure:"jwt_refresh_ttl"`
	EmailProvider       string        `mapstructure:"email_provider"`
	EmailSendGridKey    string        `mapstructure:"email_sendgrid_key"`
	EmailSMTPHost       string        `mapstructure:"email_smtp_host"`
	EmailSMTPPort       int           `mapstructure:"email_smtp_port"`
	EmailSMTPUser       string        `mapstructure:"email_smtp_user"`
	EmailSMTPPassword   string        `mapstructure:"email_smtp_password"`
	EmailFromAddress    string        `mapstructure:"email_from_address"`
	EmailFromName       string        `mapstructure:"email_from_name"`
}

func Load() *Config {
	v := viper.New()

	// Global defaults
	v.SetDefault("port", "8084")
	v.SetDefault("database_url", "postgres://fleet_user:fleet_password@localhost:5432/fleet_db?sslmode=disable")
	v.SetDefault("redis_url", "redis://localhost:6379/0")
	v.SetDefault("nats_url", "nats://localhost:4222")
	v.SetDefault("jwt_secret", "")
	v.SetDefault("jwks_url", "http://localhost:8080/.well-known/jwks.json")
	v.SetDefault("log_level", "info")
	v.SetDefault("environment", "development")

	// Feature flag defaults
	v.SetDefault("feature_flags.vehicle_map_enabled", true)
	v.SetDefault("feature_flags.real_time_updates", true)
	v.SetDefault("feature_flags.geofence_alerts", false)
	v.SetDefault("feature_flags.history_trail", true)
	v.SetDefault("feature_flags.max_vehicles_on_map", 500)

	// IAM defaults
	v.SetDefault("iam.enabled", true)
	v.SetDefault("iam.bcrypt_cost", 12)
	v.SetDefault("iam.max_login_attempts", 5)
	v.SetDefault("iam.lockout_duration", 15)
	v.SetDefault("iam.password_history_length", 5)
	v.SetDefault("iam.jwt_access_ttl", 900)
	v.SetDefault("iam.jwt_refresh_ttl", 604800)
	v.SetDefault("iam.email_provider", "noop")
	v.SetDefault("iam.email_smtp_port", 587)

	// Billing defaults
	v.SetDefault("billing.enabled", true)
	v.SetDefault("billing.chrome_timeout_seconds", 30)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Config file search
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(".")
	v.AddConfigPath("./cmd")
	v.AddConfigPath("../")
	_ = v.ReadInConfig() // ignore — config file is optional

	return &Config{
		Port:        v.GetString("port"),
		DatabaseURL: v.GetString("database_url"),
		RedisURL:    v.GetString("redis_url"),
		NatsURL:     v.GetString("nats_url"),
		JWTSecret:   v.GetString("jwt_secret"),
		JWKSURL:     v.GetString("jwks_url"),
		LogLevel:    v.GetString("log_level"),
		Environment: v.GetString("environment"),
		Flags: FeatureFlags{
			VehicleMapEnabled:  v.GetBool("feature_flags.vehicle_map_enabled"),
			RealTimeUpdates:    v.GetBool("feature_flags.real_time_updates"),
			GeofenceAlerts:     v.GetBool("feature_flags.geofence_alerts"),
			HistoryTrail:       v.GetBool("feature_flags.history_trail"),
			MaxVehiclesOnMap:   v.GetInt("feature_flags.max_vehicles_on_map"),
		},
		IAM: IAMConfig{
			Enabled:            v.GetBool("iam.enabled"),
			BcryptCost:         v.GetInt("iam.bcrypt_cost"),
			MaxLoginAttempts:   v.GetInt("iam.max_login_attempts"),
			LockoutDuration:    time.Duration(v.GetInt("iam.lockout_duration")) * time.Minute,
			PasswordHistoryLen: v.GetInt("iam.password_history_length"),
			JWTAccessTTL:       time.Duration(v.GetInt("iam.jwt_access_ttl")) * time.Second,
			JWTRefreshTTL:      time.Duration(v.GetInt("iam.jwt_refresh_ttl")) * time.Second,
			EmailProvider:      v.GetString("iam.email_provider"),
			EmailSendGridKey:   v.GetString("iam.email_sendgrid_key"),
			EmailSMTPHost:      v.GetString("iam.email_smtp_host"),
			EmailSMTPPort:      v.GetInt("iam.email_smtp_port"),
			EmailSMTPUser:      v.GetString("iam.email_smtp_user"),
			EmailSMTPPassword:  v.GetString("iam.email_smtp_password"),
			EmailFromAddress:   v.GetString("iam.email_from_address"),
			EmailFromName:      v.GetString("iam.email_from_name"),
		},
		Billing: BillingConfig{
			Enabled:              v.GetBool("billing.enabled"),
			ChromeTimeoutSeconds: v.GetInt("billing.chrome_timeout_seconds"),
		},
	}
}
