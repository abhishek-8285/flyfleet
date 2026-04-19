package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core/internal/domains/vehicle/repository"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// AutoIncidentService listens for telemetry signals and triggers emergency incidents
type AutoIncidentService struct {
	emergencySvc *EmergencyService
	nc           *nats.Conn
	log          *zap.Logger
	vehicleRepo  *repository.VehicleRepository // SECURITY: For org ownership validation
}

func NewAutoIncidentService(emergencySvc *EmergencyService, nc *nats.Conn, vehicleRepo *repository.VehicleRepository, log *zap.Logger) *AutoIncidentService {
	return &AutoIncidentService{
		emergencySvc: emergencySvc,
		nc:           nc,
		vehicleRepo:  vehicleRepo,
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
		OrgID     string  `json:"tenant_id"` // Present but should NOT be trusted
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
		s.log.Warn("High impact detected! Validating org ownership", zap.String("vehicle_id", event.VehicleID), zap.Float64("g_force", event.GForce))

		// SECURITY FIX: Validate tenant_id by querying the vehicle's actual organization
		vehicleID, err := uuid.Parse(event.VehicleID)
		if err != nil {
			s.log.Error("Invalid vehicle_id in impact event", zap.String("raw_vehicle_id", event.VehicleID))
			return
		}

		// Use background context with timeout for database query
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		vehicle, err := s.vehicleRepo.GetByID(ctx, vehicleID)
		if err != nil {
			s.log.Error("Failed to validate vehicle ownership", zap.String("vehicle_id", event.VehicleID), zap.Error(err))
			return
		}

		parsedOrgID, err := uuid.Parse(event.OrgID)
		if err != nil {
			s.log.Error("Invalid tenant_id in impact event", zap.String("vehicle_id", event.VehicleID), zap.String("raw_org_id", event.OrgID))
			return
		}

		// CRITICAL SECURITY CHECK: Ensure the tenant_id from message matches vehicle's actual org
		if vehicle.TenantID != parsedOrgID {
			s.log.Error("Cross-tenant message injection attempt detected!",
				zap.String("vehicle_id", event.VehicleID),
				zap.String("msg_org", event.OrgID),
				zap.String("actual_org", vehicle.TenantID.String()),
				zap.Float64("g_force", event.GForce))
			return // Reject - possible spoofing attack
		}

		s.triggerAutoIncident(event.VehicleID, parsedOrgID.String(), models.TypeVehicleAccident, event.Lat, event.Lng, "Automated crash detection (High G-Force)")
	}
}

func (s *AutoIncidentService) handleRolloverEvent(m *nats.Msg) {
	var event struct {
		VehicleID string  `json:"vehicle_id"`
		OrgID     string  `json:"tenant_id"`
		Lat       float64 `json:"lat"`
		Lng       float64 `json:"lng"`
	}

	if err := json.Unmarshal(m.Data, &event); err != nil {
		s.log.Error("Failed to unmarshal rollover event", zap.Error(err))
		return
	}

	s.log.Warn("Vehicle rollover detected! Validating org ownership", zap.String("vehicle_id", event.VehicleID))

	// SECURITY FIX: Validate tenant_id by querying vehicle
	vehicleID, err := uuid.Parse(event.VehicleID)
	if err != nil {
		s.log.Error("Invalid vehicle_id in rollover event", zap.String("raw_vehicle_id", event.VehicleID))
		return
	}

	// Use background context with timeout for database query
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	vehicle, err := s.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		s.log.Error("Failed to validate vehicle ownership", zap.String("vehicle_id", event.VehicleID), zap.Error(err))
		return
	}

	parsedOrgID, err := uuid.Parse(event.OrgID)
	if err != nil {
		s.log.Error("Invalid tenant_id in rollover event", zap.String("vehicle_id", event.VehicleID), zap.String("raw_org_id", event.OrgID))
		return
	}

	// CRITICAL SECURITY CHECK: Ensure org ownership
	if vehicle.TenantID != parsedOrgID {
		s.log.Error("Cross-tenant message injection attempt detected!",
			zap.String("vehicle_id", event.VehicleID),
			zap.String("msg_org", event.OrgID),
			zap.String("actual_org", vehicle.TenantID.String()))
		return
	}

	s.triggerAutoIncident(event.VehicleID, parsedOrgID.String(), models.TypeVehicleAccident, event.Lat, event.Lng, "Automated rollover detection")
}

func (s *AutoIncidentService) handleFireEvent(m *nats.Msg) {
	var event struct {
		VehicleID string  `json:"vehicle_id"`
		OrgID     string  `json:"tenant_id"`
		Lat       float64 `json:"lat"`
		Lng       float64 `json:"lng"`
	}

	if err := json.Unmarshal(m.Data, &event); err != nil {
		s.log.Error("Failed to unmarshal fire event", zap.Error(err))
		return
	}

	s.log.Error("Engine fire detected! Validating org ownership", zap.String("vehicle_id", event.VehicleID))

	// SECURITY FIX: Validate tenant_id by querying vehicle
	vehicleID, err := uuid.Parse(event.VehicleID)
	if err != nil {
		s.log.Error("Invalid vehicle_id in fire event", zap.String("raw_vehicle_id", event.VehicleID))
		return
	}

	// Use background context with timeout for database query
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	vehicle, err := s.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		s.log.Error("Failed to validate vehicle ownership", zap.String("vehicle_id", event.VehicleID), zap.Error(err))
		return
	}

	parsedOrgID, err := uuid.Parse(event.OrgID)
	if err != nil {
		s.log.Error("Invalid tenant_id in fire event", zap.String("vehicle_id", event.VehicleID), zap.String("raw_org_id", event.OrgID))
		return
	}

	// CRITICAL SECURITY CHECK: Ensure org ownership
	if vehicle.TenantID != parsedOrgID {
		s.log.Error("Cross-tenant message injection attempt detected!",
			zap.String("vehicle_id", event.VehicleID),
			zap.String("msg_org", event.OrgID),
			zap.String("actual_org", vehicle.TenantID.String()))
		return
	}

	s.triggerAutoIncident(event.VehicleID, parsedOrgID.String(), models.TypeFireAlert, event.Lat, event.Lng, "Automated engine fire detection")
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
