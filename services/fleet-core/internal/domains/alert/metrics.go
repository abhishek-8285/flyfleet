package alert

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Alert generation metrics
	AlertsGeneratedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_generated_total",
		Help: "Total alerts generated",
	}, []string{"org_id", "alert_type", "severity"})

	AlertEvaluationCyclesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "alert_evaluation_cycles_total",
		Help: "Total alert rule evaluation cycles",
	})

	AlertRulesEvaluatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_rules_evaluated_total",
		Help: "Total alert rules evaluated",
	}, []string{"org_id", "rule_type"})

	// Alert delivery metrics
	AlertsSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_sent_total",
		Help: "Total alerts delivered to notification channels",
	}, []string{"org_id", "channel", "severity"})

	AlertDeliveryFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_delivery_failures_total",
		Help: "Failed alert deliveries",
	}, []string{"org_id", "channel", "failure_reason"})

	AlertDeliveryLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "alert_delivery_latency_seconds",
		Help:    "Time from alert generation to delivery",
		Buckets: prometheus.ExponentialBuckets(0.5, 2, 8),
	}, []string{"org_id", "channel"})

	// Alert group metrics
	AlertGroupsCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_groups_created_total",
		Help: "Alert groups created (deduplication)",
	}, []string{"org_id", "grouping_key"})

	AlertGroupSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "alert_group_size",
		Help:    "Number of alerts per group",
		Buckets: prometheus.LinearBuckets(1, 5, 20),
	}, []string{"org_id"})

	// Alert lifecycle
	AlertsAcknowledgedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_acknowledged_total",
		Help: "Alerts acknowledged by operators",
	}, []string{"org_id", "acknowledged_by_role"})

	AlertsResolvedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_resolved_total",
		Help: "Alerts marked as resolved",
	}, []string{"org_id", "resolution_type"})

	AlertTimeToAcknowledge = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "alert_time_to_acknowledge_seconds",
		Help:    "Time from alert to acknowledgment",
		Buckets: prometheus.ExponentialBuckets(60, 2, 10),
	}, []string{"org_id", "severity"})

	AlertTimeToResolve = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "alert_time_to_resolve_seconds",
		Help:    "Time from alert to resolution",
		Buckets: prometheus.ExponentialBuckets(300, 2, 12),
	}, []string{"org_id", "severity", "alert_type"})

	AlertEscalationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_escalations_total",
		Help: "Alert escalation events",
	}, []string{"org_id", "escalation_level", "escalated_to_role"})

	// Suppression metrics
	AlertSuppressionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_suppressions_total",
		Help: "Alerts suppressed (maintenance windows, dependencies)",
	}, []string{"org_id", "suppression_reason"})

	SilenceActiveTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alert_silence_active_total",
		Help: "Currently active alert silences",
	}, []string{"org_id", "silence_reason"})

	// Noise metrics
	AlertNoiseScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alert_noise_score",
		Help: "Alert noise score (low=good, high=too many false positives)",
	}, []string{"org_id", "alert_type"})

	BouncingAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_bouncing_total",
		Help: "Bouncing alerts (fire, resolve, fire again quickly)",
	}, []string{"org_id", "alert_type"})

	FlappingAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_flapping_total",
		Help: "Flapping alerts (rapid state changes)",
	}, []string{"org_id", "alert_type"})

	// Alert rule metrics
	AlertRuleErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_rule_errors_total",
		Help: "Alert rule evaluation errors",
	}, []string{"org_id", "error_type", "rule_id"})

	AlertRuleCompilationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "alert_rule_compilation_duration_seconds",
		Help:    "Time to compile alert rules",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 6),
	}, []string{"org_id"})

	AlertRuleMemoryUsage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alert_rule_memory_usage_bytes",
		Help: "Alert rule evaluation engine memory usage",
	}, []string{"org_id"})

	// SLO-based alerts
	SLOBurnRateAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_slo_burn_rate_total",
		Help: "SLO burn rate alerts triggered",
	}, []string{"org_id", "slo_name", "window"})

	ErrorBudgetRemainingPercentage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alert_error_budget_remaining_percentage",
		Help: "Remaining error budget percentage",
	}, []string{"org_id", "slo_name"})

	// Multi-condition alerts
	CompositeAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_composite_total",
		Help: "Composite alerts (multiple conditions)",
	}, []string{"org_id", "composite_type"})

	// Alert enrichment
	AlertEnrichmentAppliedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_enrichment_applied_total",
		Help: "Enrichment data added to alerts",
	}, []string{"org_id", "enrichment_type"})

	// Runbook integration
	RunbookURLAttachedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_runbook_url_attached_total",
		Help: "Alerts with runbook URL attached",
	}, []string{"org_id"})

	RunbookFollowedRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alert_runbook_followed_rate_percentage",
		Help: "Percentage of resolved alerts where runbook was followed",
	}, []string{"org_id", "alert_type"})

	// Triage metrics
	AlertsNeedingTriage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_needing_triage_total",
		Help: "Alerts requiring human triage",
	}, []string{"org_id", "priority"})

	TriageSLACompliance = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alert_triage_sla_compliance_percentage",
		Help: "Triage SLA compliance percentage",
	}, []string{"org_id"})

	// Duplicate detection
	DuplicateAlertsSuppressed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_duplicate_suppressed_total",
		Help: "Duplicate alerts suppressed",
	}, []string{"org_id", "deduplication_type"})

	// Muting metrics
	MuteRulesActiveTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alert_mute_rules_active_total",
		Help: "Active mute/suppression rules",
	}, []string{"org_id"})

	MuteHitCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_mute_hit_total",
		Help: "Alerts suppressed by mute rules",
	}, []string{"org_id", "mute_rule_id"})

	// Notification rate limiting
	NotificationRateLimitHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_notification_rate_limit_hits_total",
		Help: "Notification rate limit enforcement hits",
	}, []string{"org_id", "channel"})

	// Alert correlation
	CorrelatedAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_correlated_total",
		Help: "Alerts correlated into incidents",
	}, []string{"org_id", "correlation_rule"})

	IncidentsCreatedFromAlerts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "incident_from_alert_total",
		Help: "Incidents created from alert patterns",
	}, []string{"org_id", "incident_type"})

	// Alert feedback
	AlertFeedbackTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alert_feedback_total",
		Help: "Operator feedback on alert quality",
	}, []string{"org_id", "feedback_type"}) // useful, noisy, duplicate
)

