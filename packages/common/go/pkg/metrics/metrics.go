package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTPRequestDuration tracks duration of HTTP requests
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	// GRPCRequestDuration tracks duration of gRPC requests
	GRPCRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "grpc_request_duration_seconds",
		Help:    "Duration of gRPC requests in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method", "status"})

	// DBQueryDuration tracks duration of database queries
	DBQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_query_latency_seconds",
		Help:    "Duration of database queries in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"query", "table"})

	// NATSMessageProcessed tracks NATS message counts
	NATSMessageProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nats_messages_processed_total",
		Help: "Total number of NATS messages processed",
	}, []string{"subject", "status"})
)

// Handler returns the prometheus HTTP handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// RecordHTTP records an HTTP request duration
func RecordHTTP(method, path, status string, duration float64) {
	HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration)
}

// RecordGRPC records a gRPC request duration
func RecordGRPC(service, method, status string, duration float64) {
	GRPCRequestDuration.WithLabelValues(service, method, status).Observe(duration)
}
