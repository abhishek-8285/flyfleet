package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"


	"github.com/fleet-platform/shared/pkg/middleware"
	coremiddleware "github.com/flyfleet/fleet-core-service/internal/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/cache"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/service"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BillingHandler struct {
	service           *service.BillingService
	invoiceServ       *service.InvoiceService
	subService        *service.SubscriptionService
	meterService      *service.MeteringService
	creditNoteService *service.CreditNoteService
	hwService         *service.HardwareService
	ledgerService     *service.LedgerService
	advSubService     *service.AdvancedSubService
	analyticsHandler  *AnalyticsHandler
	accountingHandler *AccountingHandler
	logger            *zap.Logger
	validator         *validator.Validate
}

func NewBillingHandler(s *service.BillingService, invServ *service.InvoiceService, subServ *service.SubscriptionService, meterServ *service.MeteringService, cnServ *service.CreditNoteService, hwServ *service.HardwareService, ledgerServ *service.LedgerService, advSubServ *service.AdvancedSubService, analyticsH *AnalyticsHandler, accountingH *AccountingHandler, logger *zap.Logger) *BillingHandler {
	return &BillingHandler{
		service:           s,
		invoiceServ:       invServ,
		subService:        subServ,
		meterService:      meterServ,
		creditNoteService: cnServ,
		hwService:         hwServ,
		ledgerService:     ledgerServ,
		advSubService:     advSubServ,
		analyticsHandler:  analyticsH,
		accountingHandler: accountingH,
		logger:            logger,
		validator:         validator.New(),
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
		r.Get("/invoices/{id}/pdf", h.DownloadInvoicePDF)
		r.Post("/invoices/{id}/email", h.EmailInvoice)

		r.Get("/settings", h.GetSettings)
		r.Put("/settings", h.UpdateSettings)
		r.Get("/metrics", h.GetMetrics)
		r.Get("/health", h.HealthCheck)
		r.Get("/stats", h.GetStats)

		// Subscription Routes
		r.Get("/plans", h.ListPlans)
		r.Get("/subscription", h.GetSubscription)
		r.Post("/subscription/subscribe", h.Subscribe)
		r.Post("/subscription/change-plan", h.ChangePlan)
		r.Post("/subscription/prorate-upgrade", h.ProrateUpgrade)
		r.Post("/subscription/redeem-coupon", h.RedeemCoupon)
		r.Post("/subscription/cancel", h.CancelSubscription)

		// Metering & Draft Invoicing
		r.Post("/usage/record", h.RecordUsage)
		r.Post("/invoices/generate-draft", h.GenerateDraftInvoice)

		// External Payment Webhooks
		r.Post("/webhooks/stripe", h.StripeWebhook)
		r.Post("/webhooks/razorpay", h.RazorpayWebhook)

		// Credit Notes
		r.Post("/credit-notes", h.IssueCreditNote)
		r.Get("/credit-notes", h.ListCreditNotes)

		// Hardware Billing
		r.Get("/hardware/products", h.ListHardwareProducts)
		r.Post("/hardware/assign", h.AssignHardware)

		// Ledger & Wallet
		r.Get("/wallet/balance", h.GetBalance)
		r.Post("/wallet/credit", h.ApplyCredit)

		// Financial Analytics
		r.Route("/analytics", func(r chi.Router) {
			r.Get("/system", h.analyticsHandler.GetSystemMetrics)
		})

		// Accounting Integration
		r.Route("/accounting", func(r chi.Router) {
			r.Post("/export/ledger", h.accountingHandler.ExportLedger)
			r.Post("/export/invoices", h.accountingHandler.ExportInvoices)
			r.Get("/history", h.accountingHandler.GetExportHistory)
		})
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

func (h *BillingHandler) errorJSON(w http.ResponseWriter, r *http.Request, message string, status int) {
	traceID := coremiddleware.GetTraceID(r.Context())
	h.json(w, map[string]string{
		"error":    message,
		"trace_id": traceID,
	}, status)
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

func (h *BillingHandler) readInt(r *http.Request, name string, defaultValue int) int {
	val := r.URL.Query().Get(name)
	if val == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return i
}

func (h *BillingHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
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
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
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
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	var inv models.Invoice
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		h.errorJSON(w, r, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// Validate invoice
	if err := h.validator.Struct(&inv); err != nil {
		h.errorJSON(w, r, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	inv.OrgID = orgID

	if err := h.service.CreateInvoice(r.Context(), &inv); err != nil {
		h.logger.Error("Failed to create invoice", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, inv, http.StatusCreated)
}

func (h *BillingHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, r, "Invalid Invoice ID", http.StatusBadRequest)
		return
	}

	invoice, err := h.service.GetInvoice(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get invoice", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if invoice == nil {
		h.errorJSON(w, r, "Invoice Not Found", http.StatusNotFound)
		return
	}

	h.json(w, invoice, http.StatusOK)
}

func (h *BillingHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, r, "Invalid Invoice ID", http.StatusBadRequest)
		return
	}

	var inv models.Invoice
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		h.errorJSON(w, r, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// Validate invoice
	if err := h.validator.Struct(&inv); err != nil {
		h.errorJSON(w, r, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	inv.ID = id

	if err := h.service.UpdateInvoice(r.Context(), &inv); err != nil {
		h.logger.Error("Failed to update invoice", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, inv, http.StatusOK)
}

func (h *BillingHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, r, "Invalid Invoice ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteInvoice(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete invoice", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BillingHandler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, r, "Invalid Invoice ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Amount float64 `json:"amount" validate:"required,gt=0"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, r, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.errorJSON(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.RecordPayment(r.Context(), id, req.Amount); err != nil {
		h.logger.Error("Failed to record payment", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"message": "Payment recorded successfully"}, http.StatusOK)
}

func (h *BillingHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	settings, err := h.service.GetSettings(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get settings", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, settings, http.StatusOK)
}

func (h *BillingHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	var settings models.BillingSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		h.errorJSON(w, r, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	settings.OrgID = orgID

	if err := h.service.UpdateSettings(r.Context(), &settings); err != nil {
		h.logger.Error("Failed to update settings", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, settings, http.StatusOK)
}

func (h *BillingHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.errorJSON(w, r, "Invalid Organization ID", http.StatusBadRequest)
		return
	}

	metrics, err := h.service.GetMetrics(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get metrics", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.json(w, metrics, http.StatusOK)
}

// DownloadInvoicePDF generates and downloads invoice PDF
func (h *BillingHandler) DownloadInvoicePDF(w http.ResponseWriter, r *http.Request) {
	if !h.invoiceServ.Enabled {
		h.errorJSON(w, r, "Invoice generation is currently disabled: "+h.invoiceServ.Error, http.StatusServiceUnavailable)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, r, "Invalid Invoice ID", http.StatusBadRequest)
		return
	}

	// Get invoice with items
	invoice, err := h.service.GetInvoice(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get invoice for PDF", zap.Error(err))
		h.errorJSON(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if invoice == nil {
		h.errorJSON(w, r, "Invoice Not Found", http.StatusNotFound)
		return
	}

	pdfData, filename, err := h.invoiceServ.DownloadPDF(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to generate PDF", zap.Error(err))
		h.errorJSON(w, r, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write(pdfData)
}

// EmailInvoice sends invoice via email
func (h *BillingHandler) EmailInvoice(w http.ResponseWriter, r *http.Request) {
	if !h.invoiceServ.Enabled {
		h.errorJSON(w, r, "Email service is currently disabled", http.StatusServiceUnavailable)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, r, "Invalid Invoice ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Email        string `json:"email" validate:"required,email"`
		CustomerName string `json:"customer_name" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, r, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.errorJSON(w, r, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.invoiceServ.SendInvoiceEmail(r.Context(), id, req.Email, req.CustomerName); err != nil {
		h.logger.Error("Failed to send invoice email", zap.Error(err))
		h.errorJSON(w, r, "Failed to send email", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"status": "sent"}, http.StatusOK)
}

// HealthCheck returns health status of billing components
func (h *BillingHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	type HealthStatus struct {
		Status     string            `json:"status"`
		Enabled    bool              `json:"enabled"`
		Components map[string]string `json:"components"`
	}

	status := HealthStatus{
		Status:     "healthy",
		Enabled:    h.invoiceServ.Enabled,
		Components: make(map[string]string),
	}

	if !h.invoiceServ.Enabled {
		status.Status = "degraded"
		status.Components["invoice_service"] = "disabled: " + h.invoiceServ.Error
	}

	// Check storage
	if h.invoiceServ.Enabled {
		if stats, err := h.invoiceServ.GetStorageStats(r.Context()); err != nil {
			status.Components["storage"] = "unhealthy"
			status.Status = "unhealthy"
		} else {
			status.Components["storage"] = fmt.Sprintf("healthy (%.2f MB, %d files)", stats.TotalSizeMB, stats.FileCount)
		}
	}

	// Check cache
	if h.invoiceServ.Enabled {
		cacheStats := h.invoiceServ.GetCacheStats()
		status.Components["cache"] = fmt.Sprintf("healthy (%d entries, %.2f MB)", cacheStats.EntryCount, cacheStats.TotalSizeMB)
	}

	h.json(w, status, http.StatusOK)
}

// GetStats returns billing statistics
func (h *BillingHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if !h.invoiceServ.Enabled {
		h.errorJSON(w, r, "Service disabled", http.StatusServiceUnavailable)
		return
	}

	type Stats struct {
		Storage *storage.StorageStats `json:"storage"`
		Cache   *cache.CacheStats     `json:"cache"`
	}

	storageStats, _ := h.invoiceServ.GetStorageStats(r.Context())
	cacheStats := h.invoiceServ.GetCacheStats()

	h.json(w, Stats{
		Storage: storageStats,
		Cache:   cacheStats,
	}, http.StatusOK)
}
