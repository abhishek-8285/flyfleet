package config

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Port                        string
	AuthServiceURL              string
	FleetCoreServiceURL         string
	AlertServiceURL             string
	AdminSupportServiceURL      string
	BillingServiceURL           string
	TrackingServiceURL          string

	// Legacy/Backlog Service URLs (for V2 compatibility if needed)
	ReportsServiceURL           string
	DRServiceURL                string
	IntegrationsServiceURL      string
	IndiaIntegrationServiceURL  string

	JWTSecret                   string
	LogLevel                    string
}

func Load() *Config {
	// Initialize Viper
	v := viper.New()

	// Set default values (same as before)
	v.SetDefault("server.port", "8081")
	v.SetDefault("server.log_level", "info")
	v.SetDefault("security.jwt_secret", "")
	v.SetDefault("services.auth_url", "http://auth-service:8080")
	v.SetDefault("services.fleet_core_url", "http://fleet-core-service:8084")
	v.SetDefault("services.alert_url", "http://alert-service:8085")
	v.SetDefault("services.admin_support_url", "http://admin-support-service:8086")
	v.SetDefault("services.billing_url", "http://billing-service:8087")
	v.SetDefault("services.tracking_url", "http://tracking-service:8082")
	v.SetDefault("services.reports_url", "http://reports-service:8083")
	v.SetDefault("services.dr_url", "http://admin-support-service:8086")
	v.SetDefault("services.integrations_url", "http://fleet-core-service:8084")

	v.SetDefault("services.india_integration_url", "http://india-integration-service:8088")

	// Config file setup
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath("../../") // Allow reading from root if running from cmd relative
	v.AddConfigPath(".")      // Current directory

	// Environment variable overrides
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config
	if err := v.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist, we fallback to defaults/env
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			zap.S().Warnf("Error reading config file: %v", err)
		}
	}

	// manual env mapping for docker compatibility (where env vars might not match toml structure exactly)
	// We map the old flat env vars to the new structure or just read them directly if viper auto-env doesn't catch them all perfectly without a prefix
	// Actually, viper.AutomaticEnv() works best with a prefix or exact match.
	// Since our docker-compose uses "AUTH_SERVICE_URL", we need to bind that.
	v.BindEnv("services.auth_url", "AUTH_SERVICE_URL")
	v.BindEnv("services.fleet_core_url", "FLEET_CORE_SERVICE_URL")
	v.BindEnv("services.alert_url", "ALERT_SERVICE_URL")
	v.BindEnv("services.admin_support_url", "ADMIN_SUPPORT_SERVICE_URL")
	v.BindEnv("services.billing_url", "BILLING_SERVICE_URL")
	v.BindEnv("services.tracking_url", "TRACKING_SERVICE_URL")
	v.BindEnv("services.reports_url", "REPORTS_SERVICE_URL")
	v.BindEnv("services.india_integration_url", "INDIA_INTEGRATION_SERVICE_URL")
	v.BindEnv("services.dr_url", "DR_SERVICE_URL")
	v.BindEnv("services.integrations_url", "INTEGRATIONS_SERVICE_URL")
	v.BindEnv("security.jwt_secret", "JWT_SECRET")
	v.BindEnv("server.port", "PORT")
	v.BindEnv("server.log_level", "LOG_LEVEL")

	return &Config{
		Port:                   v.GetString("server.port"),
		AuthServiceURL:         v.GetString("services.auth_url"),
		FleetCoreServiceURL:    v.GetString("services.fleet_core_url"),
		AlertServiceURL:        v.GetString("services.alert_url"),
		AdminSupportServiceURL: v.GetString("services.admin_support_url"),
		BillingServiceURL:      v.GetString("services.billing_url"),
		TrackingServiceURL:     v.GetString("services.tracking_url"),

		ReportsServiceURL:      v.GetString("services.reports_url"),
		DRServiceURL:           v.GetString("services.dr_url"),
		IntegrationsServiceURL: v.GetString("services.integrations_url"),

		IndiaIntegrationServiceURL: v.GetString("services.india_integration_url"),

		JWTSecret:              v.GetString("security.jwt_secret"),
		LogLevel:               v.GetString("server.log_level"),
	}
}
