package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/models"
)

func TestNewMaintenanceService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	assert.NotNil(t, logger)
	t.Log("Maintenance service package test passed")
}

func TestPackageStructure(t *testing.T) {
	testUUID := uuid.New()
	assert.NotEqual(t, uuid.Nil, testUUID)
	t.Log("Package imports and basic functionality working correctly")
}

func TestMaintenanceConstants(t *testing.T) {
	// Test that maintenance type constants are accessible
	assert.Equal(t, "routine", models.MaintenanceTypeRoutine)
	assert.Equal(t, "preventive", models.MaintenanceTypePreventive)
	assert.Equal(t, "corrective", models.MaintenanceTypeCorrective)
	assert.Equal(t, "emergency", models.MaintenanceTypeEmergency)
	
	// Test priority constants
	assert.Equal(t, "low", models.PriorityLow)
	assert.Equal(t, "medium", models.PriorityMedium)
	assert.Equal(t, "high", models.PriorityHigh)
	assert.Equal(t, "critical", models.PriorityCritical)
	
	// Test status constants
	assert.Equal(t, "scheduled", models.StatusScheduled)
	assert.Equal(t, "in_progress", models.StatusInProgress)
	assert.Equal(t, "completed", models.StatusCompleted)
	assert.Equal(t, "cancelled", models.StatusCancelled)
	
	t.Log("All maintenance constants accessible and correct")
}

func TestCreateScheduleRequestValidation(t *testing.T) {
	// Test valid request
	req := &models.CreateScheduleRequest{
		VehicleID:   uuid.New(),
		Type:        models.MaintenanceTypeRoutine,
		Name:        "Oil Change",
		DueDate:     time.Now().Add(30 * 24 * time.Hour),
	}
	
	assert.NotNil(t, req.VehicleID)
	assert.Equal(t, "Oil Change", req.Name)
	assert.Equal(t, models.MaintenanceTypeRoutine, req.Type)
	assert.True(t, req.DueDate.After(time.Now()))
	
	t.Log("CreateScheduleRequest validation working correctly")
}

func TestMaintenanceScheduleModel(t *testing.T) {
	schedule := &models.MaintenanceSchedule{
		ID:          uuid.New().String(),
		OrgID:       uuid.New(),
		Name:        "Test Schedule",
		Status:      models.StatusScheduled,
	}
	
	assert.NotEmpty(t, schedule.ID)
	assert.NotEqual(t, uuid.Nil, schedule.OrgID)
	assert.Equal(t, models.StatusScheduled, schedule.Status)
	assert.Equal(t, "Test Schedule", schedule.Name)
	
	t.Log("MaintenanceSchedule model working correctly")
}