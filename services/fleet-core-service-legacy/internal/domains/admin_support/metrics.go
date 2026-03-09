package admin_support

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds the prometheus metrics for the Admin Support Module.
type Metrics struct {
	TicketsOpenTotal       prometheus.Gauge
	TicketsResolvedTotal   prometheus.Counter
	IncidentsTotal         *prometheus.CounterVec
	AdminActionsTotal      *prometheus.CounterVec
	ImpersonationActive    prometheus.Gauge
	SLAExceededTotal       prometheus.Counter
}

// NewMetrics initializes and registers the metrics for the Admin Support Module.
func NewMetrics() *Metrics {
	return &Metrics{
		TicketsOpenTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "admin_support_tickets_open_total",
			Help: "Total number of open support tickets",
		}),
		TicketsResolvedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "admin_support_tickets_resolved_total",
			Help: "Total number of resolved support tickets",
		}),
		IncidentsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "admin_support_incidents_total",
			Help: "Total number of declared incidents",
		}, []string{"severity"}),
		AdminActionsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "admin_support_actions_total",
			Help: "Total number of administrative actions performed",
		}, []string{"action", "resource_type"}),
		ImpersonationActive: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "admin_support_impersonation_sessions_active",
			Help: "Number of active impersonation sessions",
		}),
		SLAExceededTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "admin_support_sla_exceeded_total",
			Help: "Total number of tickets that exceeded their SLA deadline",
		}),
	}
}
