package emergency

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	IncidentsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flyfleet_emergency_incidents_total",
		Help: "The total number of emergency SOS incidents triggered",
	}, []string{"organization_id", "incident_type", "severity"})

	ActiveIncidents = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flyfleet_emergency_active_incidents",
		Help: "The current number of active (triggered or acknowledged) emergency incidents",
	}, []string{"organization_id"})

	EscalationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flyfleet_emergency_escalations_total",
		Help: "The total number of emergency contact escalations attempted",
	}, []string{"organization_id", "channel"})

	NotificationsSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flyfleet_emergency_notifications_sent",
		Help: "The total number of emergency notifications sent via various channels",
	}, []string{"organization_id", "channel", "status"})

	ResponseTimeSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "flyfleet_emergency_response_time_seconds",
		Help:    "Histogram of emergency response times (trigger to acknowledgment)",
		Buckets: prometheus.ExponentialBuckets(60, 2, 8), // Start at 1 min, double up to 128 mins
	}, []string{"organization_id", "incident_type"})

	SMSSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flyfleet_emergency_sms_sent_total",
		Help: "The total number of emergency SMS alerts sent",
	}, []string{"organization_id"})

	CallAlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flyfleet_emergency_call_alerts_total",
		Help: "The total number of emergency voice call alerts sent",
	}, []string{"organization_id"})
)
