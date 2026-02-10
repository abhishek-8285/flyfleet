package config

import (
	"os"
)

type Config struct {
	Port                        string
	AuthServiceURL              string
	VehicleServiceURL           string
	MaintenanceServiceURL       string
	FuelServiceURL              string
	GovernanceServiceURL        string
	BillingServiceURL           string
	NotificationServiceURL      string
	ReportsServiceURL           string
	DRServiceURL                string
	SettingsServiceURL          string
	IntegrationsServiceURL      string
	DeviceDiagnosticsServiceURL string
	JWTSecret                   string
	LogLevel                    string
}

func Load() *Config {
	return &Config{
		Port:                        getEnv("PORT", "8081"),
		AuthServiceURL:              getEnv("AUTH_SERVICE_URL", ""),
		VehicleServiceURL:           getEnv("VEHICLE_SERVICE_URL", "http://vehicle-service:8084"),
		MaintenanceServiceURL:       getEnv("MAINTENANCE_SERVICE_URL", "http://maintenance-service:8086"),
		FuelServiceURL:              getEnv("FUEL_SERVICE_URL", "http://fuel-service:8082"),
		GovernanceServiceURL:        getEnv("GOVERNANCE_SERVICE_URL", "http://governance-service:8085"),
		BillingServiceURL:           getEnv("BILLING_SERVICE_URL", "http://billing-service:8080"),
		NotificationServiceURL:      getEnv("NOTIFICATION_SERVICE_URL", "http://notification-service:8080"),
		ReportsServiceURL:           getEnv("REPORTS_SERVICE_URL", "http://reports-service:8083"),
		DRServiceURL:                getEnv("DR_SERVICE_URL", "http://dr-service:8084"),
		SettingsServiceURL:          getEnv("SETTINGS_SERVICE_URL", "http://settings-service:8085"),
		IntegrationsServiceURL:      getEnv("INTEGRATIONS_SERVICE_URL", "http://integrations-service:8087"),
		DeviceDiagnosticsServiceURL: getEnv("DEVICE_DIAGNOSTICS_SERVICE_URL", "http://device-diagnostics-service:8088"),
		JWTSecret:                   getEnv("JWT_SECRET", ""),
		LogLevel:                    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
