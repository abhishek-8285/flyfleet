package validation

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/flyfleet/maintenance-service/internal/models"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateCreateScheduleRequest validates the create schedule request
func ValidateCreateScheduleRequest(req *models.CreateScheduleRequest) error {
	if req.VehicleID == [16]byte{} {
		return errors.New("vehicle_id is required")
	}

	if req.Name == "" {
		return errors.New("name is required")
	}

	if len(req.Name) > 100 {
		return errors.New("name must be less than 100 characters")
	}

	if req.Description != "" && len(req.Description) > 500 {
		return errors.New("description must be less than 500 characters")
	}

	if req.DueDate.IsZero() {
		return errors.New("due_date is required")
	}

	if req.DueDate.Before(time.Now()) {
		return errors.New("due_date must be in the future")
	}

	// Validate type
	validTypes := map[string]bool{
		models.MaintenanceTypeRoutine:    true,
		models.MaintenanceTypePreventive: true,
		models.MaintenanceTypeCorrective: true,
		models.MaintenanceTypeEmergency:  true,
	}
	if !validTypes[req.Type] {
		return fmt.Errorf("invalid type: %s", req.Type)
	}

	// Validate priority
	validPriorities := map[string]bool{
		models.PriorityLow:      true,
		models.PriorityMedium:   true,
		models.PriorityHigh:     true,
		models.PriorityCritical: true,
	}
	if !validPriorities[req.Priority] {
		return fmt.Errorf("invalid priority: %s", req.Priority)
	}

	if req.EstimatedCost != nil && *req.EstimatedCost < 0 {
		return errors.New("estimated_cost must be non-negative")
	}

	if req.DueOdometer != nil && *req.DueOdometer < 0 {
		return errors.New("due_odometer must be non-negative")
	}

	if req.DueEngineHours != nil && *req.DueEngineHours < 0 {
		return errors.New("due_engine_hours must be non-negative")
	}

	// Validate service items
	for i, item := range req.ServiceItems {
		if item.Name == "" {
			return fmt.Errorf("service_item[%d].name is required", i)
		}
		if item.Cost < 0 {
			return fmt.Errorf("service_item[%d].cost must be non-negative", i)
		}
	}

	// Validate reminders
	for i, reminder := range req.Reminders {
		if reminder.Type == "" {
			return fmt.Errorf("reminder[%d].type is required", i)
		}
		if reminder.Recipient == "" {
			return fmt.Errorf("reminder[%d].recipient is required", i)
		}
		if !isValidReminderType(reminder.Type) {
			return fmt.Errorf("invalid reminder type: %s", reminder.Type)
		}
	}

	return nil
}

// ValidateUpdateScheduleRequest validates the update schedule request
func ValidateUpdateScheduleRequest(req *models.UpdateScheduleRequest) error {
	if req.Name != nil && *req.Name == "" {
		return errors.New("name cannot be empty")
	}

	if req.Name != nil && len(*req.Name) > 100 {
		return errors.New("name must be less than 100 characters")
	}

	if req.Description != nil && len(*req.Description) > 500 {
		return errors.New("description must be less than 500 characters")
	}

	if req.Type != nil {
		validTypes := map[string]bool{
			models.MaintenanceTypeRoutine:    true,
			models.MaintenanceTypePreventive: true,
			models.MaintenanceTypeCorrective: true,
			models.MaintenanceTypeEmergency:  true,
		}
		if !validTypes[*req.Type] {
			return fmt.Errorf("invalid type: %s", *req.Type)
		}
	}

	if req.Priority != nil {
		validPriorities := map[string]bool{
			models.PriorityLow:      true,
			models.PriorityMedium:   true,
			models.PriorityHigh:     true,
			models.PriorityCritical: true,
		}
		if !validPriorities[*req.Priority] {
			return fmt.Errorf("invalid priority: %s", *req.Priority)
		}
	}

	if req.Status != nil {
		validStatuses := map[string]bool{
			models.StatusScheduled:   true,
			models.StatusInProgress:  true,
			models.StatusCompleted:   true,
			models.StatusCancelled:   true,
		}
		if !validStatuses[*req.Status] {
			return fmt.Errorf("invalid status: %s", *req.Status)
		}
	}

	if req.EstimatedCost != nil && *req.EstimatedCost < 0 {
		return errors.New("estimated_cost must be non-negative")
	}

	if req.ActualCost != nil && *req.ActualCost < 0 {
		return errors.New("actual_cost must be non-negative")
	}

	if req.DueOdometer != nil && *req.DueOdometer < 0 {
		return errors.New("due_odometer must be non-negative")
	}

	if req.DueEngineHours != nil && *req.DueEngineHours < 0 {
		return errors.New("due_engine_hours must be non-negative")
	}

	// Validate service items if provided
	if req.ServiceItems != nil {
		for i, item := range *req.ServiceItems {
			if item.Name == "" {
				return fmt.Errorf("service_item[%d].name is required", i)
			}
			if item.Cost < 0 {
				return fmt.Errorf("service_item[%d].cost must be non-negative", i)
			}
		}
	}

	return nil
}

