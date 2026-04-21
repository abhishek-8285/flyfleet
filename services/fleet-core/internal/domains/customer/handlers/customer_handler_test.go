package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCustomerService struct {
	mock.Mock
}

func (m *mockCustomerService) CreateCustomer(ctx context.Context, c *models.Customer) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *mockCustomerService) GetCustomerByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Customer), args.Error(1)
}

func (m *mockCustomerService) ListCustomers(ctx context.Context, tenantID uuid.UUID, filters map[string]interface{}, limit, offset int) ([]models.Customer, error) {
	args := m.Called(ctx, tenantID, filters, limit, offset)
	return args.Get(0).([]models.Customer), args.Error(1)
}

func (m *mockCustomerService) CountCustomers(ctx context.Context, tenantID uuid.UUID, filters map[string]interface{}) (int, error) {
	args := m.Called(ctx, tenantID, filters)
	return args.Int(0), args.Error(1)
}

func (m *mockCustomerService) GetCustomerSummary(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockCustomerService) UpdateCustomer(ctx context.Context, c *models.Customer) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func TestListCustomersHandler(t *testing.T) {
	svc := new(mockCustomerService)
	h := &CustomerV2Handler{
		customerService: svc,
	}

	tenantID := uuid.New()
	customers := []models.Customer{
		{ID: uuid.New(), Name: "Acme Corp", TenantID: tenantID},
	}

	filters := map[string]interface{}{}
	svc.On("ListCustomers", mock.Anything, tenantID, filters, 20, 0).Return(customers, nil)
	svc.On("CountCustomers", mock.Anything, tenantID, filters).Return(1, nil)

	req := httptest.NewRequest(http.MethodGet, "/customers?page=1&size=20", nil)
	ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.ListCustomers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var outer struct {
		Data []models.Customer `json:"data"`
		Meta map[string]int    `json:"meta"`
	}
	json.Unmarshal(rr.Body.Bytes(), &outer)
	
	assert.Len(t, outer.Data, 1)
	assert.Equal(t, "Acme Corp", outer.Data[0].Name)
	assert.Equal(t, 1, outer.Meta["total"])
	svc.AssertExpectations(t)
}

func TestListCustomersHandler_NoOrg(t *testing.T) {
	svc := new(mockCustomerService)
	h := &CustomerV2Handler{
		customerService: svc,
	}

	req := httptest.NewRequest(http.MethodGet, "/customers", nil)
	rr := httptest.NewRecorder()

	h.ListCustomers(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestListCustomersHandler_Error(t *testing.T) {
	svc := new(mockCustomerService)
	h := &CustomerV2Handler{
		customerService: svc,
	}

	tenantID := uuid.New()
	filters := map[string]interface{}{}
	svc.On("ListCustomers", mock.Anything, tenantID, filters, 20, 0).Return([]models.Customer{}, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/customers?page=1&size=20", nil)
	ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.ListCustomers(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}
