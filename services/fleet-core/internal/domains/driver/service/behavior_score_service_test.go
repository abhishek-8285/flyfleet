package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestBehaviorScoreService_ProcessBehaviorEvent(t *testing.T) {
	mockRepo := new(MockDriverRepository)
	logger := zap.NewNop()
	service := NewBehaviorScoreService(mockRepo, logger)

	ctx := context.Background()
	driverID := uuid.New().String()
	tenantID := uuid.New()

	driver := &models.Driver{
		ID:            driverID,
		TenantID:         tenantID.String(),
		BehaviorScore: 85,
	}

	updatedDriver := *driver
	updatedDriver.BehaviorScore = 85 - models.PenaltySpeeding

	t.Run("Process Speeding Event", func(t *testing.T) {
		event := &models.BehaviorEvent{
			DriverID:  driverID,
			EventType: "speeding",
		}

		// 1. ProcessBehaviorEvent calls GetByID
		mockRepo.On("GetByID", ctx, driverID).Return(driver, nil).Once()
		// 2. UpdateBehaviorScore calls GetByID again
		mockRepo.On("GetByID", ctx, driverID).Return(driver, nil).Once()

		mockRepo.On("UpdateBehaviorScore", ctx, driverID, -models.PenaltySpeeding).Return(nil).Once()
		mockRepo.On("LogBehaviorScoreChange", ctx,
			uuid.MustParse(driverID),
			tenantID,
			85,
			85-models.PenaltySpeeding,
			fmt.Sprintf("Automated penalty for %s", event.EventType),
			"speeding",
			uuid.Nil,
			"system").Return(nil).Once()
		mockRepo.On("CreateBehaviorScoreUpdate", ctx, mock.Anything).Return(nil).Once()

		// 3. UpdateBehaviorScore calls GetByID at the end
		mockRepo.On("GetByID", ctx, driverID).Return(&updatedDriver, nil).Once()

		err := service.ProcessBehaviorEvent(ctx, event)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Process Harsh Brake Event", func(t *testing.T) {
		event := &models.BehaviorEvent{
			DriverID:  driverID,
			EventType: "harsh_braking",
		}

		harshUpdatedDriver := *driver
		harshUpdatedDriver.BehaviorScore = 85 - models.PenaltyHarshBrake

		// 1. ProcessBehaviorEvent calls GetByID
		mockRepo.On("GetByID", ctx, driverID).Return(driver, nil).Once()
		// 2. UpdateBehaviorScore calls GetByID
		mockRepo.On("GetByID", ctx, driverID).Return(driver, nil).Once()

		mockRepo.On("UpdateBehaviorScore", ctx, driverID, -models.PenaltyHarshBrake).Return(nil).Once()
		mockRepo.On("LogBehaviorScoreChange", ctx,
			uuid.MustParse(driverID),
			tenantID,
			85,
			85-models.PenaltyHarshBrake,
			fmt.Sprintf("Automated penalty for %s", event.EventType),
			"harsh_braking",
			uuid.Nil,
			"system").Return(nil).Once()
		mockRepo.On("CreateBehaviorScoreUpdate", ctx, mock.Anything).Return(nil).Once()

		// 3. UpdateBehaviorScore calls GetByID at the end
		mockRepo.On("GetByID", ctx, driverID).Return(&harshUpdatedDriver, nil).Once()

		err := service.ProcessBehaviorEvent(ctx, event)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
