// Package alerts implements metrics for the alerts domain
// This domain handles alert rule management, templates, and routing

package alerts

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Alert rule management metrics
	AlertRulesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_rules_total",
		Help: "Total alert rules configured",
	}, []string{"org_id", "rule_type", "status"})

	AlertRuleCreationTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_rule_creation_total",
		Help: "Alert rules created",
	}, []string{"org_id", "rule_type"})

	AlertRuleUpdatesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_rule_updates_total",
		Help: "Alert rule modifications",
	}, []string{"org_id", "update_type"})

	AlertRuleDeletionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_rule_deletions_total",
		Help: "Alert rules deleted",
	}, []string{"org_id"})

	AlertRuleValidationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_rule_validation_errors_total",
		Help: "Alert rule syntax/semantic validation errors",
	}, []string{"org_id", "error_type", "rule_category"})

	// Alert templates
	AlertTemplatesTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "alerts_templates_total",
		Help: "Total alert templates available",
	})

	AlertTemplatesUsed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_templates_used_total",
		Help: "Alert template usage count",
	}, []string{"template_id", "alert_type"})

	TemplateCustomizationRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_template_customization_rate_percentage",
		Help: "Percentage of rules using custom templates",
	}, []string{"org_id"})

	// Routing metrics
	AlertRoutingRulesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_routing_rules_total",
		Help: "Configured alert routing rules",
	}, []string{"org_id", "route_type"})

	AlertRoutingEvaluationTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "alerts_routing_evaluation_time_ms",
		Help:    "Alert routing rule evaluation latency",
		Buckets: prometheus.LinearBuckets(1, 5, 20),
	}, []string{"org_id"})

	AlertRoutingMatchesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_routing_matches_total",
		Help: "Alerts matched to routing rules",
	}, []string{"org_id", "route_id"})

	// Notification channels
	NotificationChannelsConfigured = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_notification_channels_configured_total",
		Help: "Configured notification channels",
	}, []string{"org_id", "channel_type"})

	NotificationChannelHealth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_notification_channel_health",
		Help: "Notification channel health status (0-1)",
	}, []string{"org_id", "channel_type"})

	// Alert deduplication
	DeduplicationRulesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_deduplication_rules_total",
		Help: "Alert deduplication rules configured",
	}, []string{"org_id"})

	DuplicateAlertsDetected = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_duplicate_detected_total",
		Help: "Duplicates identified and consolidated",
	}, []string{"org_id", "deduplication_type"})

	// Alert silencing
	SilenceRulesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_silence_rules_total",
		Help: "Active silence rules",
	}, []string{"org_id", "silence_type"})

	SilenceMatchesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_silence_matches_total",
		Help: "Alerts suppressed by silence rules",
	}, []string{"org_id"})

	// Alert inhibition
	InhibitionRulesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_inhibition_rules_total",
		Help: "Inhibition rules (suppress minor when major)",
	}, []string{"org_id"})

	InhibitedAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_inhibited_total",
		Help: "Alerts inhibited by rules",
	}, []string{"org_id", "inhibition_type"})

	// Grouping
	AlertGroupingRulesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_grouping_rules_total",
		Help: "Alert grouping rules",
	}, []string{"org_id", "group_by"})

	AlertGroupsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_groups_created_total",
		Help: "Alert groups formed",
	}, []string{"org_id"})

	// Webhook metrics
	AlertWebhooksTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_webhooks_total",
		Help: "Configured alert webhooks",
	}, []string{"org_id", "webhook_type"})

	AlertWebhookDeliveries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_webhook_deliveries_total",
		Help: "Alert webhook calls",
	}, []string{"org_id", "webhook_url", "result"})

	AlertWebhookLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "alerts_webhook_latency_seconds",
		Help:    "Alert webhook response time",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 8),
	}, []string{"org_id", "webhook_type"})

	// Escalation policies
	EscalationPoliciesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_escalation_policies_total",
		Help: "Escalation policies configured",
	}, []string{"org_id", "policy_type"})

	EscalationsTriggered = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_escalations_triggered_total",
		Help: "Alert escalations executed",
	}, []string{"org_id", "escalation_level"})

	// SNOOZE metrics
	SnoozeRulesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_snooze_rules_total",
		Help: "Active snooze rules",
	}, []string{"org_id"})

	SnoozedAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_snoozed_total",
		Help: "Alerts snoozed",
	}, []string{"org_id"})

	// Query-based alerts
	QueryBasedAlertsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_query_based_total",
		Help: "Query-based alert rules",
	}, []string{"org_id"})

	QueryBasedAlertEvaluations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_query_based_evaluations_total",
		Help: "Query-based alert rule evaluations",
	}, []string{"org_id", "query_complexity"})

	QueryBasedAlertErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_query_based_errors_total",
		Help: "Query-based alert evaluation errors",
	}, []string{"org_id", "error_type"})

	// Metric threshold alerts
	MetricThresholdAlertsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_metric_threshold_total",
		Help: "Metric threshold alert rules",
	}, []string{"org_id", "threshold_type"})

	// Log-based alerts
	LogBasedAlertsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_log_based_total",
		Help: "Log-based alert rules",
	}, []string{"org_id", "log_source"})

	LogBasedAlertMatches = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_log_based_matches_total",
		Help: "Log patterns matching alert rules",
	}, []string{"org_id", "rule_id"})

	// Trace-based alerts
	TraceBasedAlertsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_trace_based_total",
		Help: "Distributed trace-based alert rules",
	}, []string{"org_id"})

	TraceBasedAlertMatches = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_trace_based_matches_total",
		Help: "Trace patterns matching alert rules",
	}, []string{"org_id"})

	// Profile alerts
	ProfileAlertsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_profile_total",
		Help: "Continuous profiling alert rules",
	}, []string{"org_id", "profile_type"})

	// Heatmap alerts
	HeatmapAlertsTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_heatmap_total",
		Help: "Heatmap-based outlier detection alerts",
	}, []string{"org_id"})

	// Alert API metrics
	AlertAPIRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_api_requests_total",
		Help: "Alert API requests",
	}, []string{"org_id", "endpoint", "method"})

	AlertAPIRequestErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_api_request_errors_total",
		Help: "Alert API request errors",
	}, []string{"org_id", "endpoint", "error_type"})

	// Alert state history
	AlertStateChangesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_state_changes_total",
		Help: "Alert state transitions",
	}, []string{"org_id", "from_state", "to_state"})

	// Peak alert volume
	PeakAlertRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_peak_rate_per_minute",
		Help: "Peak alerts per minute",
	}, []string{"org_id"})

	// Alert fatigue metrics
	AlertFatigueScore = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_fatigue_score",
		Help: "Alert fatigue score (higher = more noise)",
	}, []string{"org_id", "user_id"})

	AlertsPerOperator = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alerts_per_operator",
		Help: "Alerts assigned per operator",
	}, []string{"org_id", "operator_id"})

	// Resolution metrics
	MeanTimeToResolutionByRule = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "alerts_mttr_by_rule_seconds",
		Help:    "MTTR per alert rule",
		Buckets: prometheus.ExponentialBuckets(300, 2, 12),
	}, []string{"org_id", "rule_id"})

	RecurringAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_recurring_total",
		Help: "Recurring same alert",
	}, []string{"org_id", "alert_type", "recurrence_period"})
)

