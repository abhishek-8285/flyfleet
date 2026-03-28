package integrations

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Webhook metrics
	WebhooksRegisteredTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "integrations_webhooks_registered_total",
		Help: "Total webhook endpoints registered",
	})

	WebhooksDeliveredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_webhooks_delivered_total",
		Help: "Total webhook deliveries attempted",
	}, []string{"org_id", "webhook_type", "target_url"})

	WebhookDeliveriesSuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_webhook_deliveries_success_total",
		Help: "Total successful webhook deliveries",
	}, []string{"org_id", "webhook_type", "target_url"})

	WebhookDeliveriesFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_webhook_deliveries_failed_total",
		Help: "Total failed webhook deliveries",
	}, []string{"org_id", "webhook_type", "target_url", "failure_reason"})

	WebhookDeliveryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "integrations_webhook_delivery_duration_seconds",
		Help:    "Webhook delivery latency",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 8),
	}, []string{"org_id", "webhook_type"})

	WebhookRetryAttemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_webhook_retry_attempts_total",
		Help: "Total webhook retry attempts",
	}, []string{"org_id", "webhook_type", "retry_count"})

	WebhookQueueDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "integrations_webhook_queue_depth",
		Help: "Current webhook delivery queue depth",
	}, []string{"org_id", "priority"})

	// API integration metrics
	APIIntegrationsConfiguredTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "integrations_api_integrations_configured_total",
		Help: "Total external API integrations configured",
	})

	APIIntegrationSyncScheduledTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_api_sync_scheduled_total",
		Help: "Total API integration syncs scheduled",
	}, []string{"integration_type", "org_id"})

	APIIntegrationSyncCompletedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_api_sync_completed_total",
		Help: "Total API integration syncs completed",
	}, []string{"integration_type", "org_id", "result"})

	APIIntegrationSyncDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "integrations_api_sync_duration_seconds",
		Help:    "API integration sync duration",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10),
	}, []string{"integration_type", "org_id"})

	APIIntegrationRateLimitHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_api_rate_limit_hits_total",
		Help: "Total API calls hitting rate limits",
	}, []string{"integration_type", "org_id"})

	// Third-party sync metrics
	ExternalSystemSyncSuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_external_sync_success_total",
		Help: "Total successful external system syncs",
	}, []string{"org_id", "system_type", "sync_type"})

	ExternalSystemSyncFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_external_sync_failures_total",
		Help: "Total external system sync failures",
	}, []string{"org_id", "system_type", "failure_reason"})

	RecordsSyncedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_records_synced_total",
		Help: "Total records synchronized with external systems",
	}, []string{"org_id", "system_type", "record_type"})

	// Data transformation metrics
	DataTransformSuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_data_transform_success_total",
		Help: "Total successful data transformations",
	}, []string{"org_id", "transformation_type"})

	DataTransformFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_data_transform_failures_total",
		Help: "Total data transformation failures",
	}, []string{"org_id", "transformation_type", "error_type"})

	// File transfer metrics
	FileTransfersTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_file_transfers_total",
		Help: "Total file transfers to/from external systems",
	}, []string{"org_id", "direction", "file_type"})

	FileTransferSizeBytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_file_transfer_size_bytes",
		Help: "Total bytes transferred",
	}, []string{"org_id", "direction", "file_type"})

	FileTransferDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "integrations_file_transfer_duration_seconds",
		Help:    "File transfer duration",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10),
	}, []string{"org_id", "direction"})

	// Authentication metrics
	IntegrationAuthSuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_auth_success_total",
		Help: "Successful external system authentications",
	}, []string{"org_id", "integration_type", "auth_method"})

	IntegrationAuthFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_auth_failures_total",
		Help: "Failed external system authentications",
	}, []string{"org_id", "integration_type", "failure_reason"})

	TokenRefreshesTriggeredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_token_refreshes_triggered_total",
		Help: "OAuth token refresh attempts for integrations",
	}, []string{"org_id", "integration_type"})

	// SLA metrics
	IntegrationAvailability = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "integrations_availability_percentage",
		Help: "Integration availability percentage (uptime)",
	}, []string{"org_id", "integration_type"})

	IntegrationLatencyP95 = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "integrations_latency_p95_seconds",
		Help: "Integration API latency p95",
	}, []string{"org_id", "integration_type"})

	// Error classification
	IntegrationErrorsByType = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_errors_total",
		Help: "Total integration errors by type",
	}, []string{"org_id", "integration_type", "error_type", "severity"})

	IntegrationRetryableErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_retryable_errors_total",
		Help: "Errors that triggered retry logic",
	}, []string{"org_id", "integration_type", "error_type"})

	IntegrationPermanentFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_permanent_failures_total",
		Help: "Errors that exceeded retry limit",
	}, []string{"org_id", "integration_type", "error_type"})

	// Data quality
	DataValidationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_data_validation_errors_total",
		Help: "Data quality validation errors in syncs",
	}, []string{"org_id", "integration_type", "validation_rule"})

	DataCompletenessScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "integrations_data_completeness_score_percentage",
		Help: "Data completeness score (0-100)",
	}, []string{"org_id", "integration_type", "entity_type"})

	// Configuration metrics
	IntegrationConfigChangesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_config_changes_total",
		Help: "Integration configuration changes",
	}, []string{"org_id", "integration_type", "change_type"})

	InvalidConfigurations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integrations_invalid_configs_total",
		Help: "Invalid integration configurations detected",
	}, []string{"org_id", "validation_error"})

	// Health check metrics
	IntegrationHealthCheckResults = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "integrations_health_check_result",
		Help: "Integration health check result (0=failing, 1=healthy)",
	}, []string{"org_id", "integration_type", "check_type"})
)

