package dto

import (
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/alert/domains/notifications/models"
	"github.com/google/uuid"
)

// ========== REQUEST DTOs ==========

type ResendRequest struct {
	Reason string `json:"reason" binding:"required,min=3"`
	Force  bool   `json:"force"`
}

type SendNotificationRequest struct {
	Recipient string                 `json:"recipient" binding:"required,email"`
	Channel   string                 `json:"channel" binding:"required,oneof=EMAIL SMS PUSH IN_APP"`
	Subject   string                 `json:"subject"`
	Message   string                 `json:"message" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata"`
	Priority  string                 `json:"priority" binding:"omitempty,oneof=HIGH STANDARD LOW"`
}

// SettingUpdate represents an update to a single notification setting
type SettingUpdate struct {
	ID              uuid.UUID                  `json:"id" validate:"required"`
	Channel         models.NotificationChannel `json:"channel" validate:"required,oneof=EMAIL SMS PUSH IN_APP"`
	Enabled         bool                       `json:"enabled"`
	MaxRetries      int                        `json:"max_retries" validate:"min=0,max=10"`
	Priority        string                     `json:"priority" validate:"omitempty,oneof=HIGH STANDARD LOW"`
	RetentionMonths int                        `json:"retention_months" validate:"min=1,max=24"`
}

// ToDomain converts the DTO to a domain model
func (s SettingUpdate) ToDomain() *models.NotificationSettings {
	return &models.NotificationSettings{
		ID:              s.ID,
		Channel:         s.Channel,
		Enabled:         s.Enabled,
		MaxRetries:      s.MaxRetries,
		Priority:        s.Priority,
		RetentionMonths: s.RetentionMonths,
	}
}

// SettingsUpdateRequest is a collection of setting updates
type SettingsUpdateRequest []SettingUpdate

// PreferenceUpdateRequest is a DTO for updating user preferences
type PreferenceUpdateRequest struct {
	EmailEnabled bool   `json:"email_enabled"`
	SmsEnabled   bool   `json:"sms_enabled"`
	PushEnabled  bool   `json:"push_enabled"`
	InAppEnabled bool   `json:"in_app_enabled"`
	Reason       string `json:"reason" binding:"required,min=5"`
}

// GlobalDefaultUpdateRequest is a DTO for updating global notification defaults
type GlobalDefaultUpdateRequest struct {
	EmailEnabled bool `json:"email_enabled"`
	SmsEnabled   bool `json:"sms_enabled"`
	PushEnabled  bool `json:"push_enabled"`
	InAppEnabled bool `json:"in_app_enabled"`
}

// ========== RESPONSE DTOs ==========

type ResendResponse struct {
	NewNotificationID uuid.UUID `json:"new_notification_id"`
	Message           string    `json:"message"`
}

type NotificationEventDTO struct {
	ID             uuid.UUID `json:"id"`
	NotificationID uuid.UUID `json:"notification_id"`
	EventType      string    `json:"event_type"`
	Message        string    `json:"message"`
	CreatedAt      time.Time `json:"created_at"`
}

// NotificationLogDTO represents a notification log entry in API responses
// Excludes internal audit fields (CreatedBy, UpdatedBy, UserID) for security
type NotificationLogDTO struct {
	ID               uuid.UUID                  `json:"id"`
	UserEmail        string                     `json:"user_email"`
	Channel          models.NotificationChannel `json:"channel"`
	Type             string                     `json:"type"`
	Subject          string                     `json:"subject"`
	Body             string                     `json:"body,omitempty"`
	Status           models.NotificationStatus  `json:"status"`
	Metadata         map[string]any             `json:"metadata,omitempty"`
	CreatedAt        time.Time                  `json:"created_at"`
	RetryCount       int                        `json:"retry_count,omitempty"`
	MaxRetries       int                        `json:"max_retries,omitempty"`
	Provider         *string                    `json:"provider,omitempty"`
	ProviderResponse *string                    `json:"provider_response,omitempty"`
	Items            []NotificationEventDTO     `json:"items,omitempty"`
}

