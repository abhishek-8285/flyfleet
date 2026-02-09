package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

// MonitoringHandlers handles serving mock data for the monitoring dashboard
type MonitoringHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewMonitoringHandlers(cfg *config.Config, logger *zap.Logger) *MonitoringHandlers {
	return &MonitoringHandlers{cfg: cfg, logger: logger}
}

// GetSystemHealth serves mock data for the main dashboard
func (h *MonitoringHandlers) GetSystemHealth(w http.ResponseWriter, r *http.Request) {
	// Simulate data processing
	h.logger.Info("Serving mock system health data")

	data := map[string]interface{}{
		"status": "Operational",
		"uptime_percent": 99.99,
		"avg_latency_ms": 42 + rand.Intn(10), // Randomize slightly
		"error_rate_percent": 0.02,
		"database_status": "Healthy",
		"queue_backlog": 142 - rand.Intn(20),
		"active_integrations": map[string]int{
			"total": 30,
			"active": 28,
		},
		"top_errors": []map[string]interface{}{
			{"service": "Payment Gateway", "code": 500, "type": "TimeoutException", "count": 124},
			{"service": "Public API V1", "code": 429, "type": "RateLimitExceeded", "count": 89},
			{"service": "Reporting Service", "code": 502, "type": "BadGateway", "count": 45},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// GetMetrics serves mock data for the metrics hub
func (h *MonitoringHandlers) GetMetrics(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Serving mock metrics data")
	
	data := map[string]interface{}{
		"throughput": 4892 + rand.Intn(500),
		"db_latency_ms": 142 + rand.Intn(15),
		"cache_hit_ratio": 98.2,
		"cluster_usage": []map[string]interface{}{
			{"service": "API Gateway Core", "cpu": 42, "memory_gb": 8.2, "status": "HEALTHY"},
			{"service": "Worker Node 01", "cpu": 65, "memory_gb": 12.1, "status": "HEALTHY"},
			{"service": "DB Primary", "cpu": 28, "memory_gb": 16.4, "status": "HEALTHY"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// GetLogs serves mock error logs
func (h *MonitoringHandlers) GetLogs(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Serving mock logs")

	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-3 * time.Minute).Format(time.RFC3339),
			"service": "auth-service",
			"severity": "CRITICAL",
			"type": "DatabaseConnectionError",
			"message": "Connection refused to postgres-primary:5432. Retries exhausted.",
			"count": 42,
		},
		{
			"timestamp": time.Now().Add(-4 * time.Minute).Format(time.RFC3339),
			"service": "payment-gateway",
			"severity": "CRITICAL",
			"type": "PaymentDeclinedException",
			"message": "Transaction ID tx_998877 failed validation. Insufficient funds response.",
			"count": 12,
		},
		{
			"timestamp": time.Now().Add(-15 * time.Minute).Format(time.RFC3339),
			"service": "api-gateway",
			"severity": "WARNING",
			"type": "RateLimitExceeded",
			"message": "Client ID client_123 exceeded quota of 1000 req/min.",
			"count": 156,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs": logs,
		"total": 210,
	})
}

// GetQueues serves mock queue status
func (h *MonitoringHandlers) GetQueues(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Serving mock queue data")

	data := map[string]interface{}{
		"running_jobs": 42,
		"pending_queue": 1840,
		"failed_24h": 23,
		"dlq_count": 12,
		"active_jobs": []map[string]interface{}{
			{"id": "JOB-8821", "type": "Data Export", "progress": 45, "duration": "00:02:15"},
			{"id": "JOB-8824", "type": "Email Blast", "progress": 12, "duration": "00:00:45"},
			{"id": "JOB-8819", "type": "Report Gen", "progress": 89, "duration": "00:05:10"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// GetSLA serves mock SLA data
func (h *MonitoringHandlers) GetSLA(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Serving mock SLA data")

	// Generate heatmap data (1 = operational, 2 = warning, 3 = outage)
	heatmap := make([]int, 90)
	for i := 0; i < 90; i++ {
		val := 1
		if rand.Float32() > 0.95 {
			val = 2
		}
		if rand.Float32() > 0.99 {
			val = 3
		}
		heatmap[i] = val
	}

	data := map[string]interface{}{
		"global_uptime_90d": 99.98,
		"mttr_minutes": 14,
		"error_budget_remaining": "2h 15m",
		"active_incidents": 0,
		"heatmap": heatmap,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
