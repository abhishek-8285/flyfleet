package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core/internal/platform/billing/models"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBillingService struct {
	mock.Mock
}

func (m *MockBillingService) CheckLimit(ctx context.Context, tenantID uuid.UUID, metric string) (bool, error) {
	args := m.Called(ctx, tenantID, metric)
	return args.Bool(0), args.Error(1)
}

func (m *MockBillingService) GetOrganizationPlan(ctx context.Context, tenantID uuid.UUID) (*models.Plan, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Plan), args.Error(1)
}

func TestPlanLimitMiddleware(t *testing.T) {
	mockSvc := new(MockBillingService)
	handler := PlanLimitMiddleware(mockSvc, "vehicles")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tenantID := uuid.New()

	// Case: Allowed
	mockSvc.On("CheckLimit", mock.Anything, tenantID, "vehicles").Return(true, nil).Once()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), ctxkeys.TenantIDKey, tenantID.String())
	rr := httptest.NewRecorder()

	handler(next).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case: Forbidden
	mockSvc.On("CheckLimit", mock.Anything, tenantID, "vehicles").Return(false, nil).Once()
	rr = httptest.NewRecorder()
	handler(next).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusForbidden, rr.Code)
}
