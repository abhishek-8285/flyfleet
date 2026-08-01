package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	EventsIngested = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gonotify",
			Subsystem: "events",
			Name:      "ingested_total",
			Help:      "Total number of events ingested",
		},
		[]string{"tenant_id", "event_type", "status"},
	)

	NotificationsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gonotify",
			Subsystem: "notifications",
			Name:      "processed_total",
			Help:      "Total number of notifications processed",
		},
		[]string{"tenant_id", "channel", "status"},
	)

	DeliveryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "gonotify",
			Subsystem: "delivery",
			Name:      "duration_seconds",
			Help:      "Latency of notification delivery attempts in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"tenant_id", "channel", "status"},
	)
)
