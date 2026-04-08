package service

import (
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/emergency/models"
	"go.uber.org/zap"
)

func TestAuthorityRouting_DeterminesCorrectUnits(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewEmergencyAuthorityService(nil, nil, logger)

	incident := &models.SOSIncident{
		Severity: models.SeverityCritical,
	}

	authorities := svc.determineAuthorities(incident)

	expected := map[AuthorityType]bool{
		AuthorityPolice:    true,
		AuthorityAmbulance: true,
		AuthorityFire:      true,
	}

	if len(authorities) != 3 {
		t.Errorf("Expected 3 authorities for critical incident, got %d", len(authorities))
	}

	for _, a := range authorities {
		if !expected[a] {
			t.Errorf("Unexpected authority: %s", a)
		}
	}
}

func TestAuthorityRouting_HighSeverity(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewEmergencyAuthorityService(nil, nil, logger)

	incident := &models.SOSIncident{
		Severity: models.SeverityHigh,
	}

	authorities := svc.determineAuthorities(incident)

	if len(authorities) != 2 {
		t.Errorf("Expected 2 authorities for high severity incident, got %d", len(authorities))
	}
}

func TestAuthorityRouting_LowSeverity(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewEmergencyAuthorityService(nil, nil, logger)

	incident := &models.SOSIncident{
		Severity: models.SeverityLow,
	}

	authorities := svc.determineAuthorities(incident)

	if len(authorities) != 1 {
		t.Errorf("Expected 1 authority for low severity incident, got %d", len(authorities))
	}

	if authorities[0] != AuthorityPolice {
		t.Errorf("Expected Police for low severity, got %s", authorities[0])
	}
}
