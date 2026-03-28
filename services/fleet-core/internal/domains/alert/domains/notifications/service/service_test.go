package service

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/notifications/dto"
	"github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/notifications/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/notifications/service/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestNotificationService_ResendNotification(t *testing.T) {
	mockRepo := new(mocks.MockNotificationRepository)
	logger := zap.NewNop()
	svc := NewNotificationService(mockRepo, logger)

	originalID := uuid.New()
	original := &models.NotificationLog{
		ID:         originalID,
		UserEmail:  "test@example.com",
		Channel:    models.ChannelEmail,
		Subject:    "Original Subject",
		Body:       "Original Body",
		Provider:   strPtr("SES"),
		MaxRetries: 3,
	}

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		// Expect GetLog to fetch original
		mockRepo.On("GetLog", ctx, originalID).Return(original, nil)

		// Expect CreateLog to be called with new log (based on original)
		mockRepo.On("CreateLog", ctx, mock.MatchedBy(func(l *models.NotificationLog) bool {
			return l.UserEmail == original.UserEmail &&
				l.Status == models.StatusQueued &&
				l.Metadata["resend_reason"] == "Needed manually"
		})).Run(func(args mock.Arguments) {
			l := args.Get(1).(*models.NotificationLog)
			l.ID = uuid.New()
		}).Return(nil)

		// Expect AddEvent to record the resend
		mockRepo.On("AddEvent", ctx, mock.Anything).Return(nil)

		newID, err := svc.ResendNotification(ctx, originalID, "Needed manually", true)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, newID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		ctx := context.Background()
		notFoundID := uuid.New()

		mockRepo.On("GetLog", ctx, notFoundID).Return(nil, nil)

		_, err := svc.ResendNotification(ctx, notFoundID, "Reason", false)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestNotificationService_UpdateSettings(t *testing.T) {
	mockRepo := new(mocks.MockNotificationRepository)
	logger := zap.NewNop()
	svc := NewNotificationService(mockRepo, logger)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		updates := []dto.SettingUpdate{
			{ID: uuid.New(), Channel: models.ChannelEmail, Enabled: true},
			{ID: uuid.New(), Channel: models.ChannelSMS, Enabled: false},
		}

		mockRepo.On("UpdateSettings", ctx, mock.Anything).Return(nil).Twice()

		err := svc.UpdateSettings(ctx, updates)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func strPtr(s string) *string {
	return &s
}
