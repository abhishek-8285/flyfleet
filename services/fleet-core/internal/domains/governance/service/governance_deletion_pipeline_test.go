package service

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/governance/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestDeletionService_ProcessDeletionRequest_LegalHold(t *testing.T) {
	mockRepo := new(MockRepo) // Redefined if needed, or use existing if in same package
	logger := zap.NewNop()

	// Note: We need a db pool mock or just use the repo if we mock the whole flow.
	// Since the service uses s.db.QueryRow for delCount and s.db.Begin for Tx,
	// we might need a more complex mock or satisfy it.

	svc := &dataDeletionService{
		repo:   mockRepo,
		logger: logger,
		// db: ... we won't reach DB if legal hold is hit early
	}

	requestID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()

	req := &models.GovernanceRequest{
		ID:             requestID,
		OrganizationID: orgID,
		RequesterID:    userID,
		RequestType:    models.RequestTypeDeletion,
		Status:         models.RequestStatusApproved,
	}

	mockRepo.On("GetRequestByIDInternal", mock.Anything, requestID).Return(req, nil)
	mockRepo.On("CheckLegalHold", mock.Anything, userID).Return(true, nil)
	mockRepo.On("UpdateRequestStatus", mock.Anything, requestID, orgID, models.RequestStatusFailed, mock.Anything, mock.Anything).Return(nil)

	err := svc.ProcessDeletionRequest(context.Background(), requestID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "legal hold active")
	mockRepo.AssertExpectations(t)
}
