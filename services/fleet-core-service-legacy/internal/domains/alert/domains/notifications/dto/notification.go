package dto

import (
	"github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/notifications/models"
	"github.com/google/uuid"
)

type ResendRequest struct {
	Reason string `json:"reason" binding:"required,min=3"`
	Force  bool   `json:"force"`
}

type ResendResponse struct {
	NewNotificationID uuid.UUID `json:"new_notification_id"`
	Message           string    `json:"message"`
}

type SettingsUpdateRequest []models.NotificationSettings

type ListLogsResponse struct {
	Data  []models.NotificationLog `json:"data"`
	Total int                      `json:"total"`
	Page  int                      `json:"page"`
	Limit int                      `json:"limit"`
}

type PreferenceUpdateRequest struct {
	EmailEnabled bool   `json:"email_enabled"`
	SmsEnabled   bool   `json:"sms_enabled"`
	PushEnabled  bool   `json:"push_enabled"`
	InAppEnabled bool   `json:"in_app_enabled"`
	Reason       string `json:"reason" binding:"required,min=5"`
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

type GlobalDefaultUpdateRequest struct {
	EmailEnabled bool `json:"email_enabled"`
	SmsEnabled   bool `json:"sms_enabled"`
	PushEnabled  bool `json:"push_enabled"`
	InAppEnabled bool `json:"in_app_enabled"`
}
type SendNotificationRequest struct {
	Recipient string                 `json:"recipient" binding:"required,email"`
	Channel   string                 `json:"channel" binding:"required,oneof=EMAIL SMS PUSH IN_APP"`
	Subject   string                 `json:"subject"`
	Message   string                 `json:"message" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata"`
	Priority  string                 `json:"priority" binding:"omitempty,oneof=HIGH STANDARD LOW"`
}
