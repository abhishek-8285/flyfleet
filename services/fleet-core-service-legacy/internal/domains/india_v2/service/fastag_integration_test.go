package service

import (
	"context"
	"testing"
	"time"

	billingv2service "github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/service"
	"github.com/flyfleet/fleet-core-service/internal/domains/india_v2/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockIndiaRepository is a mock for IndiaIntegrationRepository
type MockIndiaRepository struct {
	mock.Mock
}

func (m *MockIndiaRepository) CreateFASTagTransaction(ctx context.Context, tx *models.FASTagTransaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockIndiaRepository) CreateIntegrationLog(ctx context.Context, log *models.IntegrationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

// Implement other methods as needed or as no-ops
func (m *MockIndiaRepository) GetFASTagTransaction(ctx context.Context, id string) (*models.FASTagTransaction, error) { return nil, nil }
func (m *MockIndiaRepository) UpdateFASTagTransaction(ctx context.Context, id string, updates map[string]interface{}) error { return nil }
func (m *MockIndiaRepository) DeleteFASTagTransaction(ctx context.Context, id string) error { return nil }
func (m *MockIndiaRepository) ListFASTagTransactions(ctx context.Context, vehicleID *uuid.UUID, tagNumber *string, startDate, endDate *time.Time, limit, offset int) ([]*models.FASTagTransaction, int, error) { return nil, 0, nil }
func (m *MockIndiaRepository) CreateEWayBill(ctx context.Context, bill *models.EWayBill) error { return nil }
func (m *MockIndiaRepository) GetEWayBill(ctx context.Context, id string) (*models.EWayBill, error) { return nil, nil }
func (m *MockIndiaRepository) GetEWayBillByNumber(ctx context.Context, number string) (*models.EWayBill, error) { return nil, nil }
func (m *MockIndiaRepository) UpdateEWayBill(ctx context.Context, id string, updates map[string]interface{}) error { return nil }
func (m *MockIndiaRepository) DeleteEWayBill(ctx context.Context, id string) error { return nil }
func (m *MockIndiaRepository) ListEWayBills(ctx context.Context, vehicleNumber *string, transporterID *string, startDate, endDate *time.Time, limit, offset int) ([]*models.EWayBill, int, error) { return nil, 0, nil }
func (m *MockIndiaRepository) CreateVehicleVerification(ctx context.Context, verification *models.VehicleVerification) error { return nil }
func (m *MockIndiaRepository) GetVehicleVerification(ctx context.Context, id string) (*models.VehicleVerification, error) { return nil, nil }
func (m *MockIndiaRepository) GetVehicleVerificationByRegNumber(ctx context.Context, regNumber string) (*models.VehicleVerification, error) { return nil, nil }
func (m *MockIndiaRepository) UpdateVehicleVerification(ctx context.Context, id string, updates map[string]interface{}) error { return nil }
func (m *MockIndiaRepository) DeleteVehicleVerification(ctx context.Context, id string) error { return nil }
func (m *MockIndiaRepository) ListVehicleVerifications(ctx context.Context, regNumber *string, status *string, limit, offset int) ([]*models.VehicleVerification, int, error) { return nil, 0, nil }
func (m *MockIndiaRepository) CreateDriverVerification(ctx context.Context, verification *models.DriverVerification) error { return nil }
func (m *MockIndiaRepository) GetDriverVerification(ctx context.Context, id string) (*models.DriverVerification, error) { return nil, nil }
func (m *MockIndiaRepository) GetDriverVerificationByLicenseNumber(ctx context.Context, licenseNumber string) (*models.DriverVerification, error) { return nil, nil }
func (m *MockIndiaRepository) UpdateDriverVerification(ctx context.Context, id string, updates map[string]interface{}) error { return nil }
func (m *MockIndiaRepository) DeleteDriverVerification(ctx context.Context, id string) error { return nil }
func (m *MockIndiaRepository) ListDriverVerifications(ctx context.Context, licenseNumber *string, status *string, limit, offset int) ([]*models.DriverVerification, int, error) { return nil, 0, nil }
func (m *MockIndiaRepository) CreateLocalPermit(ctx context.Context, permit *models.LocalPermit) error { return nil }
func (m *MockIndiaRepository) GetLocalPermit(ctx context.Context, id string) (*models.LocalPermit, error) { return nil, nil }
func (m *MockIndiaRepository) GetLocalPermitByNumber(ctx context.Context, permitNumber string) (*models.LocalPermit, error) { return nil, nil }
func (m *MockIndiaRepository) UpdateLocalPermit(ctx context.Context, id string, updates map[string]interface{}) error { return nil }
func (m *MockIndiaRepository) DeleteLocalPermit(ctx context.Context, id string) error { return nil }
func (m *MockIndiaRepository) ListLocalPermits(ctx context.Context, vehicleID *uuid.UUID, permitType *string, status *string, limit, offset int) ([]*models.LocalPermit, int, error) { return nil, 0, nil }
func (m *MockIndiaRepository) CreateGSTCompliance(ctx context.Context, compliance *models.GSTCompliance) error { return nil }
func (m *MockIndiaRepository) GetGSTCompliance(ctx context.Context, id string) (*models.GSTCompliance, error) { return nil, nil }
func (m *MockIndiaRepository) GetGSTComplianceByGSTIN(ctx context.Context, gstin string) (*models.GSTCompliance, error) { return nil, nil }
func (m *MockIndiaRepository) UpdateGSTCompliance(ctx context.Context, id string, updates map[string]interface{}) error { return nil }
func (m *MockIndiaRepository) DeleteGSTCompliance(ctx context.Context, id string) error { return nil }
func (m *MockIndiaRepository) ListGSTCompliances(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]*models.GSTCompliance, int, error) { return nil, 0, nil }
func (m *MockIndiaRepository) CreateIntegrationConfig(ctx context.Context, config *models.IntegrationConfig) error { return nil }
func (m *MockIndiaRepository) GetIntegrationConfig(ctx context.Context, id string) (*models.IntegrationConfig, error) { return nil, nil }
func (m *MockIndiaRepository) GetIntegrationConfigByServiceType(ctx context.Context, orgID uuid.UUID, serviceType string) (*models.IntegrationConfig, error) { return nil, nil }
func (m *MockIndiaRepository) UpdateIntegrationConfig(ctx context.Context, id string, updates map[string]interface{}) error { return nil }
func (m *MockIndiaRepository) DeleteIntegrationConfig(ctx context.Context, id string) error { return nil }
func (m *MockIndiaRepository) ListIntegrationConfigs(ctx context.Context, orgID uuid.UUID, serviceType *string, active *bool, limit, offset int) ([]*models.IntegrationConfig, int, error) { return nil, 0, nil }
func (m *MockIndiaRepository) HealthCheck(ctx context.Context) error { return nil }
func (m *MockIndiaRepository) Ping(ctx context.Context) error { return nil }
func (m *MockIndiaRepository) BulkCreateVehicleVerification(ctx context.Context, verifications []*models.VehicleVerification) error {
	return nil
}
func (m *MockIndiaRepository) BulkCreateDriverVerification(ctx context.Context, verifications []*models.DriverVerification) error {
	return nil
}
func (m *MockIndiaRepository) GetExpiringVehicleVerifications(ctx context.Context, days int) ([]*models.VehicleVerification, error) {
	return nil, nil
}
func (m *MockIndiaRepository) GetExpiringDriverVerifications(ctx context.Context, days int) ([]*models.DriverVerification, error) {
	return nil, nil
}
func (m *MockIndiaRepository) GetExpiringLocalPermits(ctx context.Context, days int) ([]*models.LocalPermit, error) {
	return nil, nil
}
func (m *MockIndiaRepository) GetNonCompliantGSTRecords(ctx context.Context) ([]*models.GSTCompliance, error) {
	return nil, nil
}

func TestFastagIntegrationService_FetchTransactions(t *testing.T) {
	mockRepo := new(MockIndiaRepository)
	logger := zap.NewNop()
	detector := billingv2service.NewDoubleSwipeDetector(mockRepo, nil, logger)
	svc := NewFastagIntegrationService(mockRepo, detector, logger)

	orgID := uuid.New()
	tagNumber := "TAG123"
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	mockRepo.On("CreateIntegrationLog", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("CreateFASTagTransaction", mock.Anything, mock.Anything).Return(nil)

	txs, err := svc.FetchTransactions(context.Background(), orgID, tagNumber, from, to)

	assert.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, tagNumber, txs[0].TagNumber)
	mockRepo.AssertExpectations(t)
}
