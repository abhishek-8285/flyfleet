package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/governance/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestExportService_ProcessExportRequest_Compliance(t *testing.T) {
	mockRepo := new(MockRepo)
	mockAgg := new(mockAggregator)
	mockMsk := new(mockMasker)
	mockGen := new(mockGenerator)
	logger := zap.NewNop()

	svc := NewExportService(mockRepo, mockAgg, mockMsk, mockGen, nil, logger)

	requestID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()

	req := &models.GovernanceRequest{
		ID:             requestID,
		OrganizationID: orgID,
		RequesterID:    userID,
		Status:         models.RequestStatusApproved,
		Format:         "JSON",
	}

	mockRepo.On("GetRequestByIDInternal", mock.Anything, requestID).Return(req, nil)
	mockRepo.On("UpdateRequestStatus", mock.Anything, requestID, orgID, models.RequestStatusProcessing, mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("UpdateRequestStatus", mock.Anything, requestID, orgID, models.RequestStatusCompleted, mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("GetSettings", mock.Anything, orgID).Return(&models.GovernanceSettings{MaskingRules: json.RawMessage("{}")}, nil)
	mockRepo.On("CreateExport", mock.Anything, mock.Anything).Return(nil)
	
	// Phase 5 & 6 expectations
	mockRepo.On("SaveExportLineage", mock.Anything, mock.MatchedBy(func(l *models.ExportLineage) bool {
		return l.RequestID == requestID && l.TableName == "aggregated_data"
	})).Return(nil)
	
	mockRepo.On("SaveComplianceReport", mock.Anything, mock.MatchedBy(func(r *models.ComplianceReport) bool {
		return r.RequestID == requestID && r.ExportChecksum == "sha256:test"
	})).Return(nil)

	dsarData := &models.DSARData{UserProfile: map[string]string{"name": "test"}}
	mockAgg.On("AggregateData", mock.Anything, userID, orgID).Return(dsarData, nil)
	mockMsk.On("MaskData", dsarData, mock.Anything).Return(dsarData)
	mockGen.On("GenerateJSON", dsarData, mock.Anything).Return(int64(1024), "sha256:test", nil)

	export, err := svc.ProcessExportRequest(context.Background(), requestID)

	assert.NoError(t, err)
	assert.NotNil(t, export)
	mockRepo.AssertExpectations(t)
	mockAgg.AssertExpectations(t)
	mockGen.AssertExpectations(t)
}
