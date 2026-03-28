package settings

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Configuration metrics
	ConfigurationsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_configurations_total",
		Help: "Total configuration settings",
	}, []string{"org_id", "config_type", "scope"})

	ConfigUpdatesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settings_config_updates_total",
		Help: "Total configuration changes",
	}, []string{"org_id", "config_type", "update_type", "user_role"})

	ConfigValidationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settings_config_validation_errors_total",
		Help: "Configuration validation errors",
	}, []string{"org_id", "error_type", "field"})

	ConfigRollbacksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settings_config_rollbacks_total",
		Help: "Configuration rollbacks performed",
	}, []string{"org_id", "rollback_reason"})

	// Organization settings
	OrganizationSettingsVersion = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_org_settings_version",
		Help: "Organization settings schema version",
	}, []string{"org_id"})

	SettingsSyncStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_sync_status",
		Help: "Settings sync status (0=failed, 1=success)",
	}, []string{"org_id", "sync_target"})

	// Feature flag metrics
	FeatureFlagsEnabledTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_feature_flags_enabled_total",
		Help: "Enabled feature flags per organization",
	}, []string{"org_id", "feature"})

	FeatureFlagOverridesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_feature_flag_overrides_total",
		Help: "Feature flag overrides (per org/user)",
	}, []string{"org_id", "feature", "override_type"})

	FeatureFlagEvaluationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settings_feature_flag_evaluations_total",
		Help: "Feature flag condition evaluations",
	}, []string{"org_id", "feature", "result"})

	// Notification settings
	NotificationRulesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_notification_rules_total",
		Help: "Configured notification rules",
	}, []string{"org_id", "rule_type"})

	NotificationChannelsConfigured = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_notification_channels_configured_total",
		Help: "Notification channels configured",
	}, []string{"org_id", "channel_type"})

	NotificationQuotaUsage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_notification_quota_usage_percentage",
		Help: "Notification quota usage percentage",
	}, []string{"org_id", "channel_type", "period"})

	// Branding settings
	BrandingCustomizationsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_branding_customizations_total",
		Help: "Custom branding elements",
	}, []string{"org_id", "customization_type"})

	// Integration settings
	IntegrationSettingsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_integration_configs_total",
		Help: "Integration configurations",
	}, []string{"org_id", "integration_type"})

	IntegrationSettingsValidationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settings_integration_validation_errors_total",
		Help: "Integration configuration validation errors",
	}, []string{"org_id", "integration_type"})

	// Security settings
	SecurityPoliciesConfigured = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_security_policies_configured_total",
		Help: "Security policies configured",
	}, []string{"org_id", "policy_type"})

	SecurityPolicyViolations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settings_security_policy_violations_total",
		Help: "Security policy violations detected",
	}, []string{"org_id", "policy_type", "violation_type"})

	// Compliance settings
	ComplianceRulesConfigured = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_compliance_rules_configured_total",
		Help: "Compliance rules configured",
	}, []string{"org_id", "regulation"})

	ComplianceRuleEvaluationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settings_compliance_rule_evaluation_errors_total",
		Help: "Compliance rule evaluation failures",
	}, []string{"org_id", "rule_id"})

	// API settings
	APIKeyRotationEnabled = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_api_key_rotation_enabled_percentage",
		Help: "API key rotation enabled percentage",
	}, []string{"org_id"})

	APIKeyExpiryAlerts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settings_api_key_expiry_alerts_total",
		Help: "API key expiry notifications sent",
	}, []string{"org_id"})

	// Billing settings
	BillingCycleConfigured = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_billing_cycle_configured_total",
		Help: "Billing cycles configured",
	}, []string{"org_id", "cycle_type"})

	BillingThresholdAlerts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settings_billing_threshold_alerts_total",
		Help: "Billing threshold alerts triggered",
	}, []string{"org_id", "threshold_type"})

	// User preferences
	UserPreferenceSyncsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settings_user_preference_syncs_total",
		Help: "User preference synchronization events",
	}, []string{"org_id", "sync_type"})

	// Retention policies
	DataRetentionPolicies = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_data_retention_policies_total",
		Help: "Data retention policies configured",
	}, []string{"org_id", "data_type"})

	RetentionPolicyCompliance = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_retention_policy_compliance_percentage",
		Help: "Compliance with retention policies",
	}, []string{"org_id"})

	// Multi-tenant isolation
	TenantIsolationViolations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settings_tenant_isolation_violations_total",
		Help: "Multi-tenancy isolation violations",
	}, []string{"org_id", "violation_type"})

	// Configuration drift
	ConfigDriftDetectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settings_config_drift_detected_total",
		Help: "Configuration drift from baseline",
	}, []string{"org_id", "drift_type", "severity"})

	ConfigDriftResolutionTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "settings_config_drift_resolution_time_seconds",
		Help:    "Time to resolve configuration drift",
		Buckets: prometheus.ExponentialBuckets(60, 2, 8),
	}, []string{"org_id"})

	// Validation metrics
	ConfigSchemaValidationTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "settings_config_schema_validation_time_ms",
		Help:    "Configuration schema validation latency",
		Buckets: prometheus.LinearBuckets(1, 10, 20),
	}, []string{"org_id"})

	ConfigChangeImpactAssessment = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "settings_config_change_impact_score",
		Help: "Impact score of configuration changes (0-100)",
	}, []string{"org_id", "change_type"})
)

