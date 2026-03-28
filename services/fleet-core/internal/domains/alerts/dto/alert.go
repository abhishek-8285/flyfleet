package dto

import (
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/models"
	"github.com/google/uuid"
)

// ========== REQUEST DTOs ==========

// AlertRequestDTO is the request payload for triggering an alert
// Note: OrganizationID is injected from context, not from client
type AlertRequestDTO struct {
	UserID    *uuid.UUID           `json:"user_id,omitempty"`
	Recipient string               `json:"recipient,omitempty"` // For non-user alerts
	AlertType string               `json:"alert_type" validate:"required"`
	Priority  models.AlertPriority `json:"priority" validate:"required,oneof=CRITICAL HIGH NORMAL LOW"`
	Data      map[string]any       `json:"data" validate:"required"`
	Metadata  map[string]any       `json:"metadata,omitempty"`
}

// ToModel converts the DTO to a domain model, setting the organization ID
func (r AlertRequestDTO) ToModel(orgID uuid.UUID) models.AlertRequest {
	return models.AlertRequest{
		UserID:         r.UserID,
		Recipient:      r.Recipient,
		AlertType:      r.AlertType,
		Priority:       r.Priority,
		Data:           r.Data,
		Metadata:       r.Metadata,
		OrganizationID: orgID, // Injected from context, never from client
	}
}

// UserPreferenceUpdateDTO represents an update to user notification preferences
type UserPreferenceUpdateDTO struct {
	Channel models.Channel `json:"channel" validate:"required,oneof=EMAIL SMS PUSH IN_APP WEBHOOK"`
	Enabled bool           `json:"enabled"`
}

// ToModel converts the DTO to a domain model
func (r UserPreferenceUpdateDTO) ToModel(userID, orgID uuid.UUID) models.UserPreference {
	return models.UserPreference{
		UserID:         userID,
		OrganizationID: orgID,
		Channel:        r.Channel,
		Enabled:        r.Enabled,
	}
}

// CreateTemplateRequestDTO represents a request to create a notification template
type CreateTemplateRequestDTO struct {
	Name     string                 `json:"name" validate:"required"`
	Channel  models.Channel         `json:"channel" validate:"required,oneof=EMAIL SMS PUSH IN_APP WEBHOOK"`
	Language string                 `json:"language" validate:"required"`
	Subject  string                 `json:"subject" validate:"required"`
	Body     string                 `json:"body" validate:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToModel converts the DTO to a domain model, setting the organization ID
func (r CreateTemplateRequestDTO) ToModel(orgID uuid.UUID) models.Template {
	return models.Template{
		Name:           r.Name,
		Channel:        r.Channel,
		Language:       r.Language,
		Subject:        r.Subject,
		Body:           r.Body,
		Metadata:       r.Metadata,
		OrganizationID: orgID, // Injected from context
	}
}

// ========== RESPONSE DTOs ==========

// UserPreferenceResponseDTO represents a user preference in API responses
// Exposes only safe fields, excludes internal database structure
type UserPreferenceResponseDTO struct {
	UserID    uuid.UUID      `json:"user_id"`
	Channel   models.Channel `json:"channel"`
	Enabled   bool           `json:"enabled"`
	UpdatedAt string         `json:"updated_at"` // ISO 8601 format
}

// ToResponseDTO converts a model to a response DTO
func ToUserPreferenceResponseDTO(p models.UserPreference) UserPreferenceResponseDTO {
	return UserPreferenceResponseDTO{
		UserID:    p.UserID,
		Channel:   p.Channel,
		Enabled:   p.Enabled,
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}

// TemplateResponseDTO represents a template in API responses
type TemplateResponseDTO struct {
	ID             uuid.UUID      `json:"id"`
	Name           string         `json:"name"`
	Channel        models.Channel `json:"channel"`
	Language       string         `json:"language"`
	Subject        string         `json:"subject,omitempty"`
	Body           string         `json:"body"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

// ToTemplateResponseDTO converts a model to a response DTO
func ToTemplateResponseDTO(t models.Template) TemplateResponseDTO {
	return TemplateResponseDTO{
		ID:             t.ID,
		Name:           t.Name,
		Channel:        t.Channel,
		Language:       t.Language,
		Subject:        t.Subject,
		Body:           t.Body,
		Metadata:       t.Metadata,
		OrganizationID: t.OrganizationID,
		CreatedAt:      t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      t.UpdatedAt.Format(time.RFC3339),
	}
}
