package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type FuelSettings struct {
	OrganizationID     uuid.UUID       `json:"organization_id" db:"organization_id" validate:"required,uuid"`
	DefaultCurrency    string          `json:"default_currency" db:"default_currency" validate:"required,len=3"`
	AllowedFuelTypes   json.RawMessage `json:"allowed_fuel_types" db:"allowed_fuel_types"`
	ValidationRules    json.RawMessage `json:"validation_rules" db:"validation_rules"`
	AutoSyncEnabled    bool            `json:"auto_sync_enabled" db:"auto_sync_enabled"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
}

type UpdateFuelSettingsRequest struct {
	DefaultCurrency    *string          `json:"default_currency,omitempty" validate:"omitempty,len=3"`
	AllowedFuelTypes   *json.RawMessage `json:"allowed_fuel_types,omitempty"`
	ValidationRules    *json.RawMessage `json:"validation_rules,omitempty"`
	AutoSyncEnabled    *bool            `json:"auto_sync_enabled,omitempty"`
}

type FuelAuditLog struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	FuelLogID uuid.UUID       `json:"fuel_log_id" db:"fuel_log_id"`
	Action    string          `json:"action" db:"action"`
	ChangedBy uuid.UUID       `json:"changed_by" db:"changed_by"`
	ChangedAt time.Time       `json:"changed_at" db:"changed_at"`
	OldValues json.RawMessage `json:"old_values" db:"old_values"`
	NewValues json.RawMessage `json:"new_values" db:"new_values"`
	Diff      json.RawMessage `json:"diff" db:"diff"`
}
