package reports

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Report generation metrics
	ReportsRequestedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_requested_total",
		Help: "Total report generation requests",
	}, []string{"tenant_id", "report_type", "format"})

	ReportsGeneratedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_generated_total",
		Help: "Total reports successfully generated",
	}, []string{"tenant_id", "report_type", "format"})

	ReportsFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_failed_total",
		Help: "Total report generation failures",
	}, []string{"tenant_id", "report_type", "failure_reason"})

	ReportGenerationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "reports_generation_duration_seconds",
		Help:    "Report generation time",
		Buckets: prometheus.ExponentialBuckets(1, 2, 12),
	}, []string{"tenant_id", "report_type"})

	ReportSizeBytes = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "reports_size_bytes",
		Help:    "Generated report file size",
		Buckets: prometheus.ExponentialBuckets(1024, 2, 12),
	}, []string{"tenant_id", "report_type", "format"})

	// Scheduled reports
	ScheduledReportsConfiguredTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "reports_scheduled_configured_total",
		Help: "Configured scheduled reports",
	}, []string{"tenant_id", "frequency"})

	ScheduledReportsExecutedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_scheduled_executed_total",
		Help: "Scheduled reports executed",
	}, []string{"tenant_id", "frequency"})

	ScheduledReportsDeliveredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_scheduled_delivered_total",
		Help: "Scheduled reports successfully delivered",
	}, []string{"tenant_id", "delivery_method"})

	ScheduledReportsMissedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_scheduled_missed_total",
		Help: "Scheduled reports that failed to deliver",
	}, []string{"tenant_id", "reason"})

	// Data export metrics
	DataExportsRequestedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_data_exports_requested_total",
		Help: "Custom data export requests",
	}, []string{"tenant_id", "export_type", "format"})

	DataExportsDeliveredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_data_exports_delivered_total",
		Help: "Data exports delivered",
	}, []string{"tenant_id"})

	DataExportSizeGB = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "reports_data_export_size_gb",
		Help:    "Data export size in GB",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
	}, []string{"tenant_id", "export_type"})

	// Template metrics
	ReportTemplatesActiveTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "reports_templates_active_total",
		Help: "Active report templates",
	})

	ReportTemplatesCustom = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "reports_templates_custom_total",
		Help: "Custom report templates per organization",
	}, []string{"tenant_id"})

	TemplateUsageCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_template_usage_total",
		Help: "Template usage count",
	}, []string{"template_id"})

	// Delivery metrics
	ReportEmailsSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_emails_sent_total",
		Help: "Reports delivered via email",
	}, []string{"tenant_id", "report_type"})

	ReportUploadsToS3Total = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_uploads_to_s3_total",
		Help: "Reports uploaded to S3/MinIO",
	}, []string{"tenant_id", "bucket"})

	ReportAPIDownloadsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_api_downloads_total",
		Help: "Reports downloaded via API",
	}, []string{"tenant_id"})

	ReportDeliveryFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_delivery_failures_total",
		Help: "Report delivery failures",
	}, []string{"tenant_id", "delivery_method", "failure_reason"})

	// Caching metrics
	ReportCacheHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_cache_hits_total",
		Help: "Report cache hits",
	}, []string{"tenant_id", "cache_type"})

	ReportCacheMissesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_cache_misses_total",
		Help: "Report cache misses",
	}, []string{"tenant_id", "cache_type"})

	// Query metrics
	ReportQueryExecutionTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "reports_query_execution_time_seconds",
		Help:    "Report query execution time",
		Buckets: prometheus.ExponentialBuckets(1, 2, 12),
	}, []string{"tenant_id", "query_complexity"})

	ReportQueryRowsReturned = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "reports_query_rows_returned",
		Help:    "Number of rows returned by report queries",
		Buckets: prometheus.ExponentialBuckets(100, 2, 12),
	}, []string{"tenant_id", "report_type"})

	// Scheduling metrics
	ScheduledReportTriggersTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_scheduled_triggers_total",
		Help: "Scheduled report trigger evaluations",
	}, []string{"tenant_id"})

	ScheduledReportDueTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "reports_scheduled_due_total",
		Help: "Reports due for generation",
	}, []string{"tenant_id", "frequency"})

	// Data freshness metrics
	ReportDataFreshnessHours = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "reports_data_freshness_hours",
		Help: "Report data age in hours (time since data cutoff)",
	}, []string{"tenant_id", "report_type"})

	StaleDataAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_stale_data_alerts_total",
		Help: "Alerts triggered for stale report data",
	}, []string{"tenant_id"})

	// Permissions metrics
	ReportAccessDeniedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_access_denied_total",
		Help: "Report access permission denied errors",
	}, []string{"tenant_id", "user_role", "report_type"})

	// PDF/Excel generation metrics
	PDFGenerationTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "reports_pdf_generation_time_seconds",
		Help:    "PDF report generation time",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10),
	}, []string{"tenant_id"})

	ExcelGenerationTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "reports_excel_generation_time_seconds",
		Help:    "Excel report generation time",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10),
	}, []string{"tenant_id"})

	// DRM/security metrics
	ReportDownloadsUnauthorizedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_unauthorized_downloads_total",
		Help: "Unauthorized report download attempts",
	}, []string{"tenant_id", "report_type"})

	ReportLinkExpiryViolation = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_link_expiry_violations_total",
		Help: "Report link used after expiry",
	}, []string{"tenant_id"})

	// Performance metrics
	ReportConcurrentGeneration = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "reports_concurrent_generation_count",
		Help: "Currently running report generations",
	}, []string{"tenant_id"})

	ReportQueueDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "reports_queue_depth",
		Help: "Report generation queue depth",
	}, []string{"tenant_id", "priority"})

	ReportGenerationTimeout = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reports_generation_timeouts_total",
		Help: "Report generation timeouts",
	}, []string{"tenant_id", "report_type"})
)

