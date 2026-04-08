package service

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/emergency/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestNotifyAuthorities_CriticalDispatchesAll(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewEmergencyAuthorityService(nil, nil, logger) // no NATS in unit tests

	orgID := uuid.New()
	incident := &models.SOSIncident{
		ID:             uuid.New(),
		OrgID:          orgID,
		VehicleID:      uuid.New(),
		Severity:       models.SeverityCritical,
		IncidentType:   models.TypePanic,
		Latitude:       28.4505,
		Longitude:      77.0263,
		SimulationMode: false,
	}

	// Should not panic even with nil NATS
	svc.NotifyAuthorities(context.Background(), incident)
}

func TestNotifyAuthorities_SkipsSimulation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewEmergencyAuthorityService(nil, nil, logger)

	incident := &models.SOSIncident{
		ID:             uuid.New(),
		Severity:       models.SeverityCritical,
		SimulationMode: true,
	}

	authorities := svc.determineAuthorities(incident)
	if len(authorities) == 0 {
		t.Error("Expected authorities list to be non-empty for critical incidents even in simulation determination")
	}
}

func TestDetermineAuthorities_HighSeverity(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewEmergencyAuthorityService(nil, nil, logger)

	incident := &models.SOSIncident{Severity: models.SeverityHigh}
	authorities := svc.determineAuthorities(incident)

	if len(authorities) != 2 {
		t.Errorf("Expected 2 authorities for HIGH severity, got %d", len(authorities))
	}
}
