package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fleet_requests_total",
		Help: "Total number of platform requests",
	}, []string{"method", "path", "status"})

	RequestLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "fleet_request_latency_seconds",
		Help:    "Request latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	ActiveOrganizations = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "fleet_active_organizations",
		Help: "Number of active organizations",
	})

	SubscriptionUsage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fleet_subscription_usage",
		Help: "Usage of subscription quotas",
	}, []string{"organization_id", "metric"})
)
