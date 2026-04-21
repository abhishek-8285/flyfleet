package service

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/customer/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) CreateCustomer(ctx context.Context, c *models.Customer) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *mockRepo) GetCustomerByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Customer), args.Error(1)
}

func (m *mockRepo) ListCustomers(ctx context.Context, tenantID uuid.UUID, filters map[string]interface{}, limit, offset int) ([]models.Customer, error) {
	args := m.Called(ctx, tenantID, filters, limit, offset)
	return args.Get(0).([]models.Customer), args.Error(1)
}

func (m *mockRepo) CountCustomers(ctx context.Context, tenantID uuid.UUID, filters map[string]interface{}) (int, error) {
	args := m.Called(ctx, tenantID, filters)
	return args.Int(0), args.Error(1)
}

func (m *mockRepo) CreateBookingRequest(ctx context.Context, b *models.BookingRequest) error { return nil }
func (m *mockRepo) GetBookingRequestByID(ctx context.Context, id uuid.UUID) (*models.BookingRequest, error) { return nil, nil }
func (m *mockRepo) CreateShipment(ctx context.Context, s *models.Shipment) error { return nil }
func (m *mockRepo) GetShipmentByID(ctx context.Context, id uuid.UUID) (*models.Shipment, error) { return nil, nil }
func (m *mockRepo) GetShipmentTracking(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) { return nil, nil }

// Implement other interface methods to satisfy the interface
func (m *mockRepo) GetDashboardMetrics(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error) { return nil, nil }
func (m *mockRepo) GetAnalytics(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error) { return nil, nil }
func (m *mockRepo) SearchShipments(ctx context.Context, customerID uuid.UUID, filters map[string]interface{}) ([]models.Shipment, error) { return nil, nil }
func (m *mockRepo) UpdateCustomer(ctx context.Context, c *models.Customer) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func TestListCustomers(t *testing.T) {
	repo := new(mockRepo)
	svc := NewCustomerService(repo)
	tenantID := uuid.New()
	ctx := context.Background()
	filters := map[string]interface{}{}

	expectedCustomers := []models.Customer{
		{ID: uuid.New(), Name: "Test Corp", TenantID: tenantID},
	}

	repo.On("ListCustomers", ctx, tenantID, filters, 10, 0).Return(expectedCustomers, nil)

	customers, err := svc.ListCustomers(ctx, tenantID, filters, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, expectedCustomers, customers)
	repo.AssertExpectations(t)
}

func TestListCustomers_Error(t *testing.T) {
	repo := new(mockRepo)
	svc := NewCustomerService(repo)
	tenantID := uuid.New()
	ctx := context.Background()
	filters := map[string]interface{}{}

	repo.On("ListCustomers", ctx, tenantID, filters, 10, 0).Return([]models.Customer{}, assert.AnError)

	customers, err := svc.ListCustomers(ctx, tenantID, filters, 10, 0)

	assert.Error(t, err)
	assert.Empty(t, customers)
	repo.AssertExpectations(t)
}

func TestCountCustomers(t *testing.T) {
	repo := new(mockRepo)
	svc := NewCustomerService(repo)
	tenantID := uuid.New()
	ctx := context.Background()
	filters := map[string]interface{}{}

	repo.On("CountCustomers", ctx, tenantID, filters).Return(1, nil)

	count, err := svc.CountCustomers(ctx, tenantID, filters)

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	repo.AssertExpectations(t)
}
