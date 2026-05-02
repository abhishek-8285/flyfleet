package emergency

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	IncidentsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_incidents_total",
		Help: "The total number of emergency SOS incidents triggered",
	}, []string{"tenant_id", "incident_type", "severity"})

	ActiveIncidents = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "emergency_active_incidents",
		Help: "The current number of active (triggered or acknowledged) emergency incidents",
	}, []string{"tenant_id"})

	EscalationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_escalations_total",
		Help: "The total number of emergency contact escalations attempted",
	}, []string{"tenant_id", "channel"})

	NotificationsSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_notifications_sent",
		Help: "The total number of emergency notifications sent via various channels",
	}, []string{"tenant_id", "channel", "status"})

	ResponseTimeSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "emergency_response_time_seconds",
		Help:    "Histogram of emergency response times (trigger to acknowledgment)",
		Buckets: prometheus.ExponentialBuckets(60, 2, 8), // Start at 1 min, double up to 128 mins
	}, []string{"tenant_id", "incident_type"})

	SMSSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_sms_sent_total",
		Help: "The total number of emergency SMS alerts sent",
	}, []string{"tenant_id"})

	CallAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_call_alerts_total",
		Help: "The total number of emergency voice call alerts sent",
	}, []string{"tenant_id"})

	SLABreachesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_sla_breaches_total",
		Help: "The total number of emergency SLA breaches detected",
	}, []string{"tenant_id", "severity"})

	NotificationFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_notification_failures_total",
		Help: "The total number of emergency notifications that failed after all retries",
	}, []string{"tenant_id", "channel"})

	// Phase 8: Additional Prometheus Metrics
	AuthorityNotificationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_authority_notifications_total",
		Help: "Total number of AIS-140 authority notifications dispatched",
	}, []string{"authority_type", "severity"})

	NotificationRetriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_notification_retries_total",
		Help: "Total number of notification provider retries (per provider, per channel)",
	}, []string{"provider", "channel"})

	IncidentRecoveryTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_incident_recovery_total",
		Help: "Total number of stalled incidents resumed by the recovery worker",
	}, []string{"status"})

	DrillEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_drill_events_total",
		Help: "Total number of emergency drill simulations triggered",
	}, []string{"tenant_id"})

	// Phase 3 Metrics
	ProviderFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_provider_failures_total",
		Help: "Total number of notification provider failures tracked by health service",
	}, []string{"provider"})

	AuthorityDeliverySuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_authority_delivery_success_total",
		Help: "Total number of successful emergency authority deliveries",
	}, []string{"authority_type", "status"})

	IncidentSeverityTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_incident_severity_total",
		Help: "Total number of incidents classified by severity level",
	}, []string{"level"})

	TimelineEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "emergency_timeline_events_total",
		Help: "Total number of fine-grained timeline events recorded",
	}, []string{"event_type"})

	BackupSnapshotsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "emergency_backup_snapshots_total",
		Help: "Total number of critical incident backup snapshots created",
	})

	ArchivedIncidentsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "emergency_archived_incidents_total",
		Help: "Total number of incidents moved to archival storage",
	})
)
