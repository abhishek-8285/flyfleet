package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	GovernanceRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "governance_requests_total",
		Help: "Total number of governance requests by type and status",
	}, []string{"type", "status"})

	GovernanceExportsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "governance_exports_total",
		Help: "Total number of data exports generated",
	})

	GovernanceDeletionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "governance_deletions_total",
		Help: "Total number of data deletion requests processed",
	})

	GovernanceRequestProcessingSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "governance_request_processing_seconds",
		Help:    "Time taken to process governance requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"type"})

	GovernanceExportSizeBytes = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "governance_export_size_bytes",
		Help:    "Size of generated export files in bytes",
		Buckets: prometheus.ExponentialBuckets(1024, 2, 10), // 1KB to 1MB+
	})

	GovernancePendingRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "governance_pending_requests",
		Help: "Current number of pending governance requests",
	})
)