// ToNotificationLogDTO converts a model to a DTO
func ToNotificationLogDTO(log *models.NotificationLog) *NotificationLogDTO {
	if log == nil {
		return nil
	}
	dto := &NotificationLogDTO{
		ID:               log.ID,
		UserEmail:        log.UserEmail,
		Channel:          log.Channel,
		Type:             log.Type,
		Subject:          log.Subject,
		Body:             log.Body,
		Status:           log.Status,
		Metadata:         log.Metadata,
		CreatedAt:        log.CreatedAt,
		RetryCount:       log.RetryCount,
		MaxRetries:       log.MaxRetries,
		Provider:         log.Provider,
		ProviderResponse: log.ProviderResponse,
	}
	// Convert events
	if len(log.Items) > 0 {
		dto.Items = make([]NotificationEventDTO, len(log.Items))
		for i, event := range log.Items {
			dto.Items[i] = NotificationEventDTO{
				ID:             event.ID,
				NotificationID: event.NotificationID,
				EventType:      event.EventType,
				Message:        event.Message,
				CreatedAt:      event.CreatedAt,
			}
		}
	}
	return dto
}

type ListLogsResponse struct {
	Data  []NotificationLogDTO `json:"data"`
	Total int                  `json:"total"`
	Page  int                  `json:"page"`
	Limit int                  `json:"limit"`
}

type SettingsResponse struct {
	ID              uuid.UUID                  `json:"id"`
	Channel         models.NotificationChannel `json:"channel"`
	Enabled         bool                       `json:"enabled"`
	MaxRetries      int                        `json:"max_retries"`
	Priority        string                     `json:"priority"`
	RetentionMonths int                        `json:"retention_months"`
	UpdatedAt       time.Time                  `json:"updated_at"`
}

// ToSettingsResponse converts a model to a DTO
func ToSettingsResponse(settings []models.NotificationSettings) []SettingsResponse {
	res := make([]SettingsResponse, len(settings))
	for i, s := range settings {
		res[i] = SettingsResponse{
			ID:              s.ID,
			Channel:         s.Channel,
			Enabled:         s.Enabled,
			MaxRetries:      s.MaxRetries,
			Priority:        s.Priority,
			RetentionMonths: s.RetentionMonths,
			UpdatedAt:       s.UpdatedAt,
		}
	}
	return res
}

type UserPreferenceResponse struct {
	UserID         uuid.UUID `json:"user_id"`
	UserName       string    `json:"user_name,omitempty"`
	UserEmail      string    `json:"user_email,omitempty"`
	EmailEnabled   bool      `json:"email_enabled"`
	SmsEnabled     bool      `json:"sms_enabled"`
	PushEnabled    bool      `json:"push_enabled"`
	InAppEnabled   bool      `json:"in_app_enabled"`
	LastModifiedAt string    `json:"last_modified_at"`
}

type PreferenceListResponse struct {
	Items []UserPreferenceResponse `json:"items"`
	Total int                      `json:"total"`
	Page  int                      `json:"page"`
	Limit int                      `json:"limit"`
}

type GlobalDefaultResponse struct {
	ID           int       `json:"id"`
	EmailEnabled bool      `json:"email_enabled"`
	SmsEnabled   bool      `json:"sms_enabled"`
	PushEnabled  bool      `json:"push_enabled"`
	InAppEnabled bool      `json:"in_app_enabled"`
	UpdatedBy    uuid.UUID `json:"updated_by"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ToGlobalDefaultResponse converts a model to a DTO
func ToGlobalDefaultResponse(d *models.GlobalNotificationDefault) *GlobalDefaultResponse {
	if d == nil {
		return nil
	}
	return &GlobalDefaultResponse{
		ID:           d.ID,
		EmailEnabled: d.EmailEnabled,
		SmsEnabled:   d.SmsEnabled,
		PushEnabled:  d.PushEnabled,
		InAppEnabled: d.InAppEnabled,
		UpdatedBy:    d.UpdatedBy,
		UpdatedAt:    d.UpdatedAt,
	}
}