// ValidateCreateLogRequest validates the create log request
func ValidateCreateLogRequest(req *models.CreateLogRequest) error {
	if req.VehicleID == [16]byte{} {
		return errors.New("vehicle_id is required")
	}

	if req.Name == "" {
		return errors.New("name is required")
	}

	if len(req.Name) > 100 {
		return errors.New("name must be less than 100 characters")
	}

	if req.Description != "" && len(req.Description) > 500 {
		return errors.New("description must be less than 500 characters")
	}

	if req.StartDate.IsZero() {
		return errors.New("start_date is required")
	}

	if req.EndDate.IsZero() {
		return errors.New("end_date is required")
	}

	if req.EndDate.Before(req.StartDate) {
		return errors.New("end_date must be after start_date")
	}

	// Validate type
	validTypes := map[string]bool{
		models.MaintenanceTypeRoutine:    true,
		models.MaintenanceTypePreventive: true,
		models.MaintenanceTypeCorrective: true,
		models.MaintenanceTypeEmergency:  true,
	}
	if !validTypes[req.Type] {
		return fmt.Errorf("invalid type: %s", req.Type)
	}

	if req.Cost < 0 {
		return errors.New("cost must be non-negative")
	}

	if req.Odometer < 0 {
		return errors.New("odometer must be non-negative")
	}

	if req.EngineHours != nil && *req.EngineHours < 0 {
		return errors.New("engine_hours must be non-negative")
	}

	// Validate status
	validStatuses := map[string]bool{
		models.StatusCompleted:           true,
		models.StatusPartiallyCompleted:  true,
	}
	if !validStatuses[req.Status] {
		return fmt.Errorf("invalid status: %s", req.Status)
	}

	// Validate parts used
	for i, part := range req.PartsUsed {
		if part.Name == "" {
			return fmt.Errorf("parts_used[%d].name is required", i)
		}
		if part.Quantity <= 0 {
			return fmt.Errorf("parts_used[%d].quantity must be positive", i)
		}
		if part.UnitPrice < 0 {
			return fmt.Errorf("parts_used[%d].unit_price must be non-negative", i)
		}
	}

	// Validate service items
	for i, item := range req.ServiceItems {
		if item.Name == "" {
			return fmt.Errorf("service_item[%d].name is required", i)
		}
		if item.Cost < 0 {
			return fmt.Errorf("service_item[%d].cost must be non-negative", i)
		}
	}

	return nil
}

// ValidateUpdateLogRequest validates the update log request
func ValidateUpdateLogRequest(req *models.UpdateLogRequest) error {
	if req.Name != nil && *req.Name == "" {
		return errors.New("name cannot be empty")
	}

	if req.Name != nil && len(*req.Name) > 100 {
		return errors.New("name must be less than 100 characters")
	}

	if req.Description != nil && len(*req.Description) > 500 {
		return errors.New("description must be less than 500 characters")
	}

	if req.StartDate != nil && req.EndDate != nil && req.EndDate.Before(*req.StartDate) {
		return errors.New("end_date must be after start_date")
	}

	if req.Type != nil {
		validTypes := map[string]bool{
			models.MaintenanceTypeRoutine:    true,
			models.MaintenanceTypePreventive: true,
			models.MaintenanceTypeCorrective: true,
			models.MaintenanceTypeEmergency:  true,
		}
		if !validTypes[*req.Type] {
			return fmt.Errorf("invalid type: %s", *req.Type)
		}
	}

	if req.Cost != nil && *req.Cost < 0 {
		return errors.New("cost must be non-negative")
	}

	if req.Odometer != nil && *req.Odometer < 0 {
		return errors.New("odometer must be non-negative")
	}

	if req.EngineHours != nil && *req.EngineHours < 0 {
		return errors.New("engine_hours must be non-negative")
	}

	if req.Status != nil {
		validStatuses := map[string]bool{
			models.StatusCompleted:           true,
			models.StatusPartiallyCompleted:  true,
		}
		if !validStatuses[*req.Status] {
			return fmt.Errorf("invalid status: %s", *req.Status)
		}
	}

	// Validate parts used if provided
	if req.PartsUsed != nil {
		for i, part := range *req.PartsUsed {
			if part.Name == "" {
				return fmt.Errorf("parts_used[%d].name is required", i)
			}
			if part.Quantity <= 0 {
				return fmt.Errorf("parts_used[%d].quantity must be positive", i)
			}
			if part.UnitPrice < 0 {
				return fmt.Errorf("parts_used[%d].unit_price must be non-negative", i)
			}
		}
	}

	return nil
}

// isValidReminderType checks if the reminder type is valid
func isValidReminderType(reminderType string) bool {
	validTypes := map[string]bool{
		models.ReminderTypeEmail:         true,
		models.ReminderTypeSMS:           true,
		models.ReminderTypeNotification:  true,
	}
	return validTypes[reminderType]
}

// IsValidEmail checks if the email format is valid
func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// IsValidPhone checks if the phone number format is valid
func IsValidPhone(phone string) bool {
	// Basic phone number validation - can be enhanced based on requirements
	re := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return re.MatchString(phone)
}
