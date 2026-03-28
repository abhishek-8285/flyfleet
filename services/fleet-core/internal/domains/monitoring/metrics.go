package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Metrics collection metrics
	MetricsCollectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_metrics_collected_total",
		Help: "Total metrics collected",
	}, []string{"org_id", "metric_type", "source"})

	MetricsScrapeErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_metrics_scrape_errors_total",
		Help: "Metrics scrape/collection errors",
	}, []string{"org_id", "scrape_target", "error_type"})

	ScrapeDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "monitoring_scrape_duration_seconds",
		Help:    "Time to perform metrics scrape",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 6),
	}, []string{"org_id", "scrape_target"})

	// Health checks
	HealthChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_health_checks_total",
		Help: "Health checks performed",
	}, []string{"org_id", "check_type", "target"})

	HealthChecksPassed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_health_checks_passed_total",
		Help: "Health checks passed",
	}, []string{"org_id", "check_type"})

	HealthChecksFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_health_checks_failed_total",
		Help: "Health checks failed",
	}, []string{"org_id", "check_type", "failure_type"})

	HealthCheckLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "monitoring_health_check_latency_seconds",
		Help:    "Health check execution latency",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 6),
	}, []string{"org_id", "check_type"})

	// Synthetic transactions
	SyntheticTransactionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_synthetic_transactions_total",
		Help: "Synthetic transaction tests executed",
	}, []string{"org_id", "transaction_type", "target"})

	SyntheticTransactionSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_synthetic_transactions_success_total",
		Help: "Successful synthetic transactions",
	}, []string{"org_id", "transaction_type"})

	SyntheticTransactionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "monitoring_synthetic_transaction_duration_seconds",
		Help:    "Synthetic transaction execution time",
		Buckets: prometheus.ExponentialBuckets(0.5, 2, 10),
	}, []string{"org_id", "transaction_type"})

	BlackboxProbeSuccess = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_blackbox_probe_success",
		Help: "Blackbox probe success (0=failed, 1=success)",
	}, []string{"org_id", "probe_target", "probe_type"})

	// Log collection
	LogsIngestedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_logs_ingested_total",
		Help: "Total log entries ingested",
	}, []string{"org_id", "log_level", "source"})

	LogsDroppedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_logs_dropped_total",
		Help: "Log entries dropped due to errors",
	}, []string{"org_id", "drop_reason"})

	LogIngestionRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_log_ingestion_rate_logs_per_second",
		Help: "Log ingestion rate",
	}, []string{"org_id"})

	// Tracing
	TracesCollectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_traces_collected_total",
		Help: "Total distributed traces collected",
	}, []string{"org_id", "trace_status"})

	SpansCollectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_spans_collected_total",
		Help: "Total trace spans collected",
	}, []string{"org_id", "span_kind", "service"})

	TraceSamplingRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_trace_sampling_rate_percentage",
		Help: "Trace sampling rate",
	}, []string{"org_id", "sampling_strategy"})

	// Alert pipeline
	AlertPipelineLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "monitoring_alert_pipeline_latency_seconds",
		Help:    "Time from event to alert decision",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 8),
	}, []string{"org_id"})

	AlertmanagerSyncSuccess = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_alertmanager_sync_success",
		Help: "Alertmanager synchronization success",
	}, []string{"org_id"})

	// Dashboards
	DashboardLoadsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_dashboard_loads_total",
		Help: "Dashboard loads",
	}, []string{"org_id", "dashboard_id"})

	DashboardQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "monitoring_dashboard_query_duration_seconds",
		Help:    "Dashboard panel query duration",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 8),
	}, []string{"org_id", "datasource", "panel_id"})

	DashboardRenderLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "monitoring_dashboard_render_latency_seconds",
		Help:    "Dashboard render completion time",
		Buckets: prometheus.ExponentialBuckets(0.5, 2, 6),
	}, []string{"org_id", "dashboard_id"})

	// Recording rules
	RecordingRuleEvaluationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_recording_rule_evaluations_total",
		Help: "Recording rule evaluations",
	}, []string{"org_id", "rule_group"})

	RecordingRuleErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_recording_rule_errors_total",
		Help: "Recording rule evaluation errors",
	}, []string{"org_id", "rule_group", "error_type"})

	// Storage metrics
	MetricsStorageSizeBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_metrics_storage_size_bytes",
		Help: "Time series database size",
	}, []string{"org_id", "storage_type"})

	MetricsRetentionDays = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_metrics_retention_days",
		Help: "Configured metrics retention period",
	}, []string{"org_id", "metric_class"})

	CardinalityMetrics = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_cardinality_metrics_total",
		Help: "Active time series cardinality",
	}, []string{"org_id", "metric_name"})

	HighCardinalityWarnings = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_high_cardinality_warnings_total",
		Help: "High cardinality warnings (>1000 series)",
	}, []string{"org_id", "metric_name"})

	// Remote storage
	RemoteWriteRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_remote_write_requests_total",
		Help: "Remote write requests to long-term storage",
	}, []string{"org_id", "remote_endpoint"})

	RemoteWriteFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_remote_write_failures_total",
		Help: "Remote write failures",
	}, []string{"org_id", "remote_endpoint", "failure_reason"})

	// Query metrics
	PromQLQueriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_promql_queries_total",
		Help: "PromQL queries executed",
	}, []string{"org_id", "query_complexity"})

	PromQLQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "monitoring_promql_query_duration_seconds",
		Help:    "PromQL query execution time",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
	}, []string{"org_id", "query_complexity"})

	PromQLQueryErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_promql_query_errors_total",
		Help: "PromQL query syntax/runtime errors",
	}, []string{"org_id", "error_type"})

	// Exporter metrics
	PrometheusTargetsUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_prometheus_targets_up",
		Help: "Prometheus scrape targets up (1) or down (0)",
	}, []string{"org_id", "job", "instance"})

	PrometheusTargetScrapePool = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_prometheus_scrape_pool_targets",
		Help: "Targets in scrape pool",
	}, []string{"org_id", "job", "state"})

	// Exporter health
	NodeExporterUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_node_exporter_up",
		Help: "Node exporter accessibility",
	}, []string{"org_id", "node"})

	KubeStateMetricsUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_kube_state_metrics_up",
		Help: "Kubernetes state metrics collection",
	}, []string{"org_id"})

	// Custom metrics registration
	CustomMetricsRegistered = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_custom_metrics_registered_total",
		Help: "Custom application metrics registered",
	}, []string{"org_id", "metric_family"})

	// Alertmanager
	AlertmanagerNotificationsSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_alertmanager_notifications_sent_total",
		Help: "Notifications sent by Alertmanager",
	}, []string{"org_id", "receiver", "integration"})

	AlertmanagerNotificationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_alertmanager_notification_errors_total",
		Help: "Alertmanager notification errors",
	}, []string{"org_id", "receiver", "error_type"})

	// Grafana
	GrafanaDashboardLoadsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_grafana_dashboard_loads_total",
		Help: "Grafana dashboard views",
	}, []string{"org_id", "dashboard_id", "user_role"})

	GrafanaAlertRuleEvaluations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_grafana_alert_rule_evaluations_total",
		Help: "Grafana dashboard alert rule evaluations",
	}, []string{"org_id", "dashboard_id", "rule_id"})

	// Anomaly detection
	AnomalyDetectionScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_anomaly_detection_score",
		Help: "Anomaly detection confidence score (0-1)",
	}, []string{"org_id", "metric_name", "algorithm"})

	AnomaliesDetectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_anomalies_detected_total",
		Help: "Anomalies detected by ML/statistical models",
	}, []string{"org_id", "anomaly_type", "severity"})

	// Capacity forecasting
	CapacityForecastDays = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_capacity_forecast_days",
		Help: "Days until capacity exhaustion forecast",
	}, []string{"org_id", "resource_type"})

	// Configuration drift
	ConfigMapVersion = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_configmap_version",
		Help: "Configuration version hash",
	}, []string{"org_id", "config_name"})

	ConfigurationDriftDetected = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "monitoring_configuration_drift_detected_total",
		Help: "Configuration drift from expected state",
	}, []string{"org_id", "config_name", "drift_type"})

	// Cost metrics
	MonitoringCostUSD = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_cost_usd",
		Help: "Monitoring infrastructure cost in USD",
	}, []string{"org_id", "cost_category"})

	CostPerMetricSeries = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monitoring_cost_per_metric_series_usd",
		Help: "Cost per active time series per month",
	}, []string{"org_id"})
)