// Global metrics instance
var GlobalMetrics = struct {
	ConfigurationsTotal                 *prometheus.GaugeVec
	ConfigUpdatesTotal                  *prometheus.CounterVec
	ConfigValidationErrors              *prometheus.CounterVec
	ConfigRollbacksTotal                *prometheus.CounterVec
	OrganizationSettingsVersion         *prometheus.GaugeVec
	SettingsSyncStatus                  *prometheus.GaugeVec
	FeatureFlagsEnabledTotal            *prometheus.GaugeVec
	FeatureFlagOverridesTotal           *prometheus.GaugeVec
	FeatureFlagEvaluationsTotal         *prometheus.CounterVec
	NotificationRulesTotal              *prometheus.GaugeVec
	NotificationChannelsConfigured      *prometheus.GaugeVec
	NotificationQuotaUsage              *prometheus.GaugeVec
	BrandingCustomizationsTotal         *prometheus.GaugeVec
	IntegrationSettingsTotal            *prometheus.GaugeVec
	IntegrationSettingsValidationErrors *prometheus.CounterVec
	SecurityPoliciesConfigured          *prometheus.GaugeVec
	SecurityPolicyViolations            *prometheus.CounterVec
	ComplianceRulesConfigured           *prometheus.GaugeVec
	ComplianceRuleEvaluationErrors      *prometheus.CounterVec
	APIKeyRotationEnabled               *prometheus.GaugeVec
	APIKeyExpiryAlerts                  *prometheus.CounterVec
	BillingCycleConfigured              *prometheus.GaugeVec
	BillingThresholdAlerts              *prometheus.CounterVec
	UserPreferenceSyncsTotal            *prometheus.CounterVec
	DataRetentionPolicies               *prometheus.GaugeVec
	RetentionPolicyCompliance           *prometheus.GaugeVec
	TenantIsolationViolations           *prometheus.CounterVec
	ConfigDriftDetectedTotal            *prometheus.CounterVec
	ConfigDriftResolutionTime           *prometheus.HistogramVec
	ConfigSchemaValidationTime          *prometheus.HistogramVec
	ConfigChangeImpactAssessment        *prometheus.GaugeVec
}{
	ConfigurationsTotal:                 ConfigurationsTotal,
	ConfigUpdatesTotal:                  ConfigUpdatesTotal,
	ConfigValidationErrors:              ConfigValidationErrors,
	ConfigRollbacksTotal:                ConfigRollbacksTotal,
	OrganizationSettingsVersion:         OrganizationSettingsVersion,
	SettingsSyncStatus:                  SettingsSyncStatus,
	FeatureFlagsEnabledTotal:            FeatureFlagsEnabledTotal,
	FeatureFlagOverridesTotal:           FeatureFlagOverridesTotal,
	FeatureFlagEvaluationsTotal:         FeatureFlagEvaluationsTotal,
	NotificationRulesTotal:              NotificationRulesTotal,
	NotificationChannelsConfigured:      NotificationChannelsConfigured,
	NotificationQuotaUsage:              NotificationQuotaUsage,
	BrandingCustomizationsTotal:         BrandingCustomizationsTotal,
	IntegrationSettingsTotal:            IntegrationSettingsTotal,
	IntegrationSettingsValidationErrors: IntegrationSettingsValidationErrors,
	SecurityPoliciesConfigured:          SecurityPoliciesConfigured,
	SecurityPolicyViolations:            SecurityPolicyViolations,
	ComplianceRulesConfigured:           ComplianceRulesConfigured,
	ComplianceRuleEvaluationErrors:      ComplianceRuleEvaluationErrors,
	APIKeyRotationEnabled:               APIKeyRotationEnabled,
	APIKeyExpiryAlerts:                  APIKeyExpiryAlerts,
	BillingCycleConfigured:              BillingCycleConfigured,
	BillingThresholdAlerts:              BillingThresholdAlerts,
	UserPreferenceSyncsTotal:            UserPreferenceSyncsTotal,
	DataRetentionPolicies:               DataRetentionPolicies,
	RetentionPolicyCompliance:           RetentionPolicyCompliance,
	TenantIsolationViolations:           TenantIsolationViolations,
	ConfigDriftDetectedTotal:            ConfigDriftDetectedTotal,
	ConfigDriftResolutionTime:           ConfigDriftResolutionTime,
	ConfigSchemaValidationTime:          ConfigSchemaValidationTime,
	ConfigChangeImpactAssessment:        ConfigChangeImpactAssessment,
}
