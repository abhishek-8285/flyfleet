package service

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/invoicing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockInvoiceRepo struct {
	mock.Mock
}

func (m *mockInvoiceRepo) Create(ctx context.Context, inv *invoicing.Invoice) (*invoicing.Invoice, error) {
	args := m.Called(ctx, inv)
	return args.Get(0).(*invoicing.Invoice), args.Error(1)
}

func (m *mockInvoiceRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*invoicing.Invoice, error) {
	args := m.Called(ctx, tenantID, id)
	return args.Get(0).(*invoicing.Invoice), args.Error(1)
}

func (m *mockInvoiceRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]invoicing.Invoice, int64, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	return args.Get(0).([]invoicing.Invoice), args.Get(1).(int64), args.Error(2)
}

func (m *mockInvoiceRepo) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status invoicing.InvoiceStatus, updatedBy uuid.UUID) error {
	args := m.Called(ctx, tenantID, id, status, updatedBy)
	return args.Error(0)
}

func (m *mockInvoiceRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func TestCreateInvoice(t *testing.T) {
	repo := new(mockInvoiceRepo)
	svc := NewInvoicingService(repo)
	tenantID := uuid.New()
	ctx := context.Background()

	inv := &invoicing.Invoice{
		CustomerName:  "Acme",
		CustomerPhone: "+91 98765 43210",
		DueDate:       time.Now().AddDate(0, 0, 7),
		Currency:      "INR",
		Items: []invoicing.InvoiceItem{
			{Description: "Item 1", Quantity: 2, UnitPrice: 100},
		},
	}

	repo.On("Create", ctx, mock.MatchedBy(func(i *invoicing.Invoice) bool {
		return i.TenantID == tenantID && i.TotalAmount == 236.0 // 200 + 18% tax
	})).Return(inv, nil)

	result, err := svc.CreateInvoice(ctx, tenantID, inv)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 236.0, inv.TotalAmount)
	assert.Equal(t, 36.0, inv.TaxAmount)
	repo.AssertExpectations(t)
}

func TestGetInvoice(t *testing.T) {
	repo := new(mockInvoiceRepo)
	svc := NewInvoicingService(repo)
	tenantID := uuid.New()
	id := uuid.New()
	ctx := context.Background()

	expected := &invoicing.Invoice{ID: id, TenantID: tenantID}
	repo.On("GetByID", ctx, tenantID, id).Return(expected, nil)

	result, err := svc.GetInvoice(ctx, tenantID, id)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestVoidInvoice(t *testing.T) {
	repo := new(mockInvoiceRepo)
	svc := NewInvoicingService(repo)
	tenantID := uuid.New()
	id := uuid.New()
	ctx := context.Background()

	repo.On("UpdateStatus", ctx, tenantID, id, invoicing.StatusVoid, uuid.Nil).Return(nil)

	err := svc.VoidInvoice(ctx, tenantID, id)

	assert.NoError(t, err)
}

func TestCreateInvoice_MissingPhone(t *testing.T) {
	repo := new(mockInvoiceRepo)
	svc := NewInvoicingService(repo)
	tenantID := uuid.New()
	ctx := context.Background()

	inv := &invoicing.Invoice{
		CustomerName: "Acme",
		CustomerPhone: "", // Empty phone
		DueDate:      time.Now().AddDate(0, 0, 7),
		Currency:     "INR",
		Items: []invoicing.InvoiceItem{
			{Description: "Item 1", Quantity: 2, UnitPrice: 100},
		},
	}

	result, err := svc.CreateInvoice(ctx, tenantID, inv)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "customer phone number is mandatory")
}
