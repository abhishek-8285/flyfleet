package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/domains/billing/service"
	"go.uber.org/zap"
)

type AnalyticsHandler struct {
	analyticsServ *service.AnalyticsService
	logger        *zap.Logger
}

func NewAnalyticsHandler(analyticsServ *service.AnalyticsService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsServ: analyticsServ,
		logger:        logger,
	}
}

// GetSystemMetrics handles GET /api/v1/billing/analytics/system
func (h *AnalyticsHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.analyticsServ.GetSystemMetrics(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch metrics", http.StatusInternalServerError)
		return
	}

	h.renderJSON(w, metrics)
}

func (h *AnalyticsHandler) renderJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON", zap.Error(err))
	}
}
