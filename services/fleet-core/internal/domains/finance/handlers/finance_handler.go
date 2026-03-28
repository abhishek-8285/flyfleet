package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/finance/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FinanceHandlers struct {
	service *service.FinanceService
	logger  *zap.Logger
}

func NewFinanceHandlers(service *service.FinanceService, logger *zap.Logger) *FinanceHandlers {
	return &FinanceHandlers{service: service, logger: logger}
}

func (h *FinanceHandlers) RegisterRoutes(r chi.Router) {
	r.Get("/expenses/reports", h.ListReports)
	r.Post("/fuel/ingest", h.IngestFuel)
	r.Post("/expenses/reports", h.SubmitReport)
	r.Post("/expenses/reports/{id}/approve", h.ApproveReport)
	r.Post("/expenses/reports/{id}/reject", h.RejectReport)
	r.Get("/expenses/policies", h.ListPolicies)
	r.Put("/expenses/policies", h.UpsertPolicy)
	r.Get("/revenue/forecast", h.GetRevenueForecast)
	r.Get("/revenue/experiments", h.ListRevenueExperiments)
}

func (h *FinanceHandlers) IngestFuel(w http.ResponseWriter, r *http.Request) {
	var tx models.FuelTransaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.IngestFuelData(r.Context(), &tx); err != nil {
		h.logger.Error("Failed to ingest fuel data", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": tx,
		"meta": map[string]string{"message": "Fuel transaction ingested"},
	})
}

func (h *FinanceHandlers) SubmitReport(w http.ResponseWriter, r *http.Request) {
	var report models.ExpenseReport
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.SubmitExpenseReport(r.Context(), &report); err != nil {
		h.logger.Error("Failed to submit expense report", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": report,
		"meta": map[string]string{"message": "Expense report submitted"},
	})
}

func (h *FinanceHandlers) ApproveReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// In a real app, approverID would come from Auth Context
	var body struct {
		ApproverID string `json:"approver_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.ApproveExpenseReport(r.Context(), id, body.ApproverID); err != nil {
		http.Error(w, "Failed to approve report", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Report approved"})
}

func (h *FinanceHandlers) RejectReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		ApproverID string `json:"approver_id"`
		Reason     string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.RejectExpenseReport(r.Context(), id, body.ApproverID, body.Reason); err != nil {
		http.Error(w, "Failed to reject report", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Report rejected"})
}

func (h *FinanceHandlers) ListReports(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	if orgID == "" {
		http.Error(w, "missing organization id", http.StatusBadRequest)
		return
	}
	limit := parseIntDefault(r.URL.Query().Get("limit"), 50)
	offset := parseIntDefault(r.URL.Query().Get("offset"), 0)
	status := r.URL.Query().Get("status")

	payload, err := h.service.ListExpenseReports(r.Context(), orgID, status, limit, offset)
	if err != nil {
		h.logger.Error("failed to list expense reports", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *FinanceHandlers) ListPolicies(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	if orgID == "" {
		http.Error(w, "missing organization id", http.StatusBadRequest)
		return
	}
	policies, err := h.service.ListExpensePolicies(r.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to list expense policies", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"policies": policies})
}

func (h *FinanceHandlers) UpsertPolicy(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	if orgID == "" {
		http.Error(w, "missing organization id", http.StatusBadRequest)
		return
	}
	var policy models.ExpensePolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	policy.OrgID = orgID
	if err := h.service.UpsertExpensePolicy(r.Context(), &policy); err != nil {
		h.logger.Error("failed to upsert expense policy", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": policy, "meta": map[string]string{"message": "Policy saved"}})
}

func (h *FinanceHandlers) GetRevenueForecast(w http.ResponseWriter, r *http.Request) {
	routeID := r.URL.Query().Get("route_id")
	if strings.TrimSpace(routeID) == "" {
		http.Error(w, "route_id is required", http.StatusBadRequest)
		return
	}
	forecast, err := h.service.GetRevenueForecast(r.Context(), routeID)
	if err != nil {
		h.logger.Error("failed to get revenue forecast", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"forecast": forecast})
}

func (h *FinanceHandlers) ListRevenueExperiments(w http.ResponseWriter, r *http.Request) {
	limit := parseIntDefault(r.URL.Query().Get("limit"), 20)
	experiments, err := h.service.ListABExperiments(r.Context(), limit)
	if err != nil {
		h.logger.Error("failed to list revenue experiments", zap.Error(err))
		http.Error(w, "failed to list experiments", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"experiments": experiments})
}

func getOrgID(r *http.Request) string {
	// SECURITY FIX: Get org ID from context only, not from client-supplied headers
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		switch v := val.(type) {
		case string:
			return strings.TrimSpace(v)
		case uuid.UUID:
			return v.String()
		}
	}
	return ""
}

func parseIntDefault(raw string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return v
}
