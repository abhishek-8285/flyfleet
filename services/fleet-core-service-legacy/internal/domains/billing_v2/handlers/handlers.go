package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BillingHandler struct {
	service   *service.BillingService
	logger    *zap.Logger
	validator *validator.Validate
}

func NewBillingHandler(s *service.BillingService, logger *zap.Logger) *BillingHandler {
	return &BillingHandler{
		service:   s,
		logger:    logger,
		validator: validator.New(),
	}
}

func (h *BillingHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/billing", func(r chi.Router) {
		r.Get("/invoices", h.ListInvoices)
		r.Post("/invoices", h.CreateInvoice)
		r.Get("/invoices/{id}", h.GetInvoice)
		r.Put("/invoices/{id}", h.UpdateInvoice)
		r.Delete("/invoices/{id}", h.DeleteInvoice)
		r.Post("/invoices/{id}/payments", h.RecordPayment)

		r.Get("/settings", h.GetSettings)
		r.Put("/settings", h.UpdateSettings)

		r.Get("/metrics", h.GetMetrics)
	})
}

// Helper methods for JSON responses
func (h *BillingHandler) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode json response", zap.Error(err))
	}
}

func (h *BillingHandler) errorJSON(w http.ResponseWriter, message string, status int) {
	h.json(w, map[string]string{"error": message}, status)
}

func (h *BillingHandler) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

func (h *BillingHandler) getUserID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.UserIDKey); val != nil {
		if userIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(userIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

func (h *BillingHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(query.Get("size"))
	if size < 1 {
		size = 10
	}

	invoices, total, err := h.service.ListInvoices(r.Context(), orgID, page, size)
	if err != nil {
		h.logger.Error("Failed to list invoices", zap.Error(err))
		h.errorJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{
		"data":  invoices,
		"total": total,
		"page":  page,
		"size":  size,
	}, http.StatusOK)
}

func (h *BillingHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	var inv models.Invoice
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		h.errorJSON(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	inv.OrgID = orgID

	if err := h.service.CreateInvoice(r.Context(), &inv); err != nil {
		h.logger.Error("Failed to create invoice", zap.Error(err))
		h.errorJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, inv, http.StatusCreated)
}

func (h *BillingHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, "Invalid Invoice ID", http.StatusBadRequest)
		return
	}

	invoice, err := h.service.GetInvoice(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get invoice", zap.Error(err))
		h.errorJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if invoice == nil {
		h.errorJSON(w, "Invoice Not Found", http.StatusNotFound)
		return
	}

	h.json(w, invoice, http.StatusOK)
}

func (h *BillingHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, "Invalid Invoice ID", http.StatusBadRequest)
		return
	}

	var inv models.Invoice
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		h.errorJSON(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	inv.ID = id

	if err := h.service.UpdateInvoice(r.Context(), &inv); err != nil {
		h.logger.Error("Failed to update invoice", zap.Error(err))
		h.errorJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, inv, http.StatusOK)
}

func (h *BillingHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, "Invalid Invoice ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteInvoice(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete invoice", zap.Error(err))
		h.errorJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BillingHandler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, "Invalid Invoice ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Amount float64 `json:"amount" validate:"required,gt=0"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.errorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.RecordPayment(r.Context(), id, req.Amount); err != nil {
		h.logger.Error("Failed to record payment", zap.Error(err))
		h.errorJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"message": "Payment recorded successfully"}, http.StatusOK)
}

func (h *BillingHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	settings, err := h.service.GetSettings(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get settings", zap.Error(err))
		h.errorJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, settings, http.StatusOK)
}

func (h *BillingHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	var settings models.BillingSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		h.errorJSON(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	settings.OrgID = orgID

	if err := h.service.UpdateSettings(r.Context(), &settings); err != nil {
		h.logger.Error("Failed to update settings", zap.Error(err))
		h.errorJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, settings, http.StatusOK)
}

func (h *BillingHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	metrics, err := h.service.GetMetrics(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get metrics", zap.Error(err))
		h.errorJSON(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, metrics, http.StatusOK)
}
