package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/governance/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockAggregator struct {
	mock.Mock
}

func (m *mockAggregator) AggregateData(ctx context.Context, userID, tenantID uuid.UUID) (*models.DSARData, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DSARData), args.Error(1)
}

type mockMasker struct {
	mock.Mock
}

func (m *mockMasker) MaskData(data *models.DSARData, rules json.RawMessage) *models.DSARData {
	args := m.Called(data, rules)
	return args.Get(0).(*models.DSARData)
}

type mockGenerator struct {
	mock.Mock
}

func (m *mockGenerator) GenerateJSON(data *models.DSARData, targetPath string) (int64, string, error) {
	args := m.Called(data, targetPath)
	return args.Get(0).(int64), args.String(1), args.Error(2)
}

func (m *mockGenerator) GenerateCSV(data *models.DSARData, targetPath string) (int64, string, error) {
	return 0, "", nil
}

func (m *mockGenerator) GeneratePDF(data *models.DSARData, targetPath string) (int64, string, error) {
	return 0, "", nil
}

func TestExportService_ProcessExportRequest_Compliance(t *testing.T) {
	mockRepo := new(MockRepo)
	mockAgg := new(mockAggregator)
	mockMsk := new(mockMasker)
	mockGen := new(mockGenerator)
	logger := zap.NewNop()

	svc := NewExportService(mockRepo, mockAgg, mockMsk, mockGen, nil, logger)

	requestID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()

	req := &models.GovernanceRequest{
		ID:             requestID,
		TenantID: tenantID,
		RequesterID:    userID,
		Status:         models.RequestStatusApproved,
		Format:         "JSON",
	}

	mockRepo.On("GetRequestByIDInternal", mock.Anything, requestID).Return(req, nil)
	mockRepo.On("UpdateRequestStatus", mock.Anything, requestID, tenantID, models.RequestStatusProcessing, mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("UpdateRequestStatus", mock.Anything, requestID, tenantID, models.RequestStatusCompleted, mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("GetSettings", mock.Anything, tenantID).Return(&models.GovernanceSettings{MaskingRules: json.RawMessage("{}")}, nil)
	mockRepo.On("CreateExport", mock.Anything, mock.Anything).Return(nil)

	// Phase 5 & 6 expectations
	mockRepo.On("SaveExportLineage", mock.Anything, mock.MatchedBy(func(l *models.ExportLineage) bool {
		return l.RequestID == requestID && l.TableName == "aggregated_data"
	})).Return(nil)

	mockRepo.On("SaveComplianceReport", mock.Anything, mock.MatchedBy(func(r *models.ComplianceReport) bool {
		return r.RequestID == requestID && r.ExportChecksum == "sha256:test"
	})).Return(nil)

	dsarData := &models.DSARData{UserProfile: map[string]string{"name": "test"}}
	mockAgg.On("AggregateData", mock.Anything, userID, tenantID).Return(dsarData, nil)
	mockMsk.On("MaskData", dsarData, mock.Anything).Return(dsarData)
	mockGen.On("GenerateJSON", dsarData, mock.Anything).Return(int64(1024), "sha256:test", nil)

	export, err := svc.ProcessExportRequest(context.Background(), requestID)

	assert.NoError(t, err)
	assert.NotNil(t, export)
	mockRepo.AssertExpectations(t)
	mockAgg.AssertExpectations(t)
	mockGen.AssertExpectations(t)
}
