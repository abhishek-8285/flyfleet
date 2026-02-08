package integration_tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/flyfleet/notification-service/internal/models"
	"github.com/flyfleet/notification-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreferenceRepository_Integration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	repo := repository.NewNotificationRepository(pool)

	// Seed a test user
	testUserID := uuid.New()
	_, err = pool.Exec(ctx, "INSERT INTO users (id, email, name, password_hash, status) VALUES ($1, $2, $3, $4, $5)", 
		testUserID, "test_pref@example.com", "Test Pref User", "hash", "active")
	require.NoError(t, err)
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM notification_preference_audit WHERE user_id = $1", testUserID)
		_, _ = pool.Exec(ctx, "DELETE FROM user_notification_preferences WHERE user_id = $1", testUserID)
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", testUserID)
	}()

	t.Run("GetUserPreference_DefaultFallback", func(t *testing.T) {
		pref, err := repo.GetUserPreference(ctx, testUserID)
		require.NoError(t, err)
		assert.NotNil(t, pref)
		assert.Equal(t, testUserID, pref.UserID)
		assert.True(t, pref.EmailEnabled) // Default from fallback
		assert.False(t, pref.SmsEnabled)
	})

	t.Run("UpdateAndGetPreference", func(t *testing.T) {
		p := &models.UserNotificationPreference{
			UserID:       testUserID,
			EmailEnabled: false,
			SmsEnabled:   true,
			PushEnabled:  true,
			InAppEnabled: true,
			LastModifiedBy: uuid.New(),
		}

		err := repo.UpdateUserPreference(ctx, p)
		require.NoError(t, err)

		fetched, err := repo.GetUserPreference(ctx, testUserID)
		require.NoError(t, err)
		assert.False(t, fetched.EmailEnabled)
		assert.True(t, fetched.SmsEnabled)
		assert.NotEqual(t, time.Time{}, fetched.LastModifiedAt)
	})

	t.Run("PreferenceAuditTrail", func(t *testing.T) {
		adminID := uuid.New()
		audit := &models.PreferenceAudit{
			UserID:       testUserID,
			ChangedBy:    adminID,
			OldValues:    map[string]any{"email": true},
			NewValues:    map[string]any{"email": false},
			ChangeReason: "Policy update",
		}

		err := repo.AddPreferenceAudit(ctx, audit)
		require.NoError(t, err)

		audits, err := repo.GetPreferenceAudit(ctx, testUserID)
		require.NoError(t, err)
		require.NotEmpty(t, audits)
		assert.Equal(t, "Policy update", audits[0].ChangeReason)
		assert.Equal(t, adminID, audits[0].ChangedBy)
	})

	t.Run("GlobalDefaultsPersistence", func(t *testing.T) {
		// Ensure at least one default exists
		_, _ = pool.Exec(ctx, "INSERT INTO global_notification_defaults (email_enabled, sms_enabled, push_enabled, in_app_enabled, updated_by) VALUES (true, false, true, true, $1) ON CONFLICT DO NOTHING", uuid.New())

		d, err := repo.GetGlobalDefaults(ctx)
		require.NoError(t, err)
		require.NotNil(t, d)

		d.SmsEnabled = true
		d.UpdatedBy = uuid.New()
		err = repo.UpdateGlobalDefaults(ctx, d)
		require.NoError(t, err)

		updated, err := repo.GetGlobalDefaults(ctx)
		require.NoError(t, err)
		assert.True(t, updated.SmsEnabled)
	})

	t.Run("ListPreferencesWithSearch", func(t *testing.T) {
		prefs, total, err := repo.ListPreferences(ctx, 1, 10, "test_pref@example.com")
		require.NoError(t, err)
		assert.True(t, total >= 1)
		
		var found bool
		for _, p := range prefs {
			if p.UserID == testUserID {
				found = true
				assert.Equal(t, "Test Pref User", p.UserName)
			}
		}
		assert.True(t, found, "Test user not found in list")
	})
}
