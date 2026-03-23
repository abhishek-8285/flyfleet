package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// AutoIncidentService listens for telemetry signals and triggers emergency incidents
type AutoIncidentService struct {
	emergencySvc *EmergencyService
	nc           *nats.Conn
	log          *zap.Logger
}

func NewAutoIncidentService(emergencySvc *EmergencyService, nc *nats.Conn, log *zap.Logger) *AutoIncidentService {
	return &AutoIncidentService{
		emergencySvc: emergencySvc,
		nc:           nc,
		log:          log,
	}
}

// Start begins listening for telemetry events
func (s *AutoIncidentService) Start(ctx context.Context) error {
	s.log.Info("Starting AutoIncidentService telemetry listener")

	// Subscribe to telemetry impact events
	_, err := s.nc.Subscribe("telemetry.v1.impact", func(m *nats.Msg) {
		s.handleImpactEvent(m)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to impact events: %w", err)
	}

	// Subscribe to telemetry rollover events
	_, err = s.nc.Subscribe("telemetry.v1.rollover", func(m *nats.Msg) {
		s.handleRolloverEvent(m)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to rollover events: %w", err)
	}

	// Subscribe to diagnostic fire events
	_, err = s.nc.Subscribe("diagnostic.v1.fire", func(m *nats.Msg) {
		s.handleFireEvent(m)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to fire events: %w", err)
	}

	return nil
}

func (s *AutoIncidentService) handleImpactEvent(m *nats.Msg) {
	var event struct {
		VehicleID string  `json:"vehicle_id"`
		OrgID     string  `json:"org_id"`
		GForce    float64 `json:"g_force"`
		Lat       float64 `json:"lat"`
		Lng       float64 `json:"lng"`
	}

	if err := json.Unmarshal(m.Data, &event); err != nil {
		s.log.Error("Failed to unmarshal impact event", zap.Error(err))
		return
	}

	// Threshold for crash: > 4.0G
	if event.GForce > 4.0 {
		s.log.Warn("High impact detected! Triggering incident", zap.String("vehicle_id", event.VehicleID), zap.Float64("g_force", event.GForce))
		s.triggerAutoIncident(event.VehicleID, event.OrgID, models.TypeVehicleAccident, event.Lat, event.Lng, "Automated crash detection (High G-Force)")
	}
}

func (s *AutoIncidentService) handleRolloverEvent(m *nats.Msg) {
	var event struct {
		VehicleID string  `json:"vehicle_id"`
		OrgID     string  `json:"org_id"`
		Lat       float64 `json:"lat"`
		Lng       float64 `json:"lng"`
	}

	if err := json.Unmarshal(m.Data, &event); err != nil {
		s.log.Error("Failed to unmarshal rollover event", zap.Error(err))
		return
	}

	s.log.Warn("Vehicle rollover detected! Triggering incident", zap.String("vehicle_id", event.VehicleID))
	s.triggerAutoIncident(event.VehicleID, event.OrgID, models.TypeVehicleAccident, event.Lat, event.Lng, "Automated rollover detection")
}

func (s *AutoIncidentService) handleFireEvent(m *nats.Msg) {
	var event struct {
		VehicleID string  `json:"vehicle_id"`
		OrgID     string  `json:"org_id"`
		Lat       float64 `json:"lat"`
		Lng       float64 `json:"lng"`
	}

	if err := json.Unmarshal(m.Data, &event); err != nil {
		s.log.Error("Failed to unmarshal fire event", zap.Error(err))
		return
	}

	s.log.Error("Engine fire detected! Triggering incident", zap.String("vehicle_id", event.VehicleID))
	s.triggerAutoIncident(event.VehicleID, event.OrgID, models.TypeFireAlert, event.Lat, event.Lng, "Automated engine fire detection")
}

func (s *AutoIncidentService) triggerAutoIncident(vehicleID, orgID string, incidentType models.IncidentType, lat, lng float64, notes string) {
	parsedOrgID, _ := uuid.Parse(orgID)
	
	req := models.CreateSOSRequest{
		VehicleID:       vehicleID,
		OrgID:           parsedOrgID,
		IncidentType:    incidentType,
		PanicSource:     models.SourceTelematics,
		Latitude:        lat,
		Longitude:       lng,
		LocationAddress: "Automated Detection",
		Metadata:        json.RawMessage(fmt.Sprintf(`{"auto_notes": "%s"}`, notes)),
	}

	_, err := s.emergencySvc.TriggerSOS(context.Background(), req)
	if err != nil {
		s.log.Error("Failed to trigger automated incident", zap.Error(err), zap.String("type", string(incidentType)))
	}
}
