package service

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestDutyStatusService_CheckIndiaDrivingEligibility(t *testing.T) {
	mockRepo := new(MockDriverRepository)
	logger := zap.NewNop()
	service := NewDutyStatusService(mockRepo, logger)

	ctx := context.Background()
	driverID := "driver-123"

	t.Run("Compliant - normal driving", func(t *testing.T) {
		now := time.Now()
		logs := []*models.DutyStatusLog{
			{
				Status:    StatusDriving,
				StartedAt: now.Add(-2 * time.Hour),
			},
		}
		mockRepo.On("GetDutyStatusLogs", ctx, driverID, mock.Anything, mock.Anything).Return(logs, nil).Once()

		err := service.checkIndiaDrivingEligibility(ctx, driverID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Violation - 8h continuous driving", func(t *testing.T) {
		now := time.Now()
		logs := []*models.DutyStatusLog{
			{
				Status:    StatusDriving,
				StartedAt: now.Add(-9 * time.Hour),
			},
		}
		mockRepo.On("GetDutyStatusLogs", ctx, driverID, mock.Anything, mock.Anything).Return(logs, nil).Once()

		err := service.checkIndiaDrivingEligibility(ctx, driverID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mandatory rest required after 8h driving")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Violation - 12h duty window", func(t *testing.T) {
		now := time.Now()
		logs := []*models.DutyStatusLog{
			{
				Status:    StatusOnDuty,
				StartedAt: now.Add(-13 * time.Hour),
			},
		}
		mockRepo.On("GetDutyStatusLogs", ctx, driverID, mock.Anything, mock.Anything).Return(logs, nil).Once()

		err := service.checkIndiaDrivingEligibility(ctx, driverID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duty window exceeded 12h")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Violation - No break after 5h", func(t *testing.T) {
		now := time.Now()
		logs := []*models.DutyStatusLog{
			{
				Status:    StatusDriving,
				StartedAt: now.Add(-6 * time.Hour),
			},
		}
		mockRepo.On("GetDutyStatusLogs", ctx, driverID, mock.Anything, mock.Anything).Return(logs, nil).Once()

		err := service.checkIndiaDrivingEligibility(ctx, driverID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "30-minute break required after 5h driving")
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("Compliant - With break after 5h", func(t *testing.T) {
		now := time.Now()
		drivingStart := now.Add(-6 * time.Hour)
		duration := 30
		logs := []*models.DutyStatusLog{
			{
				Status:    StatusDriving,
				StartedAt: drivingStart,
			},
			{
				Status:          StatusBreak,
				StartedAt:       now.Add(-30 * time.Minute),
				DurationMinutes: &duration,
			},
		}
		mockRepo.On("GetDutyStatusLogs", ctx, driverID, mock.Anything, mock.Anything).Return(logs, nil).Once()

		err := service.checkIndiaDrivingEligibility(ctx, driverID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
