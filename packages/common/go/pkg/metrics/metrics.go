package metrics

import (
	"net/http"
	"strings"
	"time"

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

// RecordDBQuery records database query latency with query and table labels.
// It extracts a simple query type (SELECT, INSERT, etc) and table name from the SQL.
// This is a best-effort parse; if parsing fails, it uses "unknown".
func RecordDBQuery(sql string, duration time.Duration) {
	queryType, table := extractQueryInfo(sql)
	DBQueryDuration.WithLabelValues(queryType, table).Observe(duration.Seconds())
}

// extractQueryInfo parses SQL to get operation type and first table name.
// Very simplified parser - handles basic SELECT, INSERT, UPDATE, DELETE.
func extractQueryInfo(sql string) (queryType, table string) {
	// Clean up whitespace and convert to uppercase for parsing
	cleaned := strings.TrimSpace(sql)
	if cleaned == "" {
		return "unknown", "unknown"
	}

	// Split to get first word as query type
	parts := strings.Fields(cleaned)
	if len(parts) == 0 {
		return "unknown", "unknown"
	}
	queryType = strings.ToUpper(parts[0])

	// Try to extract table name based on query type
	table = "unknown"
	switch queryType {
	case "SELECT", "DELETE":
		// Find FROM keyword
		for i, part := range parts {
			if strings.ToUpper(part) == "FROM" && i+1 < len(parts) {
				table = parts[i+1]
				// Remove any trailing commas or semicolons
				table = strings.TrimRight(table, ",;")
				break
			}
		}
	case "UPDATE":
		// UPDATE table SET ...
		if len(parts) >= 2 {
			table = parts[1]
			table = strings.TrimRight(table, " ;")
		}
	case "INSERT":
		// INSERT INTO table ...
		if len(parts) >= 3 && strings.ToUpper(parts[1]) == "INTO" {
			table = parts[2]
			table = strings.TrimRight(table, " ;(,")
		}
	}
	// Ensure table is lowercase for cardinality
	return queryType, strings.ToLower(table)
}
