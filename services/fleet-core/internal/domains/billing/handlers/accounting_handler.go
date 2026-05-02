package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/service"
	coremiddleware "github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
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
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		h.errorJSON(w, r, "Unauthorized: tenant context required", http.StatusUnauthorized)
		return
	}
	userID := h.getUserID(r)
	if userID == uuid.Nil {
		h.logger.Warn("Missing user_id in context - request likely bypassed authentication middleware")
		h.errorJSON(w, r, "Unauthorized: user context required", http.StatusUnauthorized)
		return
	}
	system := models.ExternalSystem(r.URL.Query().Get("system"))
	if system == "" {
		system = models.SystemQuickBooks
	}

	export, err := h.accountingServ.ExportLedger(r.Context(), tenantID, system, userID)
	if err != nil {
		// SECURITY FIX: Log detailed error but return generic message to client
		h.logger.Error("ExportLedger failed",
			zap.Error(err),
			zap.String("tenant_id", tenantID.String()),
			zap.String("user_id", userID.String()),
			zap.String("system", string(system)))
		h.errorJSON(w, r, "Failed to export ledger", http.StatusInternalServerError)
		return
	}

	h.json(w, export, http.StatusOK)
}

// ExportInvoices handles POST /api/v1/billing/accounting/export/invoices
func (h *AccountingHandler) ExportInvoices(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		h.errorJSON(w, r, "Unauthorized: tenant context required", http.StatusUnauthorized)
		return
	}
	userID := h.getUserID(r)
	if userID == uuid.Nil {
		h.logger.Warn("Missing user_id in context - request likely bypassed authentication middleware")
		h.errorJSON(w, r, "Unauthorized: user context required", http.StatusUnauthorized)
		return
	}
	system := models.ExternalSystem(r.URL.Query().Get("system"))
	if system == "" {
		system = models.SystemQuickBooks
	}

	export, err := h.accountingServ.ExportInvoices(r.Context(), tenantID, system, userID)
	if err != nil {
		// SECURITY FIX: Log detailed error but return generic message
		h.logger.Error("ExportInvoices failed",
			zap.Error(err),
			zap.String("tenant_id", tenantID.String()),
			zap.String("user_id", userID.String()),
			zap.String("system", string(system)))
		h.errorJSON(w, r, "Failed to export invoices", http.StatusInternalServerError)
		return
	}

	h.json(w, export, http.StatusOK)
}

// GetExportHistory handles GET /api/v1/billing/accounting/history
func (h *AccountingHandler) GetExportHistory(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		h.errorJSON(w, r, "Unauthorized: tenant context required", http.StatusUnauthorized)
		return
	}
	history, err := h.accountingServ.GetExportHistory(r.Context(), tenantID)
	if err != nil {
		// SECURITY FIX: Log detailed error but return generic message
		h.logger.Error("GetExportHistory failed",
			zap.Error(err),
			zap.String("tenant_id", tenantID.String()))
		h.errorJSON(w, r, "Failed to get export history", http.StatusInternalServerError)
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

func (h *AccountingHandler) errorJSON(w http.ResponseWriter, r *http.Request, message string, status int) {
	h.json(w, map[string]string{"error": message, "trace_id": coremiddleware.GetTraceID(r.Context())}, status)
}

func (h *AccountingHandler) getTenantID(r *http.Request) uuid.UUID {
	if id, ok := ctxkeys.GetTenantID(r.Context()); ok {
		return id
	}
	return uuid.Nil
}

func (h *AccountingHandler) getUserID(r *http.Request) uuid.UUID {
	if id, ok := ctxkeys.GetUserID(r.Context()); ok {
		return id
	}
	return uuid.Nil
}
