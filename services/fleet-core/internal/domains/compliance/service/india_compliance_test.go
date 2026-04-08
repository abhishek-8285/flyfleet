//go:build integration
// +build integration

package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core/internal/domains/compliance/models"
	"github.com/flyfleet/fleet-core/internal/domains/compliance/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

// Minimal mock repo definition required to test service layer
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) RecordDutyStatus(ctx context.Context, log *models.DriverDutyLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockRepository) GetLatestDutyStatus(ctx context.Context, driverID uuid.UUID) (*models.DriverDutyLog, error) {
	args := m.Called(ctx, driverID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.DriverDutyLog), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) GetDutyLogsByDriver(ctx context.Context, driverID uuid.UUID, start, end time.Time) ([]models.DriverDutyLog, error) {
	args := m.Called(ctx, driverID, start, end)
	return args.Get(0).([]models.DriverDutyLog), args.Error(1)
}

func (m *MockRepository) RecordBreakLog(ctx context.Context, log *models.DriverBreakLog) error {
	return nil
}
func (m *MockRepository) GetBreakLogsByDriver(ctx context.Context, driverID uuid.UUID, start, end time.Time) ([]models.DriverBreakLog, error) {
	return nil, nil
}
func (m *MockRepository) RecordViolation(ctx context.Context, violation *models.DriverComplianceViolation) error {
	return nil
}
func (m *MockRepository) GetDriverViolations(ctx context.Context, driverID uuid.UUID) ([]models.DriverComplianceViolation, error) {
	return nil, nil
}
func (m *MockRepository) CreateComplianceAlert(ctx context.Context, alert *models.ComplianceAlert) error {
	return m.Called(ctx, alert).Error(0)
}
func (m *MockRepository) RegisterDevice(ctx context.Context, device *models.AIS140Device) error {
	return nil
}
func (m *MockRepository) RecordDeviceHealth(ctx context.Context, log *models.DeviceHealthLog) error {
	return nil
}
func (m *MockRepository) GetDeviceHealthLogs(ctx context.Context, deviceID uuid.UUID, limit int) ([]models.DeviceHealthLog, error) {
	return nil, nil
}

func (m *MockRepository) CreatePanicAlert(ctx context.Context, alert *models.PanicAlert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockRepository) RecordPanicAction(ctx context.Context, action *models.PanicAlertAction) error {
	return nil
}
func (m *MockRepository) GetActivePanicAlerts(ctx context.Context) ([]models.PanicAlert, error) {
	return nil, nil
}
func (m *MockRepository) CreateVehicleDocument(ctx context.Context, doc *models.VehicleDocument) error {
	return nil
}
func (m *MockRepository) GetVehicleDocuments(ctx context.Context, vehicleID uuid.UUID) ([]models.VehicleDocument, error) {
	return nil, nil
}
func (m *MockRepository) CreateDriverDocument(ctx context.Context, doc *models.DriverDocument) error {
	return nil
}
func (m *MockRepository) GetDriverDocuments(ctx context.Context, driverID uuid.UUID) ([]models.DriverDocument, error) {
	return nil, nil
}

func (m *MockRepository) GetExpiringDocuments(ctx context.Context, days int) ([]models.DocumentExpiryAlert, error) {
	args := m.Called(ctx, days)
	return args.Get(0).([]models.DocumentExpiryAlert), args.Error(1)
}

func (m *MockRepository) CreateTripSheet(ctx context.Context, trip *models.TripSheet) error {
	return nil
}
func (m *MockRepository) UpdateTripSheet(ctx context.Context, trip *models.TripSheet) error {
	return nil
}
func (m *MockRepository) GetTripSheet(ctx context.Context, id uuid.UUID) (*models.TripSheet, error) {
	return nil, nil
}
func (m *MockRepository) GetTripSheetsByDriver(ctx context.Context, driverID uuid.UUID) ([]models.TripSheet, error) {
	return nil, nil
}
func (m *MockRepository) RecordTripEvent(ctx context.Context, event *models.TripEvent) error {
	return nil
}
func (m *MockRepository) RecordTollTransaction(ctx context.Context, txn *models.TollTransaction) error {
	return nil
}
func (m *MockRepository) RecordFASTagReconciliation(ctx context.Context, recon *models.FASTagReconciliation) error {
	return nil
}
func (m *MockRepository) RecordBorderCrossing(ctx context.Context, crossing *models.StateBorderCrossing) error {
	return nil
}
func (m *MockRepository) RecordAuditLog(ctx context.Context, log *models.ComplianceAuditLog) error {
	return nil
}
func (m *MockRepository) CleanupOldData(ctx context.Context, retentionLimit time.Time) (int64, error) {
	return 0, nil
}

func TestDriverDutyService_RecordDutyStatus(t *testing.T) {
	dummyLogger, _ := logger.New(logger.Config{Level: zapcore.InfoLevel})
	mockRepo := new(MockRepository)
	svc := service.NewDriverDutyService(dummyLogger, mockRepo)

	ctx := context.Background()
	driverID := uuid.New()
	vehicleID := uuid.New()

	// Mock getting current status (no previous rows)
	mockRepo.On("GetLatestDutyStatus", ctx, driverID).Return(nil, nil)

	// Mock recording the new status
	mockRepo.On("RecordDutyStatus", ctx, mock.AnythingOfType("*models.DriverDutyLog")).Return(nil)

	log, err := svc.RecordDutyStatus(ctx, driverID, vehicleID, models.DutyStatusDriving, "Mumbai Toll Plaza", "Starting shift")

	assert.NoError(t, err)
	assert.NotNil(t, log)
	assert.Equal(t, models.DutyStatusDriving, log.Status)
	assert.Equal(t, "Mumbai Toll Plaza", log.Location)

	mockRepo.AssertExpectations(t)
}

func TestAIS140Service_PanicAlert(t *testing.T) {
	dummyLogger, _ := logger.New(logger.Config{Level: zapcore.InfoLevel})
	mockRepo := new(MockRepository)
	svc := service.NewAIS140Service(mockRepo, dummyLogger)

	ctx := context.Background()
	vehicleID := uuid.New()

	mockRepo.On("CreatePanicAlert", ctx, mock.AnythingOfType("*models.PanicAlert")).Return(nil)
	mockRepo.On("CreateComplianceAlert", ctx, mock.AnythingOfType("*models.ComplianceAlert")).Return(nil)

	err := svc.HandlePanicAlert(ctx, vehicleID, nil, 19.0760, 72.8777) // Mumbai coords
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDocumentService_ExpiryCheck(t *testing.T) {
	dummyLogger, _ := logger.New(logger.Config{Level: zapcore.InfoLevel})
	mockRepo := new(MockRepository)
	svc := service.NewDocumentService(mockRepo, dummyLogger)

	ctx := context.Background()

	// Mock an expiring document (10 days left)
	mockRepo.On("GetExpiringDocuments", ctx, 30).Return([]models.DocumentExpiryAlert{
		{
			DocumentID:   uuid.New(),
			VehicleID:    uuid.New(),
			DocumentType: models.DocRC,
			DaysToExpiry: 10,
			IsExpired:    false,
		},
	}, nil)

	alerts, err := svc.RunExpiryCheck(ctx)
	assert.NoError(t, err)
	assert.Len(t, alerts, 1)
	assert.Equal(t, 10, alerts[0].DaysToExpiry)
	mockRepo.AssertExpectations(t)
}
