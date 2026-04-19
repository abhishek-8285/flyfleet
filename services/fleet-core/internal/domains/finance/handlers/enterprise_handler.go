package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core/internal/domains/finance/service"
	"github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EnterpriseFinanceHandlers groups all new enterprise-grade finance API handlers.
type EnterpriseFinanceHandlers struct {
	eInvoice   *service.EInvoiceService
	tds        *service.TDSTaxService
	vendorPay  *service.VendorPaymentService
	tripProfit *service.TripProfitService
	fraudSvc   *service.FraudDetectionService
	reconSvc   *service.FuelCardReconciliationService
	logger     *zap.Logger
}

func NewEnterpriseFinanceHandlers(
	eInvoice *service.EInvoiceService,
	tds *service.TDSTaxService,
	vendorPay *service.VendorPaymentService,
	tripProfit *service.TripProfitService,
	fraudSvc *service.FraudDetectionService,
	reconSvc *service.FuelCardReconciliationService,
	logger *zap.Logger,
) *EnterpriseFinanceHandlers {
	return &EnterpriseFinanceHandlers{
		eInvoice:   eInvoice,
		tds:        tds,
		vendorPay:  vendorPay,
		tripProfit: tripProfit,
		fraudSvc:   fraudSvc,
		reconSvc:   reconSvc,
		logger:     logger,
	}
}

// RegisterRoutes mounts all enterprise finance routes onto the router.
func (h *EnterpriseFinanceHandlers) RegisterRoutes(r chi.Router) {
	// GST E-Invoice
	r.Route("/gst", func(r chi.Router) {
		r.Post("/einvoice", h.GenerateEInvoice)
		r.Get("/einvoice/{id}", h.GetEInvoice)
	})

	// TDS
	r.Route("/tds", func(r chi.Router) {
		r.Post("/deduct", h.DeductTDS)
		r.Get("/report", h.GetTDSReport)
	})

	// Vendor Payments Ledger
	r.Route("/vendors/payments", func(r chi.Router) {
		r.Post("/", h.CreateVendorPayment)
		r.Get("/", h.ListVendorPayments)
	})

	// Trip Profitability
	r.Get("/trips/{id}/profit", h.GetTripProfit)
	r.Get("/finance/dashboard", h.GetFinanceDashboard)

	// Reconciliation
	r.Route("/reconciliation", func(r chi.Router) {
		r.Post("/run", h.RunReconciliation)
		r.Post("/{id}/match", h.ManualMatch)
	})
}

// GenerateEInvoice handles POST /gst/einvoice
func (h *EnterpriseFinanceHandlers) GenerateEInvoice(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, `{"error":"unauthorized: organization context required"}`, http.StatusUnauthorized)
		return
	}
	var req models.GenerateEInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	req.OrgID = orgID

	inv, err := h.eInvoice.GenerateEInvoice(r.Context(), req)
	if err != nil {
		h.logger.Error("e-invoice generation failed", zap.Error(err))
		http.Error(w, `{"error":"failed to generate e-invoice"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inv)
}

// GetEInvoice handles GET /gst/einvoice/{id}
func (h *EnterpriseFinanceHandlers) GetEInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	inv, err := h.eInvoice.GetEInvoice(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inv)
}

// DeductTDS handles POST /tds/deduct
func (h *EnterpriseFinanceHandlers) DeductTDS(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, `{"error":"unauthorized: organization context required"}`, http.StatusUnauthorized)
		return
	}
	var req models.DeductTDSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	req.OrgID = orgID

	record, err := h.tds.DeductTDS(r.Context(), req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

// GetTDSReport handles GET /tds/report?fy=2024-25&quarter=Q1
func (h *EnterpriseFinanceHandlers) GetTDSReport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, `{"error":"unauthorized: organization context required"}`, http.StatusUnauthorized)
		return
	}
	fy := r.URL.Query().Get("fy")
	quarter := r.URL.Query().Get("quarter")

	records, err := h.tds.GetTDSReport(r.Context(), orgID, fy, quarter)
	if err != nil {
		http.Error(w, `{"error":"failed to get TDS report"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"records": records,
		"total":   len(records),
	})
}

// CreateVendorPayment handles POST /vendors/payments
func (h *EnterpriseFinanceHandlers) CreateVendorPayment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, `{"error":"unauthorized: organization context required"}`, http.StatusUnauthorized)
		return
	}
	createdBy := orgID // simplified; in production use JWT claims

	var req models.CreateVendorPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	payment, err := h.vendorPay.CreatePayment(r.Context(), orgID, createdBy, req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(payment)
}

// ListVendorPayments handles GET /vendors/payments
func (h *EnterpriseFinanceHandlers) ListVendorPayments(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, `{"error":"unauthorized: organization context required"}`, http.StatusUnauthorized)
		return
	}

	var vendorID *uuid.UUID
	if v := r.URL.Query().Get("vendor_id"); v != "" {
		id, err := uuid.Parse(v)
		if err == nil {
			vendorID = &id
		}
	}

	payments, err := h.vendorPay.ListPayments(r.Context(), orgID, vendorID)
	if err != nil {
		http.Error(w, `{"error":"failed to list payments"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payments": payments,
		"total":    len(payments),
	})
}

// GetTripProfit handles GET /trips/{id}/profit
func (h *EnterpriseFinanceHandlers) GetTripProfit(w http.ResponseWriter, r *http.Request) {
	tripID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid trip id"}`, http.StatusBadRequest)
		return
	}
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, `{"error":"unauthorized: organization context required"}`, http.StatusUnauthorized)
		return
	}

	result, err := h.tripProfit.CalculateTripProfit(r.Context(), tripID, orgID)
	if err != nil {
		http.Error(w, `{"error":"failed to calculate trip profit"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetFinanceDashboard handles GET /finance/dashboard
func (h *EnterpriseFinanceHandlers) GetFinanceDashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, `{"error":"unauthorized: organization context required"}`, http.StatusUnauthorized)
		return
	}

	metrics, err := h.tripProfit.GetDashboardMetrics(r.Context(), orgID)
	if err != nil {
		http.Error(w, `{"error":"failed to get dashboard metrics"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// RunReconciliation handles POST /reconciliation/run
func (h *EnterpriseFinanceHandlers) RunReconciliation(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, `{"error":"unauthorized: organization context required"}`, http.StatusUnauthorized)
		return
	}

	matched, unmatched, err := h.reconSvc.RunReconciliation(r.Context(), orgID)
	if err != nil {
		http.Error(w, `{"error":"reconciliation failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"matched":   matched,
		"unmatched": unmatched,
	})
}

// ManualMatch handles POST /reconciliation/{id}/match
func (h *EnterpriseFinanceHandlers) ManualMatch(w http.ResponseWriter, r *http.Request) {
	reconID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		InternalTransactionID uuid.UUID `json:"internal_transaction_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, `{"error":"unauthorized: organization context required"}`, http.StatusUnauthorized)
		return
	}
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		h.logger.Warn("Missing user_id in context - request likely bypassed authentication middleware")
		http.Error(w, `{"error":"unauthorized: user context required"}`, http.StatusUnauthorized)
		return
	}
	if err := h.reconSvc.ManualMatch(r.Context(), reconID, body.InternalTransactionID, userID); err != nil {
		http.Error(w, `{"error":"failed to match"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
