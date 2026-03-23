package config

import (
	"context"
	"strings"

	"github.com/flyfleet/fleet-core-service/internal/pkg/secrets"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Port                    string
	DatabaseURL             string
	NatsURL                 string
	RedisURL                string
	LogLevel                string
	JWTSecret               string
	AuthServiceURL          string
	AlertServiceURL         string
	AdminSupportServiceURL  string
	BillingServiceURL       string
	TrackingServiceURL      string
	ReportsServiceURL       string
	SettingsServiceURL      string
	GovernanceServiceURL    string
	AnalyticsServiceURL     string
	IntegrationsServiceURL  string
	ComplianceServiceURL    string
	VehicleServiceURL       string
	TripServiceURL          string
	MaintenanceServiceURL   string
	MLServiceURL            string
	JWKSURL                 string
	BillingStorageType      string // "local" or "s3"
	BillingEngine           string // "premium" or "high_speed"
	S3Endpoint              string
	S3Bucket                string
	S3AccessKey             string
	S3SecretKey             string
	S3UseSSL                bool
}

func Load() *Config {
	v := viper.New()

	// Defaults
	v.SetDefault("server.port", "8084") // Fleet Core default port from compose
	// Database URL default for local dev
	v.SetDefault("database.url", "")
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("redis.url", "localhost:6380")
	v.SetDefault("server.log_level", "info")
	v.SetDefault("security.jwt_secret", "")
	v.SetDefault("services.auth_url", "")
	v.SetDefault("services.alert_url", "")
	v.SetDefault("services.admin_support_url", "")
	v.SetDefault("services.billing_url", "")
	v.SetDefault("services.tracking_url", "")
	v.SetDefault("services.reports_url", "")
	v.SetDefault("services.settings_url", "")
	v.SetDefault("services.governance_url", "")
	v.SetDefault("services.analytics_url", "")
	v.SetDefault("services.integrations_url", "")
	v.SetDefault("services.compliance_url", "")
	v.SetDefault("services.vehicle_url", "")
	v.SetDefault("services.trip_url", "")
	v.SetDefault("services.maintenance_url", "")
	v.SetDefault("services.ml_url", "http://localhost:8000") // Sidecar default
	v.SetDefault("security.jwks_url", "http://localhost:8084/.well-known/jwks.json")
	v.SetDefault("billing.storage_type", "local")
	v.SetDefault("billing.engine", "premium")
	v.SetDefault("billing.s3_endpoint", "s3.amazonaws.com")
	v.SetDefault("billing.s3_bucket", "flyfleet-invoices")
	v.SetDefault("billing.s3_use_ssl", true)

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
	v.BindEnv("nats.url", "NATS_URL")
	v.BindEnv("redis.url", "REDIS_URL")
	v.BindEnv("server.log_level", "LOG_LEVEL")
	v.BindEnv("security.jwt_secret", "JWT_SECRET")
	v.BindEnv("services.auth_url", "AUTH_SERVICE_URL")
	v.BindEnv("services.alert_url", "ALERT_SERVICE_URL")
	v.BindEnv("services.admin_support_url", "ADMIN_SUPPORT_SERVICE_URL")
	v.BindEnv("services.billing_url", "BILLING_SERVICE_URL")
	v.BindEnv("services.tracking_url", "TRACKING_SERVICE_URL")
	v.BindEnv("services.reports_url", "REPORTS_SERVICE_URL")
	v.BindEnv("services.settings_url", "SETTINGS_SERVICE_URL")
	v.BindEnv("services.governance_url", "GOVERNANCE_SERVICE_URL")
	v.BindEnv("services.analytics_url", "ANALYTICS_SERVICE_URL")
	v.BindEnv("services.integrations_url", "INTEGRATIONS_SERVICE_URL")
	v.BindEnv("services.compliance_url", "COMPLIANCE_SERVICE_URL")
	v.BindEnv("services.vehicle_url", "VEHICLE_SERVICE_URL")
	v.BindEnv("services.trip_url", "TRIP_SERVICE_URL")
	v.BindEnv("services.maintenance_url", "MAINTENANCE_SERVICE_URL")
	v.BindEnv("services.ml_url", "ML_SERVICE_URL")
	v.BindEnv("security.jwks_url", "AUTH_V2_JWKS_URL")
	v.BindEnv("billing.storage_type", "BILLING_STORAGE_TYPE")
	v.BindEnv("billing.engine", "BILLING_ENGINE")
	v.BindEnv("billing.s3_endpoint", "S3_ENDPOINT")
	v.BindEnv("billing.s3_bucket", "S3_BUCKET")
	v.BindEnv("billing.s3_access_key", "S3_ACCESS_KEY")
	v.BindEnv("billing.s3_secret_key", "S3_SECRET_KEY")

	ctx := context.Background()

	return &Config{
		Port:                   v.GetString("server.port"),
		DatabaseURL:            v.GetString("database.url"),
		NatsURL:                v.GetString("nats.url"),
		RedisURL:               v.GetString("redis.url"),
		LogLevel:               v.GetString("server.log_level"),
		JWTSecret:              secrets.ResolveSecret(ctx, v.GetString("security.jwt_secret")),
		AuthServiceURL:         v.GetString("services.auth_url"),
		AlertServiceURL:        v.GetString("services.alert_url"),
		AdminSupportServiceURL: v.GetString("services.admin_support_url"),
		BillingServiceURL:      v.GetString("services.billing_url"),
		TrackingServiceURL:     v.GetString("services.tracking_url"),
		ReportsServiceURL:      v.GetString("services.reports_url"),
		SettingsServiceURL:     v.GetString("services.settings_url"),
		GovernanceServiceURL:   v.GetString("services.governance_url"),
		AnalyticsServiceURL:    v.GetString("services.analytics_url"),
		IntegrationsServiceURL: v.GetString("services.integrations_url"),
		ComplianceServiceURL:   v.GetString("services.compliance_url"),
		VehicleServiceURL:      v.GetString("services.vehicle_url"),
		TripServiceURL:         v.GetString("services.trip_url"),
		MaintenanceServiceURL:  v.GetString("services.maintenance_url"),
		MLServiceURL:           v.GetString("services.ml_url"),
		JWKSURL:                v.GetString("security.jwks_url"),
		BillingStorageType:     v.GetString("billing.storage_type"),
		BillingEngine:          v.GetString("billing.engine"),
		S3Endpoint:             v.GetString("billing.s3_endpoint"),
		S3Bucket:               v.GetString("billing.s3_bucket"),
		S3AccessKey:            secrets.ResolveSecret(ctx, v.GetString("billing.s3_access_key")),
		S3SecretKey:            secrets.ResolveSecret(ctx, v.GetString("billing.s3_secret_key")),
		S3UseSSL:               v.GetBool("billing.s3_use_ssl"),
	}
}