// Global metrics instance
var GlobalMetrics = struct {
	MetricsCollectedTotal          *prometheus.CounterVec
	MetricsScrapeErrorsTotal       *prometheus.CounterVec
	ScrapeDuration                 *prometheus.HistogramVec
	HealthChecksTotal              *prometheus.CounterVec
	HealthChecksPassed             *prometheus.CounterVec
	HealthChecksFailed             *prometheus.CounterVec
	HealthCheckLatency             *prometheus.HistogramVec
	SyntheticTransactionsTotal     *prometheus.CounterVec
	SyntheticTransactionSuccess    *prometheus.CounterVec
	SyntheticTransactionDuration   *prometheus.HistogramVec
	BlackboxProbeSuccess           *prometheus.GaugeVec
	LogsIngestedTotal              *prometheus.CounterVec
	LogsDroppedTotal               *prometheus.CounterVec
	LogIngestionRate               *prometheus.GaugeVec
	TracesCollectedTotal           *prometheus.CounterVec
	SpansCollectedTotal            *prometheus.CounterVec
	TraceSamplingRate              *prometheus.GaugeVec
	AlertPipelineLatency           *prometheus.HistogramVec
	AlertmanagerSyncSuccess        *prometheus.GaugeVec
	DashboardLoadsTotal            *prometheus.CounterVec
	DashboardQueryDuration         *prometheus.HistogramVec
	DashboardRenderLatency         *prometheus.HistogramVec
	RecordingRuleEvaluationsTotal  *prometheus.CounterVec
	RecordingRuleErrorsTotal       *prometheus.CounterVec
	MetricsStorageSizeBytes        *prometheus.GaugeVec
	MetricsRetentionDays           *prometheus.GaugeVec
	CardinalityMetrics             *prometheus.GaugeVec
	HighCardinalityWarnings        *prometheus.CounterVec
	RemoteWriteRequestsTotal       *prometheus.CounterVec
	RemoteWriteFailuresTotal       *prometheus.CounterVec
	PromQLQueriesTotal             *prometheus.CounterVec
	PromQLQueryDuration            *prometheus.HistogramVec
	PromQLQueryErrorsTotal         *prometheus.CounterVec
	PrometheusTargetsUp            *prometheus.GaugeVec
	PrometheusTargetScrapePool     *prometheus.GaugeVec
	NodeExporterUp                 *prometheus.GaugeVec
	KubeStateMetricsUp             *prometheus.GaugeVec
	CustomMetricsRegistered        *prometheus.GaugeVec
	AlertmanagerNotificationsSent  *prometheus.CounterVec
	AlertmanagerNotificationErrors *prometheus.CounterVec
	GrafanaDashboardLoadsTotal     *prometheus.CounterVec
	GrafanaAlertRuleEvaluations    *prometheus.CounterVec
	AnomalyDetectionScore          *prometheus.GaugeVec
	AnomaliesDetectedTotal         *prometheus.CounterVec
	CapacityForecastDays           *prometheus.GaugeVec
	ConfigMapVersion               *prometheus.GaugeVec
	ConfigurationDriftDetected     *prometheus.CounterVec
	MonitoringCostUSD              *prometheus.GaugeVec
	CostPerMetricSeries            *prometheus.GaugeVec
}{
	MetricsCollectedTotal:          MetricsCollectedTotal,
	MetricsScrapeErrorsTotal:       MetricsScrapeErrorsTotal,
	ScrapeDuration:                 ScrapeDuration,
	HealthChecksTotal:              HealthChecksTotal,
	HealthChecksPassed:             HealthChecksPassed,
	HealthChecksFailed:             HealthChecksFailed,
	HealthCheckLatency:             HealthCheckLatency,
	SyntheticTransactionsTotal:     SyntheticTransactionsTotal,
	SyntheticTransactionSuccess:    SyntheticTransactionSuccess,
	SyntheticTransactionDuration:   SyntheticTransactionDuration,
	BlackboxProbeSuccess:           BlackboxProbeSuccess,
	LogsIngestedTotal:              LogsIngestedTotal,
	LogsDroppedTotal:               LogsDroppedTotal,
	LogIngestionRate:               LogIngestionRate,
	TracesCollectedTotal:           TracesCollectedTotal,
	SpansCollectedTotal:            SpansCollectedTotal,
	TraceSamplingRate:              TraceSamplingRate,
	AlertPipelineLatency:           AlertPipelineLatency,
	AlertmanagerSyncSuccess:        AlertmanagerSyncSuccess,
	DashboardLoadsTotal:            DashboardLoadsTotal,
	DashboardQueryDuration:         DashboardQueryDuration,
	DashboardRenderLatency:         DashboardRenderLatency,
	RecordingRuleEvaluationsTotal:  RecordingRuleEvaluationsTotal,
	RecordingRuleErrorsTotal:       RecordingRuleErrorsTotal,
	MetricsStorageSizeBytes:        MetricsStorageSizeBytes,
	MetricsRetentionDays:           MetricsRetentionDays,
	CardinalityMetrics:             CardinalityMetrics,
	HighCardinalityWarnings:        HighCardinalityWarnings,
	RemoteWriteRequestsTotal:       RemoteWriteRequestsTotal,
	RemoteWriteFailuresTotal:       RemoteWriteFailuresTotal,
	PromQLQueriesTotal:             PromQLQueriesTotal,
	PromQLQueryDuration:            PromQLQueryDuration,
	PromQLQueryErrorsTotal:         PromQLQueryErrorsTotal,
	PrometheusTargetsUp:            PrometheusTargetsUp,
	PrometheusTargetScrapePool:     PrometheusTargetScrapePool,
	NodeExporterUp:                 NodeExporterUp,
	KubeStateMetricsUp:             KubeStateMetricsUp,
	CustomMetricsRegistered:        CustomMetricsRegistered,
	AlertmanagerNotificationsSent:  AlertmanagerNotificationsSent,
	AlertmanagerNotificationErrors: AlertmanagerNotificationErrors,
	GrafanaDashboardLoadsTotal:     GrafanaDashboardLoadsTotal,
	GrafanaAlertRuleEvaluations:    GrafanaAlertRuleEvaluations,
	AnomalyDetectionScore:          AnomalyDetectionScore,
	AnomaliesDetectedTotal:         AnomaliesDetectedTotal,
	CapacityForecastDays:           CapacityForecastDays,
	ConfigMapVersion:               ConfigMapVersion,
	ConfigurationDriftDetected:     ConfigurationDriftDetected,
	MonitoringCostUSD:              MonitoringCostUSD,
	CostPerMetricSeries:            CostPerMetricSeries,
}