// Global metrics instance
var GlobalMetrics = struct {
	ReportsRequestedTotal            *prometheus.CounterVec
	ReportsGeneratedTotal            *prometheus.CounterVec
	ReportsFailedTotal               *prometheus.CounterVec
	ReportGenerationDuration         *prometheus.HistogramVec
	ReportSizeBytes                  *prometheus.HistogramVec
	ScheduledReportsConfiguredTotal  *prometheus.GaugeVec
	ScheduledReportsExecutedTotal    *prometheus.CounterVec
	ScheduledReportsDeliveredTotal   *prometheus.CounterVec
	ScheduledReportsMissedTotal      *prometheus.CounterVec
	DataExportsRequestedTotal        *prometheus.CounterVec
	DataExportsDeliveredTotal        *prometheus.CounterVec
	DataExportSizeGB                 *prometheus.HistogramVec
	ReportTemplatesActiveTotal       prometheus.Gauge
	ReportTemplatesCustom            *prometheus.GaugeVec
	TemplateUsageCount               *prometheus.CounterVec
	ReportEmailsSentTotal            *prometheus.CounterVec
	ReportUploadsToS3Total           *prometheus.CounterVec
	ReportAPIDownloadsTotal          *prometheus.CounterVec
	ReportDeliveryFailures           *prometheus.CounterVec
	ReportCacheHitsTotal             *prometheus.CounterVec
	ReportCacheMissesTotal           *prometheus.CounterVec
	ReportQueryExecutionTime         *prometheus.HistogramVec
	ReportQueryRowsReturned          *prometheus.HistogramVec
	ScheduledReportTriggersTotal     *prometheus.CounterVec
	ScheduledReportDueTotal          *prometheus.GaugeVec
	ReportDataFreshnessHours         *prometheus.GaugeVec
	StaleDataAlertsTotal             *prometheus.CounterVec
	ReportAccessDeniedTotal          *prometheus.CounterVec
	PDFGenerationTime                *prometheus.HistogramVec
	ExcelGenerationTime              *prometheus.HistogramVec
	ReportDownloadsUnauthorizedTotal *prometheus.CounterVec
	ReportLinkExpiryViolation        *prometheus.CounterVec
	ReportConcurrentGeneration       *prometheus.GaugeVec
	ReportQueueDepth                 *prometheus.GaugeVec
	ReportGenerationTimeout          *prometheus.CounterVec
}{
	ReportsRequestedTotal:            ReportsRequestedTotal,
	ReportsGeneratedTotal:            ReportsGeneratedTotal,
	ReportsFailedTotal:               ReportsFailedTotal,
	ReportGenerationDuration:         ReportGenerationDuration,
	ReportSizeBytes:                  ReportSizeBytes,
	ScheduledReportsConfiguredTotal:  ScheduledReportsConfiguredTotal,
	ScheduledReportsExecutedTotal:    ScheduledReportsExecutedTotal,
	ScheduledReportsDeliveredTotal:   ScheduledReportsDeliveredTotal,
	ScheduledReportsMissedTotal:      ScheduledReportsMissedTotal,
	DataExportsRequestedTotal:        DataExportsRequestedTotal,
	DataExportsDeliveredTotal:        DataExportsDeliveredTotal,
	DataExportSizeGB:                 DataExportSizeGB,
	ReportTemplatesActiveTotal:       ReportTemplatesActiveTotal,
	ReportTemplatesCustom:            ReportTemplatesCustom,
	TemplateUsageCount:               TemplateUsageCount,
	ReportEmailsSentTotal:            ReportEmailsSentTotal,
	ReportUploadsToS3Total:           ReportUploadsToS3Total,
	ReportAPIDownloadsTotal:          ReportAPIDownloadsTotal,
	ReportDeliveryFailures:           ReportDeliveryFailures,
	ReportCacheHitsTotal:             ReportCacheHitsTotal,
	ReportCacheMissesTotal:           ReportCacheMissesTotal,
	ReportQueryExecutionTime:         ReportQueryExecutionTime,
	ReportQueryRowsReturned:          ReportQueryRowsReturned,
	ScheduledReportTriggersTotal:     ScheduledReportTriggersTotal,
	ScheduledReportDueTotal:          ScheduledReportDueTotal,
	ReportDataFreshnessHours:         ReportDataFreshnessHours,
	StaleDataAlertsTotal:             StaleDataAlertsTotal,
	ReportAccessDeniedTotal:          ReportAccessDeniedTotal,
	PDFGenerationTime:                PDFGenerationTime,
	ExcelGenerationTime:              ExcelGenerationTime,
	ReportDownloadsUnauthorizedTotal: ReportDownloadsUnauthorizedTotal,
	ReportLinkExpiryViolation:        ReportLinkExpiryViolation,
	ReportConcurrentGeneration:       ReportConcurrentGeneration,
	ReportQueueDepth:                 ReportQueueDepth,
	ReportGenerationTimeout:          ReportGenerationTimeout,
}
