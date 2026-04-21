package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core/internal/domains/tracking/analytics"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AnalyticsHandler struct {
	analytics *analytics.TelemetryAnalytics
	logger    *zap.Logger
}

func NewAnalyticsHandler(analytics *analytics.TelemetryAnalytics, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analytics: analytics,
		logger:    logger,
	}
}

// RegisterRoutes registers analytics-related routes
func (ah *AnalyticsHandler) RegisterRoutes(r chi.Router) {
	r.Route("/analytics", func(r chi.Router) {
		r.Get("/vehicle/{id}", ah.GetVehicleAnalytics)
		r.Get("/anomalies/vehicle/{id}", ah.GetVehicleAnomalies)
	})
}

// Helper methods for JSON responses
func (ah *AnalyticsHandler) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (ah *AnalyticsHandler) errorJSON(w http.ResponseWriter, message string, status int) {
	ah.json(w, map[string]string{"error": message}, status)
}

// GetVehicleAnalytics handles GET /analytics/vehicle/:id
func (ah *AnalyticsHandler) GetVehicleAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == "" {
		ah.errorJSON(w, "missing organization context", http.StatusForbidden)
		return
	}

	vehicleIDStr := chi.URLParam(r, "id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		ah.errorJSON(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	startTime := time.Now().AddDate(0, 0, -7)
	endTime := time.Now()

	// Process analytics (RLS-enforced via TenantID and Transaction)
	err = ah.analytics.ProcessAdvancedAnalytics(r.Context(), tenantID, vehicleID, startTime, endTime)
	if err != nil {
		ah.logger.Error("Failed to process analytics", zap.Error(err))
		ah.errorJSON(w, "Failed to process analytics", http.StatusInternalServerError)
		return
	}

	ah.json(w, map[string]string{"status": "success", "message": "Analytics processed"}, http.StatusOK)
}

// GetVehicleAnomalies handles GET /analytics/anomalies/vehicle/:id
func (ah *AnalyticsHandler) GetVehicleAnomalies(w http.ResponseWriter, r *http.Request) {
	// ... implementation using tx.Query to fetch anomalies for the tenant ...
	ah.json(w, []interface{}{}, http.StatusOK)
}
