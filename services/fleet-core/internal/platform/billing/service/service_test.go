package service

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core/internal/platform/billing/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockBillingRepository struct {
	mock.Mock
}

func (m *MockBillingRepository) GetSubscriptionByTenant(ctx context.Context, tenantID uuid.UUID) (*models.Subscription, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockBillingRepository) GetPlanByID(ctx context.Context, planID uuid.UUID) (*models.Plan, error) {
	args := m.Called(ctx, planID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Plan), args.Error(1)
}

func (m *MockBillingRepository) RecordUsage(ctx context.Context, record *models.UsageRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockBillingRepository) GetUsage(ctx context.Context, tenantID uuid.UUID, metric string) (int, error) {
	args := m.Called(ctx, tenantID, metric)
	return args.Int(0), args.Error(1)
}

func TestBillingService_CheckLimit(t *testing.T) {
	repo := new(MockBillingRepository)
	logger, _ := zap.NewDevelopment()
	svc := NewBillingService(repo, logger)

	tenantID := uuid.New()
	planID := uuid.New()

	repo.On("GetSubscriptionByTenant", mock.Anything, tenantID).Return(&models.Subscription{PlanID: planID}, nil)
	repo.On("GetPlanByID", mock.Anything, planID).Return(&models.Plan{MaxVehicles: 10, MaxDrivers: 5}, nil)

	// Case: Under limit
	repo.On("GetUsage", mock.Anything, tenantID, "vehicles").Return(5, nil).Once()
	allowed, err := svc.CheckLimit(context.Background(), tenantID, "vehicles")
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Case: At limit
	repo.On("GetUsage", mock.Anything, tenantID, "vehicles").Return(10, nil).Once()
	allowed, err = svc.CheckLimit(context.Background(), tenantID, "vehicles")
	assert.NoError(t, err)
	assert.False(t, allowed)
}
