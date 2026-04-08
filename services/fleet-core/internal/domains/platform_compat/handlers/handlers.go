package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flyfleet/fleet-core/internal/domains/platform_compat/models"
	"github.com/flyfleet/fleet-core/internal/domains/platform_compat/service"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handlers struct {
	service service.PlatformCompatService
	hub     *service.Hub
	logger  *zap.Logger
}

func NewHandlers(service service.PlatformCompatService, hub *service.Hub, logger *zap.Logger) *Handlers {
	return &Handlers{
		service: service,
		hub:     hub,
		logger:  logger,
	}
}

func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Get("/ws", h.WebsocketCompat)
	r.Get("/dashboard/stats", h.DashboardStats)
	r.Get("/admin/dashboard/data", h.AdminDashboardData)

	r.Post("/compliance/hos", h.ComplianceHOSCompat)
	r.Post("/telemetry", h.TelemetryCompat)
	r.Post("/telemetry/dashcam", h.TelemetryDashcamCompat)

	r.Get("/ml/health", h.MLHealth)
	r.Post("/ml/demand/predict", h.PredictDemand)
	r.Post("/ml/pricing/suggest", h.SuggestPricing)
	r.Post("/ml/jobs/train", h.TrainJob)

	r.Get("/routes/{route_id}", h.GetOptimizedRoute)
	r.Post("/optimize", h.OptimizeRoute)

	r.Get("/services/{name}/health", h.ServiceHealth)
}

func (h *Handlers) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

func (h *Handlers) writeErrorResponse(w http.ResponseWriter, status int, err error) {
	resp := models.ErrorResponse{
		Error: err.Error(),
	}

	if platformErr, ok := err.(*service.PlatformError); ok {
		resp.Code = platformErr.Code
		resp.Field = platformErr.Field
	}

	h.writeJSON(w, status, resp)
}

func (h *Handlers) WebsocketCompat(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.Header.Get("X-Org-ID")
	if orgIDStr == "" {
		orgIDStr = r.URL.Query().Get("org_id")
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		if t := r.URL.Query().Get("token"); t != "" {
			authHeader = "Bearer " + t
		}
	}

	if orgIDStr == "" {
		h.logger.Warn("Websocket connect attempt without X-Org-ID")
		http.Error(w, "X-Org-ID header required", http.StatusForbidden)
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.logger.Warn("Websocket connect attempt with invalid OrgID", zap.String("org_id", orgIDStr))
		http.Error(w, "Invalid OrgID", http.StatusForbidden)
		return
	}

	var jwtOrgID string
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
		if err == nil {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if orgClaim, ok := claims["org_id"].(string); ok {
					jwtOrgID = orgClaim
				}
			}
		}
	}

	if jwtOrgID != orgIDStr {
		h.logger.Warn("Tenant Isolation Violation", zap.String("requested_org", orgIDStr), zap.String("jwt_org", jwtOrgID))
		http.Error(w, "Tenant isolation violation", http.StatusForbidden)
		return
	}

	h.hub.HandleWebsocket(w, r, orgID)
}

func (h *Handlers) DashboardStats(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetDashboardStats(r.Context())
	if err != nil {
		h.logger.Error("Failed to get dashboard stats", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) AdminDashboardData(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetAdminDashboardData(r.Context())
	if err != nil {
		h.logger.Error("Failed to get admin dashboard data", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) ComplianceHOSCompat(w http.ResponseWriter, r *http.Request) {
	var req models.ComplianceHOSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	resp, err := h.service.RecordHOS(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to record HOS", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	h.writeJSON(w, http.StatusAccepted, resp)
}

func (h *Handlers) TelemetryCompat(w http.ResponseWriter, r *http.Request) {
	var req models.TelemetryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	resp, err := h.service.IngestTelemetry(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to ingest telemetry", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	h.writeJSON(w, http.StatusAccepted, resp)
}

func (h *Handlers) TelemetryDashcamCompat(w http.ResponseWriter, r *http.Request) {
	var req models.DashcamTelemetryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	resp, err := h.service.IngestDashcamTelemetry(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to ingest dashcam telemetry", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	h.writeJSON(w, http.StatusAccepted, resp)
}

func (h *Handlers) MLHealth(w http.ResponseWriter, r *http.Request) {
	resp := h.service.GetMLHealth(r.Context())
	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) PredictDemand(w http.ResponseWriter, r *http.Request) {
	var req models.MLPredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	resp, err := h.service.PredictDemand(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to predict demand", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) SuggestPricing(w http.ResponseWriter, r *http.Request) {
	var req models.MLPricingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	resp, err := h.service.SuggestPricing(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to suggest pricing", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) TrainJob(w http.ResponseWriter, r *http.Request) {
	var req models.MLTrainingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	resp, err := h.service.TrainModel(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to train model", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	h.writeJSON(w, http.StatusAccepted, resp)
}

func (h *Handlers) GetOptimizedRoute(w http.ResponseWriter, r *http.Request) {
	routeID := chi.URLParam(r, "route_id")

	resp, err := h.service.GetOptimizedRoute(r.Context(), routeID)
	if err != nil {
		h.logger.Error("Failed to get optimized route", zap.Error(err))

		if err == service.ErrRouteNotFound {
			h.writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "route not found"})
			return
		}
		if err == service.ErrRateLimitExceeded {
			w.Header().Set("Retry-After", "60")
			h.writeJSON(w, http.StatusTooManyRequests, models.ErrorResponse{Error: "rate limit exceeded"})
			return
		}

		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) OptimizeRoute(w http.ResponseWriter, r *http.Request) {
	var req models.RouteOptimizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	resp, err := h.service.OptimizeRoute(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to optimize route", zap.Error(err))

		if err == service.ErrInsufficientLocations {
			h.writeJSON(w, http.StatusUnprocessableEntity, models.ErrorResponse{Error: "at least two locations are required"})
			return
		}

		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) ServiceHealth(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(chi.URLParam(r, "name"))
	if name == "" {
		h.writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "service name required"})
		return
	}

	resp := h.service.CheckServiceHealth(r.Context(), name)
	h.writeJSON(w, http.StatusOK, resp)
}
