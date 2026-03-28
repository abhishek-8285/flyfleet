package service

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AIS140Service interface {
	HandlePanicAlert(ctx context.Context, vehicleID uuid.UUID, driverID *uuid.UUID, lat, lon float64) error
	HandleTamperAlert(ctx context.Context, deviceID, vehicleID uuid.UUID) error
}

type ais140Service struct {
	repo   repository.ComplianceRepository
	logger logger.Logger
}

func NewAIS140Service(repo repository.ComplianceRepository, logger logger.Logger) AIS140Service {
	return &ais140Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *ais140Service) HandlePanicAlert(ctx context.Context, vehicleID uuid.UUID, driverID *uuid.UUID, lat, lon float64) error {
	alert := &models.PanicAlert{
		ID:        uuid.New(),
		VehicleID: vehicleID,
		DriverID:  driverID,
		Status:    models.PanicAlertActive,
		Timestamp: time.Now(),
		Latitude:  lat,
		Longitude: lon,
	}

	s.logger.Warn("AIS-140 PANIC ALERT RECEIVED",
		zap.String("vehicle_id", vehicleID.String()),
		zap.Float64("lat", lat),
		zap.Float64("lon", lon),
	)

	if err := s.repo.CreatePanicAlert(ctx, alert); err != nil {
		return err
	}

	// Audit Log
	audit := &models.ComplianceAuditLog{
		ID:         uuid.New(),
		Action:     "panic_alert_created",
		EntityType: "VEHICLE",
		EntityID:   vehicleID,
		Timestamp:  time.Now(),
	}
	_ = s.repo.RecordAuditLog(ctx, audit)

	// Elevate to unified compliance alert
	compAlert := &models.ComplianceAlert{
		ID:             uuid.New(),
		OrganizationID: uuid.Nil, // Should be fetched from vehicle context in real app
		Type:           models.AlertPanic,
		EntityID:       alert.ID,
		Severity:       models.AlertSeverityCritical,
	}

	return s.repo.CreateComplianceAlert(ctx, compAlert)
}

func (s *ais140Service) HandleTamperAlert(ctx context.Context, deviceID, vehicleID uuid.UUID) error {
	log := &models.DeviceHealthLog{
		ID:        uuid.New(),
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Status:    string(models.DeviceStatusTampered),
	}

	s.logger.Warn("AIS-140 DEVICE TAMPER ALERT RECEIVED",
		zap.String("device_id", deviceID.String()),
	)

	if err := s.repo.RecordDeviceHealth(ctx, log); err != nil {
		return err
	}

	compAlert := &models.ComplianceAlert{
		ID:             uuid.New(),
		OrganizationID: uuid.Nil,
		Type:           models.AlertDeviceTamper,
		EntityID:       deviceID,
		Severity:       models.AlertSeverityHigh,
	}

	return s.repo.CreateComplianceAlert(ctx, compAlert)
}
