package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/flyfleet/fleet-core/internal/domains/dispatch/engine"
	"github.com/flyfleet/fleet-core/internal/domains/dispatch/service"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DispatchHandler struct {
	dispatchSvc service.DispatchService
	suggestEng  engine.SuggestionEngine
	logger      *zap.Logger
}

func NewDispatchHandler(svc service.DispatchService, eng engine.SuggestionEngine, logger *zap.Logger) *DispatchHandler {
	return &DispatchHandler{
		dispatchSvc: svc,
		suggestEng:  eng,
		logger:      logger,
	}
}

func (h *DispatchHandler) RegisterRoutes(r chi.Router) {
	r.Route("/dispatch", func(r chi.Router) {
		r.Post("/suggest", h.GetSuggestions)
		r.Get("/dashboard", h.GetDashboard)
	})
}

func (h *DispatchHandler) Assign(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := ctxkeys.GetTenantID(r.Context())
	if !ok {
		http.Error(w, `{"error":"missing tenant_id"}`, http.StatusUnauthorized)
		return
	}

	var req service.AssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	result, err := h.dispatchSvc.Assign(r.Context(), tenantID, req)
	if err != nil {
		http.Error(w, `{"error":"failed to assign"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (h *DispatchHandler) GetSuggestions(w http.ResponseWriter, r *http.Request) {
	var req engine.SuggestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	tenantID, ok := ctxkeys.GetTenantID(r.Context())
	if !ok {
		http.Error(w, `{"error":"missing tenant_id"}`, http.StatusUnauthorized)
		return
	}
	req.TenantID = tenantID

	if req.ShipmentID == uuid.Nil {
		http.Error(w, `{"error":"invalid shipment_id"}`, http.StatusBadRequest)
		return
	}

	suggestions, err := h.suggestEng.Suggest(r.Context(), req)
	if err != nil {
		http.Error(w, `{"error":"failed to get suggestions"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(suggestions)
}

func (h *DispatchHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
}

func (h *DispatchHandler) Redispatch(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := ctxkeys.GetTenantID(r.Context())
	if !ok {
		http.Error(w, `{"error":"missing tenant_id"}`, http.StatusUnauthorized)
		return
	}

	var req service.RedispatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	result, err := h.dispatchSvc.Redispatch(r.Context(), tenantID, req)
	if err != nil {
		http.Error(w, `{"error":"failed to redispatch"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