// Global metrics instance
var GlobalMetrics = struct {
	AlertRulesTotal                *prometheus.GaugeVec
	AlertRuleCreationTotal         *prometheus.CounterVec
	AlertRuleUpdatesTotal          *prometheus.CounterVec
	AlertRuleDeletionsTotal        *prometheus.CounterVec
	AlertRuleValidationErrors      *prometheus.CounterVec
	AlertTemplatesTotal            prometheus.Gauge
	AlertTemplatesUsed             *prometheus.CounterVec
	TemplateCustomizationRate      *prometheus.GaugeVec
	AlertRoutingRulesTotal         *prometheus.GaugeVec
	AlertRoutingEvaluationTime     *prometheus.HistogramVec
	AlertRoutingMatchesTotal       *prometheus.CounterVec
	NotificationChannelsConfigured *prometheus.GaugeVec
	NotificationChannelHealth      *prometheus.GaugeVec
	DeduplicationRulesTotal        *prometheus.GaugeVec
	DuplicateAlertsDetected        *prometheus.CounterVec
	SilenceRulesTotal              *prometheus.GaugeVec
	SilenceMatchesTotal            *prometheus.CounterVec
	InhibitionRulesTotal           *prometheus.GaugeVec
	InhibitedAlertsTotal           *prometheus.CounterVec
	AlertGroupingRulesTotal        *prometheus.GaugeVec
	AlertGroupsCreated             *prometheus.CounterVec
	AlertWebhooksTotal             *prometheus.GaugeVec
	AlertWebhookDeliveries         *prometheus.CounterVec
	AlertWebhookLatency            *prometheus.HistogramVec
	EscalationPoliciesTotal        *prometheus.GaugeVec
	EscalationsTriggered           *prometheus.CounterVec
	SnoozeRulesTotal               *prometheus.GaugeVec
	SnoozedAlertsTotal             *prometheus.CounterVec
	QueryBasedAlertsTotal          *prometheus.GaugeVec
	QueryBasedAlertEvaluations     *prometheus.CounterVec
	QueryBasedAlertErrors          *prometheus.CounterVec
	MetricThresholdAlertsTotal     *prometheus.GaugeVec
	LogBasedAlertsTotal            *prometheus.GaugeVec
	LogBasedAlertMatches           *prometheus.CounterVec
	TraceBasedAlertsTotal          *prometheus.GaugeVec
	TraceBasedAlertMatches         *prometheus.CounterVec
	ProfileAlertsTotal             *prometheus.GaugeVec
	HeatmapAlertsTotal             *prometheus.GaugeVec
	AlertAPIRequestsTotal          *prometheus.CounterVec
	AlertAPIRequestErrors          *prometheus.CounterVec
	AlertStateChangesTotal         *prometheus.CounterVec
	PeakAlertRate                  *prometheus.GaugeVec
	AlertFatigueScore              *prometheus.GaugeVec
	AlertsPerOperator              *prometheus.GaugeVec
	MeanTimeToResolutionByRule     *prometheus.HistogramVec
	RecurringAlertsTotal           *prometheus.CounterVec
}{
	AlertRulesTotal:                AlertRulesTotal,
	AlertRuleCreationTotal:         AlertRuleCreationTotal,
	AlertRuleUpdatesTotal:          AlertRuleUpdatesTotal,
	AlertRuleDeletionsTotal:        AlertRuleDeletionsTotal,
	AlertRuleValidationErrors:      AlertRuleValidationErrors,
	AlertTemplatesTotal:            AlertTemplatesTotal,
	AlertTemplatesUsed:             AlertTemplatesUsed,
	TemplateCustomizationRate:      TemplateCustomizationRate,
	AlertRoutingRulesTotal:         AlertRoutingRulesTotal,
	AlertRoutingEvaluationTime:     AlertRoutingEvaluationTime,
	AlertRoutingMatchesTotal:       AlertRoutingMatchesTotal,
	NotificationChannelsConfigured: NotificationChannelsConfigured,
	NotificationChannelHealth:      NotificationChannelHealth,
	DeduplicationRulesTotal:        DeduplicationRulesTotal,
	DuplicateAlertsDetected:        DuplicateAlertsDetected,
	SilenceRulesTotal:              SilenceRulesTotal,
	SilenceMatchesTotal:            SilenceMatchesTotal,
	InhibitionRulesTotal:           InhibitionRulesTotal,
	InhibitedAlertsTotal:           InhibitedAlertsTotal,
	AlertGroupingRulesTotal:        AlertGroupingRulesTotal,
	AlertGroupsCreated:             AlertGroupsCreated,
	AlertWebhooksTotal:             AlertWebhooksTotal,
	AlertWebhookDeliveries:         AlertWebhookDeliveries,
	AlertWebhookLatency:            AlertWebhookLatency,
	EscalationPoliciesTotal:        EscalationPoliciesTotal,
	EscalationsTriggered:           EscalationsTriggered,
	SnoozeRulesTotal:               SnoozeRulesTotal,
	SnoozedAlertsTotal:             SnoozedAlertsTotal,
	QueryBasedAlertsTotal:          QueryBasedAlertsTotal,
	QueryBasedAlertEvaluations:     QueryBasedAlertEvaluations,
	QueryBasedAlertErrors:          QueryBasedAlertErrors,
	MetricThresholdAlertsTotal:     MetricThresholdAlertsTotal,
	LogBasedAlertsTotal:            LogBasedAlertsTotal,
	LogBasedAlertMatches:           LogBasedAlertMatches,
	TraceBasedAlertsTotal:          TraceBasedAlertsTotal,
	TraceBasedAlertMatches:         TraceBasedAlertMatches,
	ProfileAlertsTotal:             ProfileAlertsTotal,
	HeatmapAlertsTotal:             HeatmapAlertsTotal,
	AlertAPIRequestsTotal:          AlertAPIRequestsTotal,
	AlertAPIRequestErrors:          AlertAPIRequestErrors,
	AlertStateChangesTotal:         AlertStateChangesTotal,
	PeakAlertRate:                  PeakAlertRate,
	AlertFatigueScore:              AlertFatigueScore,
	AlertsPerOperator:              AlertsPerOperator,
	MeanTimeToResolutionByRule:     MeanTimeToResolutionByRule,
	RecurringAlertsTotal:           RecurringAlertsTotal,
}
