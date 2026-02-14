package integration_tests

import (
	"context"
	"os"
	"testing"

	"github.com/flyfleet/notification-service/internal/models"
	"github.com/flyfleet/notification-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TestUserEmail = "user@example.com"

func TestNotificationRepository_Integration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = ""
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	repo := repository.NewNotificationRepository(pool)

	t.Run("CreateAndGetLog", func(t *testing.T) {
		log := &models.NotificationLog{
			CorrelationID: strPtr("test-corr-1"),
			UserEmail:    "user@example.com",
			Channel:      models.ChannelEmail,
			Type:         "SECURITY",
			Subject:      "Integrated Test",
			Body:         "Hello world",
			Status:       models.StatusQueued,
			MaxRetries:   3,
			Metadata:     map[string]any{"test": true},
		}

		err := repo.CreateLog(ctx, log)
		require.NoError(t, err, "Failed to create log")
		assert.NotEqual(t, uuid.Nil, log.ID)

		// Get it back
		fetched, err := repo.GetLog(ctx, log.ID)
		require.NoError(t, err, "Failed to get log")
		require.NotNil(t, fetched, "Fetched log is nil")
		assert.Equal(t, log.Subject, fetched.Subject)
		assert.Equal(t, log.Status, fetched.Status)
		assert.Equal(t, "SECURITY", fetched.Type)
	})

	t.Run("AddEventAndFetchTimeline", func(t *testing.T) {
		log := &models.NotificationLog{
			UserEmail: "timeline@example.com",
			Channel:   models.ChannelSMS,
			Type:      "SYSTEM",
			Status:    models.StatusSent,
			Subject:   "Timeline Test",
			Body:      "Body",
		}
		err := repo.CreateLog(ctx, log)
		require.NoError(t, err, "Failed to create log for timeline test")

		event := &models.NotificationEvent{
			NotificationID: log.ID,
			EventType:      "SENT",
			Message:        "Delivered to gateway",
		}
		err = repo.AddEvent(ctx, event)
		require.NoError(t, err, "Failed to add event")

		fetched, err := repo.GetLog(ctx, log.ID)
		require.NoError(t, err, "Failed to get log for timeline test")
		require.NotNil(t, fetched, "Fetched log is nil for timeline test")
		assert.Len(t, fetched.Items, 1)
		assert.Equal(t, "SENT", fetched.Items[0].EventType)
	})

	t.Run("SettingsPersistence", func(t *testing.T) {
		s := &models.NotificationSettings{
			Channel:         models.ChannelPush,
			Enabled:         true,
			MaxRetries:      5,
			Priority:        "HIGH",
			RetentionMonths: 12,
		}

		err := repo.UpdateSettings(ctx, s)
		require.NoError(t, err, "Failed to update settings")

		settings, err := repo.GetSettings(ctx)
		require.NoError(t, err, "Failed to get settings")
		
		var found bool
		for _, setting := range settings {
			if setting.Channel == models.ChannelPush {
				found = true
				assert.Equal(t, 5, setting.MaxRetries)
				assert.Equal(t, "HIGH", setting.Priority)
			}
		}
		assert.True(t, found, "Push channel settings not found")
	})

	t.Run("ListFilters", func(t *testing.T) {
		// Cleanup/Seed
		_, _ = pool.Exec(ctx, "DELETE FROM notification_logs WHERE user_email = 'filter@example.com'")
		
		err1 := repo.CreateLog(ctx, &models.NotificationLog{UserEmail: "filter@example.com", Channel: models.ChannelEmail, Status: models.StatusSent, Subject: "S1", Type: "MAINTENANCE"})
		require.NoError(t, err1)
		err2 := repo.CreateLog(ctx, &models.NotificationLog{UserEmail: "filter@example.com", Channel: models.ChannelSMS, Status: models.StatusFailed, Subject: "S2", Type: "MAINTENANCE"})
		require.NoError(t, err2)

		logs, total, err := repo.ListLogs(ctx, 1, 10, "SENT", "", "filter@example.com")
		require.NoError(t, err, "Failed to list logs")
		assert.Equal(t, 1, len(logs))
		assert.True(t, total >= 1)
		assert.Equal(t, "SENT", string(logs[0].Status))
	})
}

func strPtr(s string) *string {
	return &s
}
