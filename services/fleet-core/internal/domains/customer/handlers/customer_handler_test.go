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

func (m *mockCustomerService) ListCustomers(ctx context.Context, orgID uuid.UUID) ([]models.Customer, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]models.Customer), args.Error(1)
}

func TestListCustomersHandler(t *testing.T) {
	svc := new(mockCustomerService)
	h := &CustomerV2Handler{
		customerService: svc,
	}

	orgID := uuid.New()
	customers := []models.Customer{
		{ID: uuid.New(), CompanyName: "Acme Corp", OrganizationID: orgID},
	}

	svc.On("ListCustomers", mock.Anything, orgID).Return(customers, nil)

	req := httptest.NewRequest(http.MethodGet, "/customers", nil)
	ctx := context.WithValue(req.Context(), middleware.OrgIDKey, orgID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.ListCustomers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp []models.Customer
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, "Acme Corp", resp[0].CompanyName)
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

	orgID := uuid.New()
	svc.On("ListCustomers", mock.Anything, orgID).Return([]models.Customer{}, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/customers", nil)
	ctx := context.WithValue(req.Context(), middleware.OrgIDKey, orgID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.ListCustomers(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}
