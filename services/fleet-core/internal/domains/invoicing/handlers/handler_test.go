package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/middleware"
	"github.com/flyfleet/fleet-core/internal/domains/invoicing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockInvoicingService struct {
	mock.Mock
}

func (m *mockInvoicingService) CreateInvoice(ctx context.Context, orgID uuid.UUID, inv *invoicing.Invoice) (*invoicing.Invoice, error) {
	args := m.Called(ctx, orgID, inv)
	return args.Get(0).(*invoicing.Invoice), args.Error(1)
}

func (m *mockInvoicingService) GetInvoice(ctx context.Context, orgID, id uuid.UUID) (*invoicing.Invoice, error) {
	args := m.Called(ctx, orgID, id)
	return args.Get(0).(*invoicing.Invoice), args.Error(1)
}

func (m *mockInvoicingService) ListInvoices(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]invoicing.Invoice, int64, error) {
	args := m.Called(ctx, orgID, limit, offset)
	return args.Get(0).([]invoicing.Invoice), args.Get(1).(int64), args.Error(2)
}

func (m *mockInvoicingService) VoidInvoice(ctx context.Context, orgID, id uuid.UUID) error {
	args := m.Called(ctx, orgID, id)
	return args.Error(0)
}

func TestCreateHandler(t *testing.T) {
	svc := new(mockInvoicingService)
	h := NewHandler(svc)
	orgID := uuid.New()

	reqBody := invoicing.CreateInvoiceRequest{
		CustomerName:   "Test Client",
		BillingAddress: "123 Road",
		DueDate:        time.Now().AddDate(0, 0, 7),
		Currency:       "INR",
		Items: []invoicing.CreateInvoiceItem{
			{Description: "Service 1", Quantity: 1, UnitPrice: 1000},
		},
	}

	svc.On("CreateInvoice", mock.Anything, orgID, mock.Anything).Return(&invoicing.Invoice{ID: uuid.New()}, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	ctx := context.WithValue(req.Context(), middleware.OrgIDKey, orgID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	svc.AssertExpectations(t)
}

func TestCreateHandler_InvalidJSON(t *testing.T) {
	svc := new(mockInvoicingService)
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{invalid}"))
	ctx := context.WithValue(req.Context(), middleware.OrgIDKey, uuid.New())
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateHandler_NoOrg(t *testing.T) {
	svc := new(mockInvoicingService)
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
