package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/monitoring/service"
	"go.uber.org/zap"
)

type MonitoringHandlers struct {
	logger *zap.Logger
	pulse  service.PulseService
}

func NewMonitoringHandlers(logger *zap.Logger, pulse service.PulseService) *MonitoringHandlers {
	return &MonitoringHandlers{
		logger: logger,
		pulse:  pulse,
	}
}

func (h *MonitoringHandlers) GetSLAMetrics(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"uptime":            h.pulse.GetUptime(),
		"avg_response_ms":   185.0, // This could be dynamic later
		"alert_sla_percent": 96.5,
		"period":            "real_time",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *MonitoringHandlers) GetSystemHealth(w http.ResponseWriter, r *http.Request) {
	health := h.pulse.CheckHealth(r.Context())

	resp := map[string]interface{}{
		"status":    "healthy",
		"services":  health,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// If any critical service is down, overall status is degraded
	for _, status := range health {
		if status == "down" {
			resp["status"] = "degraded"
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
