package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type EmergencyService struct {
	repo repository.EmergencyRepository
	nc   *nats.Conn
	log  *zap.Logger
}

func NewEmergencyService(repo repository.EmergencyRepository, nc *nats.Conn, log *zap.Logger) *EmergencyService {
	return &EmergencyService{
		repo: repo,
		nc:   nc,
		log:  log,
	}
}

func (s *EmergencyService) TriggerSOS(ctx context.Context, req models.CreateSOSRequest) (*models.SOSIncident, error) {
	vehicleID, err := uuid.Parse(req.VehicleID)
	if err != nil {
		return nil, fmt.Errorf("invalid vehicle_id: %w", err)
	}

	// 1. Create Incident Model
	incident := &models.SOSIncident{
		ID:              uuid.New(),
		VehicleID:       vehicleID,
		Status:          models.StatusTriggered,
		Severity:        req.Severity,
		Timestamp:       time.Now(),
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		LocationAddress: req.LocationAddress,
		Metadata:        req.Metadata,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if req.DriverID != "" {
		id, parseErr := uuid.Parse(req.DriverID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid driver_id: %w", parseErr)
		}
		incident.DriverID = &id
	}
	if req.TripID != "" {
		id, parseErr := uuid.Parse(req.TripID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid trip_id: %w", parseErr)
		}
		incident.TripID = &id
	}

	if incident.Severity == "" {
		incident.Severity = models.SeverityCritical // Default to Critical
	}

	// 2. Persist to DB
	if err := s.repo.CreateIncident(ctx, incident); err != nil {
		s.log.Error("Failed to create SOS incident", zap.Error(err))
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}

	// 3. Log Action
	s.logAction(ctx, incident.ID, nil, "TRIGGERED", "SOS Triggered by Device/App")

	// 4. Publish to NATS (Real-time Alert)
	s.publishEvent("emergency.sos.triggered", incident)

	return incident, nil
}

func (s *EmergencyService) AcknowledgeIncident(ctx context.Context, id uuid.UUID, userID uuid.UUID, notes string) error {
	incident, err := s.repo.GetIncidentByID(ctx, id)
	if err != nil {
		return err
	}

	if incident.Status != models.StatusTriggered {
		return fmt.Errorf("incident is already %s", incident.Status)
	}

	if err := s.repo.UpdateIncidentStatus(ctx, id, models.StatusAcknowledged); err != nil {
		return err
	}

	s.logAction(ctx, id, &userID, "ACKNOWLEDGED", notes)
	
	// Publish update
	incident.Status = models.StatusAcknowledged
	s.publishEvent("emergency.sos.acknowledged", incident)

	return nil
}

func (s *EmergencyService) ResolveIncident(ctx context.Context, id uuid.UUID, userID uuid.UUID, notes string, status models.SOSStatus) error {
	if status != models.StatusResolved && status != models.StatusFalseAlarm {
		return fmt.Errorf("invalid resolution status")
	}

	if err := s.repo.UpdateIncidentStatus(ctx, id, status); err != nil {
		return err
	}

	s.logAction(ctx, id, &userID, string(status), notes)
	
	// Publish update
	// Fetch updated incident to publish full state? Or just event?
	s.publishEvent("emergency.sos.resolved", map[string]interface{}{
		"id": id,
		"status": status,
		"resolved_by": userID,
		"notes": notes,
	})

	return nil
}

func (s *EmergencyService) ListActiveIncidents(ctx context.Context) ([]models.SOSIncident, error) {
	return s.repo.ListActiveIncidents(ctx)
}

func (s *EmergencyService) logAction(ctx context.Context, incidentID uuid.UUID, userID *uuid.UUID, action, notes string) {
	log := &models.IncidentLog{
		ID:         uuid.New(),
		IncidentID: incidentID,
		UserID:     userID,
		Action:     action,
		Notes:      notes,
		Timestamp:  time.Now(),
	}
	if err := s.repo.LogIncidentAction(ctx, log); err != nil {
		s.log.Error("Failed to log incident action", zap.Error(err))
	}
}

func (s *EmergencyService) publishEvent(subject string, data interface{}) {
	if s.nc == nil {
		s.log.Warn("Skipping emergency event publish: no NATS connection", zap.String("subject", subject))
		return
	}
	payload, _ := json.Marshal(data)
	if err := s.nc.Publish(subject, payload); err != nil {
		s.log.Error("Failed to publish NATS event", zap.String("subject", subject), zap.Error(err))
	}
}
