package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors struct {
	Error            string            `json:"error"`
	ValidationErrors []ValidationError `json:"validation_errors"`
}

// Validator provides request validation
type Validator struct{}

// NewValidator creates a new Validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateCreateTripRequest validates the create trip request
func (v *Validator) ValidateCreateTripRequest(data map[string]interface{}) []ValidationError {
	var errors []ValidationError

	// Required fields
	if _, ok := data["vehicle_id"]; !ok {
		errors = append(errors, ValidationError{Field: "vehicle_id", Message: "vehicle_id is required"})
	} else if !v.isValidUUID(data["vehicle_id"]) {
		errors = append(errors, ValidationError{Field: "vehicle_id", Message: "vehicle_id must be a valid UUID"})
	}

	if _, ok := data["start_location"]; !ok {
		errors = append(errors, ValidationError{Field: "start_location", Message: "start_location is required"})
	} else if !v.isValidString(data["start_location"], 1, 255) {
		errors = append(errors, ValidationError{Field: "start_location", Message: "start_location must be 1-255 characters"})
	}

	if _, ok := data["end_location"]; !ok {
		errors = append(errors, ValidationError{Field: "end_location", Message: "end_location is required"})
	} else if !v.isValidString(data["end_location"], 1, 255) {
		errors = append(errors, ValidationError{Field: "end_location", Message: "end_location must be 1-255 characters"})
	}

	if _, ok := data["start_time"]; !ok {
		errors = append(errors, ValidationError{Field: "start_time", Message: "start_time is required"})
	} else if !v.isValidDateTime(data["start_time"]) {
		errors = append(errors, ValidationError{Field: "start_time", Message: "start_time must be a valid ISO 8601 date-time"})
	}

	// Optional fields
	if driverID, ok := data["driver_id"]; ok && driverID != nil {
		if !v.isValidUUID(driverID) {
			errors = append(errors, ValidationError{Field: "driver_id", Message: "driver_id must be a valid UUID"})
		}
	}

	return errors
}

// ValidateUpdateTripRequest validates the update trip request
func (v *Validator) ValidateUpdateTripRequest(data map[string]interface{}) []ValidationError {
	var errors []ValidationError

	// Optional fields with validation
	if driverID, ok := data["driver_id"]; ok && driverID != nil {
		if !v.isValidUUID(driverID) {
			errors = append(errors, ValidationError{Field: "driver_id", Message: "driver_id must be a valid UUID"})
		}
	}

	if loc, ok := data["start_location"]; ok && loc != nil {
		if !v.isValidString(loc, 1, 255) {
			errors = append(errors, ValidationError{Field: "start_location", Message: "start_location must be 1-255 characters"})
		}
	}

	if loc, ok := data["end_location"]; ok && loc != nil {
		if !v.isValidString(loc, 1, 255) {
			errors = append(errors, ValidationError{Field: "end_location", Message: "end_location must be 1-255 characters"})
		}
	}

	if distance, ok := data["distance"]; ok && distance != nil {
		if !v.isValidFloat(distance, 0, 1000000) {
			errors = append(errors, ValidationError{Field: "distance", Message: "distance must be a non-negative number"})
		}
	}

	if status, ok := data["status"]; ok && status != nil {
		if !v.isValidStatus(status) {
			errors = append(errors, ValidationError{Field: "status", Message: "status must be one of: scheduled, in_progress, completed, cancelled"})
		}
	}

	if endTime, ok := data["end_time"]; ok && endTime != nil {
		if !v.isValidDateTime(endTime) {
			errors = append(errors, ValidationError{Field: "end_time", Message: "end_time must be a valid ISO 8601 date-time"})
		}
	}

	return errors
}

// ValidateTripSettings validates trip settings
func (v *Validator) ValidateTripSettings(data map[string]interface{}) []ValidationError {
	var errors []ValidationError

	if retention, ok := data["retention_period"]; ok {
		if !v.isValidInt(retention, 1, 120) {
			errors = append(errors, ValidationError{Field: "retention_period", Message: "retention_period must be between 1 and 120 months"})
		}
	}

	return errors
}

// ValidateListParams validates list query parameters
func (v *Validator) ValidateListParams(page, limit int, status string) []ValidationError {
	var errors []ValidationError

	if page < 1 {
		errors = append(errors, ValidationError{Field: "page", Message: "page must be at least 1"})
	}

	if limit < 1 || limit > 100 {
		errors = append(errors, ValidationError{Field: "limit", Message: "limit must be between 1 and 100"})
	}

	if status != "" && !v.isValidStatusString(status) {
		errors = append(errors, ValidationError{Field: "status", Message: "status must be one of: scheduled, in_progress, completed, cancelled"})
	}

	return errors
}

// ValidateTelemetryParams validates telemetry query parameters
func (v *Validator) ValidateTelemetryParams(limit int) []ValidationError {
	var errors []ValidationError

	if limit < 1 || limit > 500 {
		errors = append(errors, ValidationError{Field: "limit", Message: "limit must be between 1 and 500"})
	}

	return errors
}

// Helper functions
func (v *Validator) isValidUUID(val interface{}) bool {
	str, ok := val.(string)
	if !ok {
		return false
	}
	_, err := uuid.Parse(str)
	return err == nil
}

func (v *Validator) isValidString(val interface{}, minLen, maxLen int) bool {
	str, ok := val.(string)
	if !ok {
		return false
	}
	length := len(str)
	return length >= minLen && length <= maxLen
}

func (v *Validator) isValidDateTime(val interface{}) bool {
	str, ok := val.(string)
	if !ok {
		return false
	}
	_, err := time.Parse(time.RFC3339, str)
	return err == nil
}

func (v *Validator) isValidFloat(val interface{}, min, max float64) bool {
	switch v := val.(type) {
	case float64:
		return v >= min && v <= max
	case float32:
		return float64(v) >= min && float64(v) <= max
	case int:
		return float64(v) >= min && float64(v) <= max
	default:
		return false
	}
}

func (v *Validator) isValidInt(val interface{}, min, max int) bool {
	switch n := val.(type) {
	case float64:
		return int(n) >= min && int(n) <= max
	case int:
		return n >= min && n <= max
	default:
		return false
	}
}

func (v *Validator) isValidStatus(val interface{}) bool {
	str, ok := val.(string)
	if !ok {
		return false
	}
	return v.isValidStatusString(str)
}

func (v *Validator) isValidStatusString(status string) bool {
	validStatuses := []string{"scheduled", "in_progress", "completed", "cancelled"}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// WriteValidationError writes validation errors as JSON response
func WriteValidationError(w http.ResponseWriter, errors []ValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(ValidationErrors{
		Error:            "Validation failed",
		ValidationErrors: errors,
	})
}