// Global metrics instance
var GlobalMetrics = struct {
	WebhooksRegisteredTotal          prometheus.Counter
	WebhooksDeliveredTotal           *prometheus.CounterVec
	WebhookDeliveriesSuccessTotal    *prometheus.CounterVec
	WebhookDeliveriesFailedTotal     *prometheus.CounterVec
	WebhookDeliveryDuration          *prometheus.HistogramVec
	WebhookRetryAttemptsTotal        *prometheus.CounterVec
	WebhookQueueDepth                *prometheus.GaugeVec
	APIIntegrationsConfiguredTotal   prometheus.Counter
	APIIntegrationSyncScheduledTotal *prometheus.CounterVec
	APIIntegrationSyncCompletedTotal *prometheus.CounterVec
	APIIntegrationSyncDuration       *prometheus.HistogramVec
	APIIntegrationRateLimitHits      *prometheus.CounterVec
	ExternalSystemSyncSuccessTotal   *prometheus.CounterVec
	ExternalSystemSyncFailuresTotal  *prometheus.CounterVec
	RecordsSyncedTotal               *prometheus.CounterVec
	DataTransformSuccessTotal        *prometheus.CounterVec
	DataTransformFailuresTotal       *prometheus.CounterVec
	FileTransfersTotal               *prometheus.CounterVec
	FileTransferSizeBytes            *prometheus.CounterVec
	FileTransferDuration             *prometheus.HistogramVec
	IntegrationAuthSuccessTotal      *prometheus.CounterVec
	IntegrationAuthFailuresTotal     *prometheus.CounterVec
	TokenRefreshesTriggeredTotal     *prometheus.CounterVec
	IntegrationAvailability          *prometheus.GaugeVec
	IntegrationLatencyP95            *prometheus.GaugeVec
	IntegrationErrorsByType          *prometheus.CounterVec
	IntegrationRetryableErrors       *prometheus.CounterVec
	IntegrationPermanentFailures     *prometheus.CounterVec
	DataValidationErrors             *prometheus.CounterVec
	DataCompletenessScore            *prometheus.GaugeVec
	IntegrationConfigChangesTotal    *prometheus.CounterVec
	InvalidConfigurations            *prometheus.CounterVec
	IntegrationHealthCheckResults    *prometheus.GaugeVec
}{
	WebhooksRegisteredTotal:          WebhooksRegisteredTotal,
	WebhooksDeliveredTotal:           WebhooksDeliveredTotal,
	WebhookDeliveriesSuccessTotal:    WebhookDeliveriesSuccessTotal,
	WebhookDeliveriesFailedTotal:     WebhookDeliveriesFailedTotal,
	WebhookDeliveryDuration:          WebhookDeliveryDuration,
	WebhookRetryAttemptsTotal:        WebhookRetryAttemptsTotal,
	WebhookQueueDepth:                WebhookQueueDepth,
	APIIntegrationsConfiguredTotal:   APIIntegrationsConfiguredTotal,
	APIIntegrationSyncScheduledTotal: APIIntegrationSyncScheduledTotal,
	APIIntegrationSyncCompletedTotal: APIIntegrationSyncCompletedTotal,
	APIIntegrationSyncDuration:       APIIntegrationSyncDuration,
	APIIntegrationRateLimitHits:      APIIntegrationRateLimitHits,
	ExternalSystemSyncSuccessTotal:   ExternalSystemSyncSuccessTotal,
	ExternalSystemSyncFailuresTotal:  ExternalSystemSyncFailuresTotal,
	RecordsSyncedTotal:               RecordsSyncedTotal,
	DataTransformSuccessTotal:        DataTransformSuccessTotal,
	DataTransformFailuresTotal:       DataTransformFailuresTotal,
	FileTransfersTotal:               FileTransfersTotal,
	FileTransferSizeBytes:            FileTransferSizeBytes,
	FileTransferDuration:             FileTransferDuration,
	IntegrationAuthSuccessTotal:      IntegrationAuthSuccessTotal,
	IntegrationAuthFailuresTotal:     IntegrationAuthFailuresTotal,
	TokenRefreshesTriggeredTotal:     TokenRefreshesTriggeredTotal,
	IntegrationAvailability:          IntegrationAvailability,
	IntegrationLatencyP95:            IntegrationLatencyP95,
	IntegrationErrorsByType:          IntegrationErrorsByType,
	IntegrationRetryableErrors:       IntegrationRetryableErrors,
	IntegrationPermanentFailures:     IntegrationPermanentFailures,
	DataValidationErrors:             DataValidationErrors,
	DataCompletenessScore:            DataCompletenessScore,
	IntegrationConfigChangesTotal:    IntegrationConfigChangesTotal,
	InvalidConfigurations:            InvalidConfigurations,
	IntegrationHealthCheckResults:    IntegrationHealthCheckResults,
}
