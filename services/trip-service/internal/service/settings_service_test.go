package service

import (
	"context"
	"testing"

	"github.com/flyfleet/trip-service/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestGetDefaultSettings(t *testing.T) {
	s := &SettingsService{}
	defaults := s.GetDefaultSettings()

	assert.True(t, defaults.AutoStartEnabled)
	assert.Equal(t, 5, defaults.IdleThresholdMinutes)
	assert.Equal(t, 80, *defaults.SpeedLimitKph)
	assert.Equal(t, "22:00", defaults.NightDrivingStart)
}

func TestSettingsValidation(t *testing.T) {
	s := &SettingsService{}
	orgID := "test-org"
	userID := "test-user"

	// Valid settings
	req := models.UpdateSettingsRequest{
		Settings: models.TripSettings{
			HardBrakingThreshold:       0.5,
			RapidAccelerationThreshold: 0.5,
		},
	}
	// We'd need to mock the repo to test the full UpdateSettings call
	// For now, let's just assume validation passes if no error is returned before repo call
	
	// Invalid HardBrakingThreshold
	req.Settings.HardBrakingThreshold = 1.5
	err := s.UpdateSettings(context.Background(), orgID, req, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hard_braking_threshold")

	// Invalid RapidAccelerationThreshold
	req.Settings.HardBrakingThreshold = 0.5
	req.Settings.RapidAccelerationThreshold = -0.1
	err = s.UpdateSettings(context.Background(), orgID, req, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rapid_acceleration_threshold")
}
