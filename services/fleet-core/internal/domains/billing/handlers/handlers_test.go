package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ─── Test Helpers ───────────────────────────────────────────────────────────

type mockBillingService struct {
	invoices  []*models.Invoice
	settings  *models.BillingSettings
	metrics   *models.BillingMetrics
	listErr   error
	getErr    error
	createErr error
	updateErr error
	deleteErr error
	payErr    error
}

func (m *mockBillingService) ListInvoices(_ interface{}, orgID uuid.UUID, page, size int) ([]models.Invoice, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var result []models.Invoice
	for _, inv := range m.invoices {
		if inv.OrgID == orgID {
			result = append(result, *inv)
		}
	}
	return result, len(result), nil
}

func (m *mockBillingService) CreateInvoice(_ interface{}, inv *models.Invoice) error {
	if m.createErr != nil {
		return m.createErr
	}
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	m.invoices = append(m.invoices, inv)
	return nil
}

func (m *mockBillingService) GetInvoice(_ interface{}, id uuid.UUID) (*models.Invoice, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, inv := range m.invoices {
		if inv.ID == id {
			return inv, nil
		}
	}
	return nil, nil
}

func (m *mockBillingService) UpdateInvoice(_ interface{}, inv *models.Invoice) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	for i, existing := range m.invoices {
		if existing.ID == inv.ID {
			m.invoices[i] = inv
			return nil
		}
	}
	return service.ErrInvoiceNotFound
}

func (m *mockBillingService) DeleteInvoice(_ interface{}, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for i, inv := range m.invoices {
		if inv.ID == id {
			m.invoices = append(m.invoices[:i], m.invoices[i+1:]...)
			return nil
		}
	}
	return service.ErrInvoiceNotFound
}

func (m *mockBillingService) RecordPayment(_ interface{}, id uuid.UUID, amount float64) error {
	if m.payErr != nil {
		return m.payErr
	}
	for _, inv := range m.invoices {
		if inv.ID == id {
			inv.AmountPaid += amount
			inv.BalanceDue = inv.TotalAmount - inv.AmountPaid
			return nil
		}
	}
	return service.ErrInvoiceNotFound
}

func (m *mockBillingService) GetSettings(_ interface{}, orgID uuid.UUID) (*models.BillingSettings, error) {
	if m.settings == nil {
		return &models.BillingSettings{
			OrgID:                orgID,
			DefaultTaxRate:       18.0,
			BaseCurrency:         "USD",
			GracePeriodDays:      14,
			AutoInvoicingEnabled: false,
		}, nil
	}
	return m.settings, nil
}

func (m *mockBillingService) UpdateSettings(_ interface{}, s *models.BillingSettings) error {
	m.settings = s
	return nil
}

func (m *mockBillingService) GetMetrics(_ interface{}, _ uuid.UUID) (*models.BillingMetrics, error) {
	if m.metrics == nil {
		return &models.BillingMetrics{}, nil
	}
	return m.metrics, nil
}

func newTestHandler(svc *mockBillingService) *BillingHandler {
	return NewBillingHandler(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		zap.NewNop(),
	)
}

// ─── Test ListInvoices Handler ──────────────────────────────────────────────

func TestListInvoicesHandler(t *testing.T) {
	orgID := uuid.New()

	tests := []struct {
		name       string
		invoices   []*models.Invoice
		query      string
		wantStatus int
		wantCount  int
	}{
		{
			name: "returns empty list",
			invoices: []*models.Invoice{},
			query:    "/api/v1/billing/invoices",
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name: "returns invoices for org",
			invoices: []*models.Invoice{
				{ID: uuid.New(), OrgID: orgID, CustomerName: "Acme", InvoiceDate: parseTime("2026-04-01"), DueDate: parseTime("2026-05-01"), Status: "PENDING", Currency: "INR", TotalAmount: 10000, AmountPaid: 0, BalanceDue: 10000},
			},
			query:      "/api/v1/billing/invoices",
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockBillingService{invoices: tt.invoices}
			h := newTestHandler(svc)

			req := httptest.NewRequest(http.MethodGet, tt.query, nil)
			rr := httptest.NewRecorder()

			h.ListInvoices(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("ListInvoices() status = %d, want %d", rr.Code, tt.wantStatus)
			}

			if tt.wantCount > 0 {
				var resp map[string]interface{}
				json.Unmarshal(rr.Body.Bytes(), &resp)
				if data, ok := resp["data"].([]interface{}); ok {
					if len(data) != tt.wantCount {
						t.Errorf("ListInvoices() returned %d items, want %d", len(data), tt.wantCount)
					}
				}
			}
		})
	}
}

// ─── Test CreateInvoice Handler ─────────────────────────────────────────────

func TestCreateInvoiceHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{
			name: "valid invoice creation",
			body: map[string]interface{}{
				"customer_name": "Acme Corp",
				"invoice_date":  "2026-04-01",
				"due_date":      "2026-05-01",
				"currency":      "INR",
				"items": []map[string]interface{}{
					{"description": "Fleet Management", "quantity": 1, "unit_price": 5000.0, "tax_percent": 18.0},
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid JSON body",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockBillingService{}
			h := newTestHandler(svc)

			var bodyBytes []byte
			if str, ok := tt.body.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/invoices", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			h.CreateInvoice(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("CreateInvoice() status = %d, want %d, body = %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}
}

// ─── Test GetInvoice Handler ────────────────────────────────────────────────

func TestGetInvoiceHandler(t *testing.T) {
	invID := uuid.New()
	svc := &mockBillingService{
		invoices: []*models.Invoice{
			{ID: invID, OrgID: uuid.New(), CustomerName: "Acme", InvoiceDate: parseTime("2026-04-01"), DueDate: parseTime("2026-05-01"), Status: "PENDING", Currency: "INR", TotalAmount: 10000},
		},
	}
	h := newTestHandler(svc)

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{"valid UUID", invID.String(), http.StatusOK},
		{"invalid UUID", "not-a-uuid", http.StatusBadRequest},
		{"non-existent UUID", uuid.New().String(), http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/billing/invoices/"+tt.id, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.GetInvoice(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("GetInvoice() status = %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

// ─── Test RecordPayment Handler ─────────────────────────────────────────────

func TestRecordPaymentHandler(t *testing.T) {
	invID := uuid.New()
	svc := &mockBillingService{
		invoices: []*models.Invoice{
			{ID: invID, OrgID: uuid.New(), CustomerName: "Acme", InvoiceDate: parseTime("2026-04-01"), DueDate: parseTime("2026-05-01"), Status: "PENDING", Currency: "INR", TotalAmount: 10000, AmountPaid: 0, BalanceDue: 10000},
		},
	}
	h := newTestHandler(svc)

	tests := []struct {
		name       string
		id         string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "valid payment",
			id:         invID.String(),
			body:       map[string]interface{}{"amount": 5000.0},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid UUID",
			id:         "not-a-uuid",
			body:       map[string]interface{}{"amount": 5000.0},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing amount",
			id:         invID.String(),
			body:       map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "negative amount",
			id:         invID.String(),
			body:       map[string]interface{}{"amount": -100.0},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/invoices/"+tt.id+"/payments", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.RecordPayment(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("RecordPayment() status = %d, want %d, body = %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}
}

// ─── Test GetSettings Handler ───────────────────────────────────────────────

func TestGetSettingsHandler(t *testing.T) {
	svc := &mockBillingService{}
	h := newTestHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/billing/settings", nil)
	rr := httptest.NewRecorder()

	h.GetSettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetSettings() status = %d, want %d", rr.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if _, ok := resp["default_tax_rate"]; !ok {
		t.Error("GetSettings() missing default_tax_rate")
	}
}

// ─── Test GetMetrics Handler ────────────────────────────────────────────────

func TestGetMetricsHandler(t *testing.T) {
	svc := &mockBillingService{
		metrics: &models.BillingMetrics{
			TotalInvoicedMTD: 50000.0,
			PendingPayment:   20000.0,
			OverdueAmount:    5000.0,
			PaidCount:        10,
			OpenCount:        5,
		},
	}
	h := newTestHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/billing/metrics", nil)
	rr := httptest.NewRecorder()

	h.GetMetrics(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GetMetrics() status = %d, want %d", rr.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if total, ok := resp["total_invoiced_mtd"].(float64); !ok || total != 50000.0 {
		t.Errorf("GetMetrics() total_invoiced_mtd = %v, want 50000.0", resp["total_invoiced_mtd"])
	}
}

// ─── Test DeleteInvoice Handler ─────────────────────────────────────────────

func TestDeleteInvoiceHandler(t *testing.T) {
	invID := uuid.New()
	svc := &mockBillingService{
		invoices: []*models.Invoice{
			{ID: invID, OrgID: uuid.New(), CustomerName: "Acme", InvoiceDate: parseTime("2026-04-01"), DueDate: parseTime("2026-05-01"), Status: "DRAFT", Currency: "INR", TotalAmount: 10000},
		},
	}
	h := newTestHandler(svc)

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{"valid delete", invID.String(), http.StatusNoContent},
		{"invalid UUID", "not-a-uuid", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/billing/invoices/"+tt.id, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.DeleteInvoice(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("DeleteInvoice() status = %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func parseTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}
