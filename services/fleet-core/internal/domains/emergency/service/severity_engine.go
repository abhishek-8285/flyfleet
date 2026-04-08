package service

import (
	"github.com/flyfleet/fleet-core/internal/domains/emergency/models"
)

type EmergencySeverityEngine struct{}

func NewEmergencySeverityEngine() *EmergencySeverityEngine {
	return &EmergencySeverityEngine{}
}

// Classify determines the severity of an incident based on various signals.
func (e *EmergencySeverityEngine) Classify(speed float64, collisionDetected, panicButton, driverHealthSignal bool) models.SOSSeverity {
	// Critical: Collision at high speed or collision + driver health signal loss
	if collisionDetected && (speed > 40 || !driverHealthSignal) {
		return models.SeverityCritical
	}

	// High: Collision at low speed or panic button + high speed
	if collisionDetected || (panicButton && speed > 60) {
		return models.SeverityHigh
	}

	// Medium: Panic button pressed or high speed without collision
	if panicButton || speed > 100 {
		return models.SeverityMedium
	}

	// Low: Default for other triggers
	return models.SeverityLow
}
