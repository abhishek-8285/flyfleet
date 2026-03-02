package service

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockRepository is a mock of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateInvoice(ctx context.Context, inv *models.Invoice) error {
	args := m.Called(ctx, inv)
	return args.Error(0)
}

func (m *MockRepository) ListInvoices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Invoice, int, error) {
	args := m.Called(ctx, orgID, page, size)
	return args.Get(0).([]models.Invoice), args.Int(1), args.Error(2)
}

func (m *MockRepository) GetInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invoice), args.Error(1)
}

func (m *MockRepository) UpdateInvoice(ctx context.Context, inv *models.Invoice) error {
	args := m.Called(ctx, inv)
	return args.Error(0)
}

func (m *MockRepository) DeleteInvoice(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) RecordPayment(ctx context.Context, id uuid.UUID, amount float64) error {
	args := m.Called(ctx, id, amount)
	return args.Error(0)
}

func (m *MockRepository) GetSettings(ctx context.Context, orgID uuid.UUID) (*models.BillingSettings, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BillingSettings), args.Error(1)
}

func (m *MockRepository) UpdateSettings(ctx context.Context, s *models.BillingSettings) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockRepository) GetMetrics(ctx context.Context, orgID uuid.UUID) (*models.BillingMetrics, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).(*models.BillingMetrics), args.Error(1)
}

func TestCalculateInvoiceTotals(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(MockRepository)
	svc := NewBillingService(mockRepo, logger)

	inv := &models.Invoice{
		Items: []models.InvoiceItem{
			{Quantity: 2, UnitPrice: 100, TaxPercent: 10}, // Total: 220
			{Quantity: 1, UnitPrice: 500, TaxPercent: 18}, // Total: 590
		},
	}

	mockRepo.On("CreateInvoice", mock.Anything, inv).Return(nil)

	err := svc.CreateInvoice(context.Background(), inv)

	assert.NoError(t, err)
	assert.Equal(t, 810.0, inv.TotalAmount)
	assert.Equal(t, 220.0, inv.Items[0].Total)
	assert.Equal(t, 590.0, inv.Items[1].Total)
	mockRepo.AssertExpectations(t)
}

func TestGetSettings_Fallback(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(MockRepository)
	svc := NewBillingService(mockRepo, logger)

	orgID := uuid.New()
	mockRepo.On("GetSettings", mock.Anything, orgID).Return(nil, nil)

	settings, err := svc.GetSettings(context.Background(), orgID)

	assert.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, 18.0, settings.DefaultTaxRate)
	assert.Equal(t, "USD", settings.BaseCurrency)
	mockRepo.AssertExpectations(t)
}
