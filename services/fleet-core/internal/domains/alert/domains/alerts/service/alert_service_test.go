package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/alerts/models"
)

func TestAlertTypesConstants(t *testing.T) {
	// Test that constants are properly defined
	assert.Equal(t, "overspeed", models.AlertTypeOverspeed)
	assert.Equal(t, "geofence_entry", models.AlertTypeGeofenceEntry)
	assert.Equal(t, "geofence_exit", models.AlertTypeGeofenceExit)
	assert.Equal(t, "critical", models.SeverityCritical)
	assert.Equal(t, "info", models.SeverityInfo)
	assert.Equal(t, "new", models.AlertStatusNew)
	assert.Equal(t, "acknowledged", models.AlertStatusAcknowledged)
	assert.Equal(t, "resolved", models.AlertStatusResolved)
}

func TestCreateRuleRequestValidation(t *testing.T) {
	req := &models.CreateRuleRequest{
		Name:        "Valid Rule",
		Type:        models.AlertTypeOverspeed,
		Severity:    models.SeverityWarning,
		Conditions:  map[string]interface{}{"speed_limit": 80},
		Actions:     []models.AlertAction{{Type: models.ActionTypeNotification, Target: "test@example.com"}},
	}

	assert.Equal(t, "Valid Rule", req.Name)
	assert.Equal(t, models.AlertTypeOverspeed, req.Type)
	assert.Equal(t, models.SeverityWarning, req.Severity)
	assert.Len(t, req.Actions, 1)
	assert.NotNil(t, req.Conditions)
}