// Global metrics instance
var GlobalMetrics = struct {
	AlertsGeneratedTotal           *prometheus.CounterVec
	AlertEvaluationCyclesTotal     prometheus.Counter
	AlertRulesEvaluatedTotal       *prometheus.CounterVec
	AlertsSentTotal                *prometheus.CounterVec
	AlertDeliveryFailuresTotal     *prometheus.CounterVec
	AlertDeliveryLatency           *prometheus.HistogramVec
	AlertGroupsCreatedTotal        *prometheus.CounterVec
	AlertGroupSize                 *prometheus.HistogramVec
	AlertsAcknowledgedTotal        *prometheus.CounterVec
	AlertsResolvedTotal            *prometheus.CounterVec
	AlertTimeToAcknowledge         *prometheus.HistogramVec
	AlertTimeToResolve             *prometheus.HistogramVec
	AlertEscalationsTotal          *prometheus.CounterVec
	AlertSuppressionsTotal         *prometheus.CounterVec
	SilenceActiveTotal             *prometheus.GaugeVec
	AlertNoiseScore                *prometheus.GaugeVec
	BouncingAlertsTotal            *prometheus.CounterVec
	FlappingAlertsTotal            *prometheus.CounterVec
	AlertRuleErrorsTotal           *prometheus.CounterVec
	AlertRuleCompilationDuration   *prometheus.HistogramVec
	AlertRuleMemoryUsage           *prometheus.GaugeVec
	SLOBurnRateAlertsTotal         *prometheus.CounterVec
	ErrorBudgetRemainingPercentage *prometheus.GaugeVec
	CompositeAlertsTotal           *prometheus.CounterVec
	AlertEnrichmentAppliedTotal    *prometheus.CounterVec
	RunbookURLAttachedTotal        *prometheus.CounterVec
	RunbookFollowedRate            *prometheus.GaugeVec
	AlertsNeedingTriage            *prometheus.GaugeVec
	TriageSLACompliance            *prometheus.GaugeVec
	DuplicateAlertsSuppressed      *prometheus.CounterVec
	MuteRulesActiveTotal           *prometheus.GaugeVec
	MuteHitCount                   *prometheus.CounterVec
	NotificationRateLimitHits      *prometheus.CounterVec
	CorrelatedAlertsTotal          *prometheus.CounterVec
	IncidentsCreatedFromAlerts     *prometheus.CounterVec
	AlertFeedbackTotal             *prometheus.CounterVec
}{
	AlertsGeneratedTotal:           AlertsGeneratedTotal,
	AlertEvaluationCyclesTotal:     AlertEvaluationCyclesTotal,
	AlertRulesEvaluatedTotal:       AlertRulesEvaluatedTotal,
	AlertsSentTotal:                AlertsSentTotal,
	AlertDeliveryFailuresTotal:     AlertDeliveryFailuresTotal,
	AlertDeliveryLatency:           AlertDeliveryLatency,
	AlertGroupsCreatedTotal:        AlertGroupsCreatedTotal,
	AlertGroupSize:                 AlertGroupSize,
	AlertsAcknowledgedTotal:        AlertsAcknowledgedTotal,
	AlertsResolvedTotal:            AlertsResolvedTotal,
	AlertTimeToAcknowledge:         AlertTimeToAcknowledge,
	AlertTimeToResolve:             AlertTimeToResolve,
	AlertEscalationsTotal:          AlertEscalationsTotal,
	AlertSuppressionsTotal:         AlertSuppressionsTotal,
	SilenceActiveTotal:             SilenceActiveTotal,
	AlertNoiseScore:                AlertNoiseScore,
	BouncingAlertsTotal:            BouncingAlertsTotal,
	FlappingAlertsTotal:            FlappingAlertsTotal,
	AlertRuleErrorsTotal:           AlertRuleErrorsTotal,
	AlertRuleCompilationDuration:   AlertRuleCompilationDuration,
	AlertRuleMemoryUsage:           AlertRuleMemoryUsage,
	SLOBurnRateAlertsTotal:         SLOBurnRateAlertsTotal,
	ErrorBudgetRemainingPercentage: ErrorBudgetRemainingPercentage,
	CompositeAlertsTotal:           CompositeAlertsTotal,
	AlertEnrichmentAppliedTotal:    AlertEnrichmentAppliedTotal,
	RunbookURLAttachedTotal:        RunbookURLAttachedTotal,
	RunbookFollowedRate:            RunbookFollowedRate,
	AlertsNeedingTriage:            AlertsNeedingTriage,
	TriageSLACompliance:            TriageSLACompliance,
	DuplicateAlertsSuppressed:      DuplicateAlertsSuppressed,
	MuteRulesActiveTotal:           MuteRulesActiveTotal,
	MuteHitCount:                   MuteHitCount,
	NotificationRateLimitHits:      NotificationRateLimitHits,
	CorrelatedAlertsTotal:          CorrelatedAlertsTotal,
	IncidentsCreatedFromAlerts:     IncidentsCreatedFromAlerts,
	AlertFeedbackTotal:             AlertFeedbackTotal,
}
