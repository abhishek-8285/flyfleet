package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/models"
	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/repository"
	"github.com/google/uuid"
)

type TelemetryService struct {
	repo          *repository.DiagnosticsRepository
	healthService *DeviceHealthService
}

func NewTelemetryService(repo *repository.DiagnosticsRepository, hs *DeviceHealthService) *TelemetryService {
	return &TelemetryService{repo: repo, healthService: hs}
}

func (s *TelemetryService) IngestTelemetry(ctx context.Context, deviceID uuid.UUID, tenantID uuid.UUID, tel *models.DeviceTelemetry) error {
	device, err := s.repo.GetDevice(ctx, deviceID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to load device: %w", err)
	}

	tel.DeviceID = &deviceID
	now := time.Now()
	if tel.Timestamp == nil {
		tel.Timestamp = &now
	}

	if err := s.repo.InsertTelemetry(ctx, tel); err != nil {
		return fmt.Errorf("failed to store telemetry payload: %w", err)
	}

	device.LastSeenAt = tel.Timestamp
	if err := s.repo.UpdateDevice(ctx, device); err != nil {
		return fmt.Errorf("failed to update device last_seen: %w", err)
	}

	health := s.healthService.AssessHealth(device, tel)
	_ = health // in a real app, this might publish an event if health changed or degraded

	return nil
}
