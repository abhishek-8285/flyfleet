package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/tracking/models"
	"github.com/flyfleet/fleet-core/internal/domains/tracking/repository"
	"github.com/flyfleet/fleet-core/internal/messaging"
	"go.uber.org/zap"
)

// AlertService handles real-time alert detection and processing
type AlertService struct {
	repo       repository.AlertRepository
	eventBus   messaging.EventBus
	mqttClient *messaging.MQTTClient
	logger     *zap.Logger
}

// NewAlertService creates a new alert service
func NewAlertService(
	repo repository.AlertRepository,
	eventBus messaging.EventBus,
	mqttClient *messaging.MQTTClient,
	logger *zap.Logger,
) *AlertService {
	return &AlertService{
		repo:       repo,
		eventBus:   eventBus,
		mqttClient: mqttClient,
		logger:     logger,
	}
}

// ProcessTelemetryAlert processes an alert from telemetry ingestion
func (s *AlertService) ProcessTelemetryAlert(ctx context.Context, alert *models.VehicleAlert) error {
	s.logger.Info("processing alert",
		zap.String("type", alert.Type),
		zap.String("severity", alert.Severity),
		zap.String("vehicle_id", alert.VehicleID))

	// Store alert in database
	if err := s.repo.SaveAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to save alert: %w", err)
	}

	// Publish to MQTT for real-time notifications
	if s.mqttClient != nil {
		topic := fmt.Sprintf("fleet/%s/alerts/%s", alert.OrgID, alert.VehicleID)
		if err := s.mqttClient.PublishJSON(topic, alert); err != nil {
			s.logger.Warn("failed to publish alert to MQTT", zap.Error(err))
		}
	}

	return nil
}

// GetVehicleAlerts retrieves alerts for a vehicle
func (s *AlertService) GetVehicleAlerts(ctx context.Context, vehicleID string, from, to time.Time, limit int) ([]*models.VehicleAlert, error) {
	return s.repo.GetAlerts(ctx, vehicleID, from, to, limit)
}

// GetUnprocessedAlerts retrieves unprocessed alerts for a vehicle
func (s *AlertService) GetUnprocessedAlerts(ctx context.Context, vehicleID string) ([]*models.VehicleAlert, error) {
	return s.repo.GetUnprocessedAlerts(ctx, vehicleID)
}

// MarkAlertProcessed marks an alert as processed
func (s *AlertService) MarkAlertProcessed(ctx context.Context, alertID string) error {
	return s.repo.MarkAlertProcessed(ctx, alertID)
}

// GetAlertCount returns the count of alerts for a vehicle
func (s *AlertService) GetAlertCount(ctx context.Context, vehicleID string, from, to time.Time) (int64, error) {
	return s.repo.GetAlertCount(ctx, vehicleID, from, to)
}

// DeleteOldAlerts removes alerts older than the specified date
func (s *AlertService) DeleteOldAlerts(ctx context.Context, before time.Time, processedOnly bool) (int64, error) {
	return s.repo.DeleteOldAlerts(ctx, before, processedOnly)
}
