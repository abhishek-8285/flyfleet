package validation

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/flyfleet/maintenance-service/internal/models"
)

func TestValidateCreateScheduleRequest(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.CreateScheduleRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "valid request",
			request: &models.CreateScheduleRequest{
				VehicleID:   uuid.New(),
				Type:        models.MaintenanceTypePreventive,
				Name:        "Test Schedule",
				DueDate:     time.Now().AddDate(0, 0, 7), // 7 days from now
				Priority:    models.PriorityHigh,
			},
			expectError: false,
		},
		{
			name: "missing vehicle_id",
			request: &models.CreateScheduleRequest{
				Type:     models.MaintenanceTypePreventive,
				Name:     "Test Schedule",
				DueDate:  time.Now().AddDate(0, 0, 7),
				Priority: models.PriorityHigh,
			},
			expectError:   true,
			errorContains: "vehicle_id is required",
		},
		{
			name: "missing name",
			request: &models.CreateScheduleRequest{
				VehicleID: uuid.New(),
				Type:      models.MaintenanceTypePreventive,
				DueDate:   time.Now().AddDate(0, 0, 7),
				Priority:  models.PriorityHigh,
			},
			expectError:   true,
			errorContains: "name is required",
		},
		{
			name: "invalid type",
			request: &models.CreateScheduleRequest{
				VehicleID: uuid.New(),
				Type:      "invalid_type",
				Name:      "Test Schedule",
				DueDate:   time.Now().AddDate(0, 0, 7),
				Priority:  models.PriorityHigh,
			},
			expectError:   true,
			errorContains: "invalid type",
		},
		{
			name: "invalid priority",
			request: &models.CreateScheduleRequest{
				VehicleID: uuid.New(),
				Type:      models.MaintenanceTypePreventive,
				Name:      "Test Schedule",
				DueDate:   time.Now().AddDate(0, 0, 7),
				Priority:  "invalid_priority",
			},
			expectError:   true,
			errorContains: "invalid priority",
		},
		{
			name: "past due date",
			request: &models.CreateScheduleRequest{
				VehicleID: uuid.New(),
				Type:      models.MaintenanceTypePreventive,
				Name:      "Test Schedule",
				DueDate:   time.Now().AddDate(0, 0, -1), // yesterday
				Priority:  models.PriorityHigh,
			},
			expectError:   true,
			errorContains: "due_date must be in the future",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateScheduleRequest(tt.request)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCreateLogRequest(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.CreateLogRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "valid request",
			request: &models.CreateLogRequest{
				VehicleID:   uuid.New(),
				Type:        models.MaintenanceTypePreventive,
				Name:        "Test Log",
				StartDate:   time.Now().Add(-24 * time.Hour),
				EndDate:     time.Now(),
				PerformedBy: uuid.New(),
				Cost:        100.0,
				Odometer:    50000,
				Status:      models.StatusCompleted,
			},
			expectError: false,
		},
		{
			name: "missing vehicle_id",
			request: &models.CreateLogRequest{
				Type:        models.MaintenanceTypePreventive,
				Name:        "Test Log",
				StartDate:   time.Now().Add(-24 * time.Hour),
				EndDate:     time.Now(),
				PerformedBy: uuid.New(),
				Cost:        100.0,
				Odometer:    50000,
				Status:      models.StatusCompleted,
			},
			expectError:   true,
			errorContains: "vehicle_id is required",
		},
		{
			name: "missing name",
			request: &models.CreateLogRequest{
				VehicleID:   uuid.New(),
				Type:        models.MaintenanceTypePreventive,
				StartDate:   time.Now().Add(-24 * time.Hour),
				EndDate:     time.Now(),
				PerformedBy: uuid.New(),
				Cost:        100.0,
				Odometer:    50000,
				Status:      models.StatusCompleted,
			},
			expectError:   true,
			errorContains: "name is required",
		},
		{
			name: "invalid type",
			request: &models.CreateLogRequest{
				VehicleID:   uuid.New(),
				Type:        "invalid_type",
				Name:        "Test Log",
				StartDate:   time.Now().Add(-24 * time.Hour),
				EndDate:     time.Now(),
				PerformedBy: uuid.New(),
				Cost:        100.0,
				Odometer:    50000,
				Status:      models.StatusCompleted,
			},
			expectError:   true,
			errorContains: "invalid type",
		},
		{
			name: "invalid status",
			request: &models.CreateLogRequest{
				VehicleID:   uuid.New(),
				Type:        models.MaintenanceTypePreventive,
				Name:        "Test Log",
				StartDate:   time.Now().Add(-24 * time.Hour),
				EndDate:     time.Now(),
				PerformedBy: uuid.New(),
				Cost:        100.0,
				Odometer:    50000,
				Status:      "invalid_status",
			},
			expectError:   true,
			errorContains: "invalid status",
		},
		{
			name: "end date before start date",
			request: &models.CreateLogRequest{
				VehicleID:   uuid.New(),
				Type:        models.MaintenanceTypePreventive,
				Name:        "Test Log",
				StartDate:   time.Now(),
				EndDate:     time.Now().Add(-24 * time.Hour), // before start
				PerformedBy: uuid.New(),
				Cost:        100.0,
				Odometer:    50000,
				Status:      models.StatusCompleted,
			},
			expectError:   true,
			errorContains: "end_date must be after start_date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateLogRequest(tt.request)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailValidation(t *testing.T) {
	assert.True(t, IsValidEmail("test@example.com"))
	assert.True(t, IsValidEmail("user.name+tag@example.co.uk"))
	assert.False(t, IsValidEmail("invalid-email"))
	assert.False(t, IsValidEmail("test@"))
	assert.False(t, IsValidEmail("@example.com"))
}

func TestPhoneValidation(t *testing.T) {
	assert.True(t, IsValidPhone("+1234567890"))
	assert.True(t, IsValidPhone("1234567890"))
	assert.True(t, IsValidPhone("+1234567890"))
	assert.False(t, IsValidPhone("+1-234-567-8900"))
	assert.False(t, IsValidPhone("abc-def-ghij"))
	assert.False(t, IsValidPhone(""))
	assert.False(t, IsValidPhone("0"))
	assert.False(t, IsValidPhone("+"))

}