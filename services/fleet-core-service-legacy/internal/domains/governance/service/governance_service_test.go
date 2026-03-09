package service

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/governance/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockRepo is a minimal mock for testing
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) CreateRequest(ctx context.Context, req *models.GovernanceRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockRepo) GetRequestByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.GovernanceRequest, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GovernanceRequest), args.Error(1)
}

func (m *MockRepo) GetRequestByIDInternal(ctx context.Context, id uuid.UUID) (*models.GovernanceRequest, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.GovernanceRequest), args.Error(1)
}

func (m *MockRepo) ListRequests(ctx context.Context, filter models.GovernanceRequestFilter) (*models.GovernanceRequestListResponse, error) {
	return nil, nil
}

func (m *MockRepo) UpdateRequestStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status models.GovernanceRequestStatus, comments *string, expiryDate *time.Time) error {
	args := m.Called(ctx, id, orgID, status, comments, expiryDate)
	return args.Error(0)
}

func (m *MockRepo) GetSettings(ctx context.Context, orgID uuid.UUID) (*models.GovernanceSettings, error) {
	return &models.GovernanceSettings{DownloadTTLHours: 48}, nil
}

func (m *MockRepo) UpsertSettings(ctx context.Context, settings *models.GovernanceSettings) error {
	return nil
}

func (m *MockRepo) CreateAuditLog(ctx context.Context, log *models.GovernanceAuditLog) error {
	return nil
}

func (m *MockRepo) ListAuditLogs(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]models.GovernanceAuditLog, int, error) {
	return nil, 0, nil
}

func (m *MockRepo) ExpireRequests(ctx context.Context) ([]models.ExpiredRequestInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.ExpiredRequestInfo), args.Error(1)
}

func (m *MockRepo) CreateExport(ctx context.Context, export *models.GovernanceExport) error { return nil }
func (m *MockRepo) GetExportByRequestID(ctx context.Context, requestID uuid.UUID) (*models.GovernanceExport, error) { return nil, nil }
func (m *MockRepo) CreateDownloadToken(ctx context.Context, token *models.DownloadToken) error { return nil }
func (m *MockRepo) GetExportByToken(ctx context.Context, token uuid.UUID) (*models.GovernanceExport, error) { return nil, nil }

func (m *MockRepo) CheckLegalHold(ctx context.Context, entityID uuid.UUID) (bool, error) {
	args := m.Called(ctx, entityID)
	return args.Bool(0), args.Error(1)
}
func (m *MockRepo) SaveExportLineage(ctx context.Context, lineage *models.ExportLineage) error {
	args := m.Called(ctx, lineage)
	return args.Error(0)
}
func (m *MockRepo) SaveComplianceReport(ctx context.Context, report *models.ComplianceReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}
func (m *MockRepo) UpdateVerificationStatus(ctx context.Context, id, orgID uuid.UUID, status string, verifiedBy uuid.UUID) error {
	args := m.Called(ctx, id, orgID, status, verifiedBy)
	return args.Error(0)
}
func (m *MockRepo) ListAllSettings(ctx context.Context) ([]models.GovernanceSettings, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.GovernanceSettings), args.Error(1)
}
func (m *MockRepo) EnforceRetentionPolicy(ctx context.Context, orgID uuid.UUID, tableName string, retentionDays int) (int64, error) {
	args := m.Called(ctx, orgID, tableName, retentionDays)
	return args.Get(0).(int64), args.Error(1)
}

func TestExpireRequests(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := NewGovernanceService(mockRepo, nil, zap.NewNop())
	
	expiredInfo := []models.ExpiredRequestInfo{
		{ID: uuid.New(), OrganizationID: uuid.New()},
	}
	
	mockRepo.On("ExpireRequests", mock.Anything).Return(expiredInfo, nil)
	
	count, err := svc.ExpireRequests(context.Background())
	
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	mockRepo.AssertExpectations(t)
}
