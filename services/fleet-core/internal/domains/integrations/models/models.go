package models

import (
	"time"

	"github.com/google/uuid"
)

type Integration struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	OrgID     uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name      string     `json:"name" db:"name"`
	Type      string     `json:"type" db:"type"`
	Provider  string     `json:"provider" db:"provider"`
	Status    string     `json:"status" db:"status"`
	Config    any        `json:"config" db:"config"`
	LastSync  *time.Time `json:"last_sync" db:"last_sync"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	CreatedBy uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy uuid.UUID  `json:"updated_by" db:"updated_by"`
}

type APIClient struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	OrgID            uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	ClientID         string     `json:"client_id" db:"client_id"`
	ClientSecretHash string     `json:"-" db:"client_secret_hash"`
	Name             string     `json:"name" db:"name"`
	Scopes           []string   `json:"scopes" db:"scopes"`
	Expiry           *time.Time `json:"expiry" db:"expiry"`
	Status           string     `json:"status" db:"status"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	CreatedBy        uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy        uuid.UUID  `json:"updated_by" db:"updated_by"`
}

type Webhook struct {
	ID        uuid.UUID `json:"id" db:"id"`
	OrgID     uuid.UUID `json:"tenant_id" db:"tenant_id"`
	URL       string    `json:"url" db:"url"`
	Events    []string  `json:"events" db:"events"`
	Secret    string    `json:"secret" db:"secret"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy uuid.UUID `json:"updated_by" db:"updated_by"`
}

type IntegrationLog struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OrgID         uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	IntegrationID *uuid.UUID `json:"integration_id" db:"integration_id"`
	Method        string     `json:"method" db:"method"`
	Endpoint      string     `json:"endpoint" db:"endpoint"`
	RequestBody   any        `json:"request_body" db:"request_body"`
	ResponseBody  any        `json:"response_body" db:"response_body"`
	StatusCode    int        `json:"status_code" db:"status_code"`
	Status        string     `json:"status" db:"status"`
	DurationMS    int        `json:"duration_ms" db:"duration_ms"`
	Timestamp     time.Time  `json:"timestamp" db:"timestamp"`
}
