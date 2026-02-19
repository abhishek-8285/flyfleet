package service

import (
	"context"
	"fmt"

	"github.com/flyfleet/trip-service/internal/models"
	"github.com/flyfleet/trip-service/internal/repository"
	"go.uber.org/zap"
)

type SettingsService struct {
	repo   *repository.SettingsRepository
	logger *zap.Logger
}

func NewSettingsService(repo *repository.SettingsRepository, logger *zap.Logger) *SettingsService {
	return &SettingsService{
		repo:   repo,
		logger: logger,
	}
}

func (s *SettingsService) GetSettings(ctx context.Context, orgID string) (*models.SettingsResponse, error) {
	settings, meta, err := s.repo.GetSettings(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to get settings from repo", zap.String("org_id", orgID), zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve settings: %w", err)
	}

	defaults := s.GetDefaultSettings()

	if settings == nil {
		return &models.SettingsResponse{
			Data: defaults,
			Meta: models.SettingsMeta{
				Version: "1.0",
			},
		}, nil
	}

	// In a more complex scenario, we'd deep merge. 
	// For now, if we have settings, we use them.
	// We can ensure defaults for missing fields if needed.
	
	return &models.SettingsResponse{
		Data: *settings,
		Meta: *meta,
	}, nil
}

func (s *SettingsService) UpdateSettings(ctx context.Context, orgID string, req models.UpdateSettingsRequest, userID string) error {
	// Validation
	if req.Settings.HardBrakingThreshold < 0 || req.Settings.HardBrakingThreshold > 1 {
		return fmt.Errorf("invalid hard_braking_threshold: must be between 0 and 1")
	}
	if req.Settings.RapidAccelerationThreshold < 0 || req.Settings.RapidAccelerationThreshold > 1 {
		return fmt.Errorf("invalid rapid_acceleration_threshold: must be between 0 and 1")
	}

	version := "1.1" // Simple version increment or constant for now
	err := s.repo.UpdateSettings(ctx, orgID, req.Settings, userID, version)
	if err != nil {
		s.logger.Error("failed to update settings", zap.String("org_id", orgID), zap.Error(err))
		return err
	}

	return nil
}

func (s *SettingsService) GetDefaultSettings() models.TripSettings {
	return models.TripSettings{
		AutoStartEnabled:           true,
		IdleThresholdMinutes:       5,
		SpeedLimitKph:              ptrInt(80),
		GeofenceAlertsEnabled:      true,
		MaintenanceReminderDays:    30,
		DriverBreakReminderMinutes: 240,
		FuelEfficiencyAlerts:       true,
		HardBrakingThreshold:       0.3,
		RapidAccelerationThreshold: 0.3,
		NightDrivingRestricted:     false,
		NightDrivingStart:          "22:00",
		NightDrivingEnd:            "06:00",
	}
}

func ptrInt(i int) *int {
	return &i
}
