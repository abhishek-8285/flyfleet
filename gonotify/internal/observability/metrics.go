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

	RetriesScheduled = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gonotify",
			Subsystem: "notifications",
			Name:      "retry_scheduled_total",
			Help:      "Total number of notifications scheduled for retry",
		},
		[]string{"tenant_id", "channel"},
	)

	DeadLetterMoved = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gonotify",
			Subsystem: "notifications",
			Name:      "dead_letter_total",
			Help:      "Total number of notifications moved to dead-letter",
		},
		[]string{"tenant_id", "channel"},
	)

	DeadLetterPending = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gonotify",
			Subsystem: "notifications",
			Name:      "dead_letter_pending",
			Help:      "Current number of dead-letter notifications",
		},
		[]string{"tenant_id"},
	)

	RetryProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gonotify",
			Subsystem: "notifications",
			Name:      "retry_processed_total",
			Help:      "Total number of retry attempts processed by the retry worker",
		},
		[]string{"tenant_id", "channel"},
	)

	RetryFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gonotify",
			Subsystem: "notifications",
			Name:      "retry_failed_total",
			Help:      "Total number of retry attempts that failed during processing",
		},
		[]string{"tenant_id", "channel"},
	)
)
