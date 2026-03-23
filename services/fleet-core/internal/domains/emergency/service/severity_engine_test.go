package service

import (
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
)

func TestEmergencySeverityEngine_Classify(t *testing.T) {
	engine := NewEmergencySeverityEngine()

	tests := []struct {
		name               string
		speed              float64
		collisionDetected  bool
		panicButton        bool
		driverHealthSignal bool
		expected           models.SOSSeverity
	}{
		{"Critical: high speed collision", 50, true, true, true, models.SeverityCritical},
		{"Critical: collision with health loss", 10, true, true, false, models.SeverityCritical},
		{"High: low speed collision", 20, true, true, true, models.SeverityHigh},
		{"High: panic at high speed", 70, false, true, true, models.SeverityHigh},
		{"Medium: panic button", 30, false, true, true, models.SeverityMedium},
		{"Medium: extreme speed", 110, false, false, true, models.SeverityMedium},
		{"Low: default", 10, false, false, true, models.SeverityLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.Classify(tt.speed, tt.collisionDetected, tt.panicButton, tt.driverHealthSignal)
			if got != tt.expected {
				t.Errorf("Classify() = %v, want %v", got, tt.expected)
			}
		})
	}
}
