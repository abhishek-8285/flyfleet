package integrations

import (
	"time"

	"github.com/google/uuid"
)

type IntegrationType string

const (
	IntegrationTypeSlack    IntegrationType = "SLACK"
	IntegrationTypeJira     IntegrationType = "JIRA"
	IntegrationTypeWebhook  IntegrationType = "WEBHOOK"
	IntegrationTypePagerDuty IntegrationType = "PAGERDUTY"
)

type IntegrationStatus string

const (
	IntegrationStatusActive   IntegrationStatus = "ACTIVE"
	IntegrationStatusInactive IntegrationStatus = "INACTIVE"
	IntegrationStatusError    IntegrationStatus = "ERROR"
)

type Integration struct {
	ID        uuid.UUID         `json:"id" db:"id"`
	Name      string            `json:"name" db:"name"`
	Type      IntegrationType   `json:"type" db:"type"`
	Config    map[string]string `json:"config" db:"config"`
	Status    IntegrationStatus `json:"status" db:"status"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"`
}

type CreateIntegrationRequest struct {
	Name   string            `json:"name" binding:"required"`
	Type   IntegrationType   `json:"type" binding:"required"`
	Config map[string]string `json:"config" binding:"required"`
}
