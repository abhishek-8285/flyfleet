package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AccountingHandler struct {
	accountingServ *service.AccountingService
	logger         *zap.Logger
}

func NewAccountingHandler(accServ *service.AccountingService, logger *zap.Logger) *AccountingHandler {
	return &AccountingHandler{
		accountingServ: accServ,
		logger:         logger,
	}
}

// ExportLedger handles POST /api/v1/billing/accounting/export/ledger
func (h *AccountingHandler) ExportLedger(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	system := models.ExternalSystem(r.URL.Query().Get("system"))
	if system == "" {
		system = models.SystemQuickBooks
	}

	export, err := h.accountingServ.ExportLedger(r.Context(), orgID, system, userID)
	if err != nil {
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, export, http.StatusOK)
}

// ExportInvoices handles POST /api/v1/billing/accounting/export/invoices
func (h *AccountingHandler) ExportInvoices(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	system := models.ExternalSystem(r.URL.Query().Get("system"))
	if system == "" {
		system = models.SystemQuickBooks
	}

	export, err := h.accountingServ.ExportInvoices(r.Context(), orgID, system, userID)
	if err != nil {
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, export, http.StatusOK)
}

// GetExportHistory handles GET /api/v1/billing/accounting/history
func (h *AccountingHandler) GetExportHistory(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	history, err := h.accountingServ.GetExportHistory(r.Context(), orgID)
	if err != nil {
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, history, http.StatusOK)
}

// Helper methods for JSON responses
func (h *AccountingHandler) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode json response", zap.Error(err))
	}
}

func (h *AccountingHandler) errorJSON(w http.ResponseWriter, message string, status int) {
	h.json(w, map[string]string{"error": message}, status)
}

func (h *AccountingHandler) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

func (h *AccountingHandler) getUserID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.UserIDKey); val != nil {
		if userIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(userIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}
