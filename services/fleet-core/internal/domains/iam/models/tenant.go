package models

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Name          string     `json:"name" db:"name"`
	ThemeColor    string     `json:"theme_color,omitempty" db:"theme_color"`
	ThemeLogoURL  string     `json:"theme_logo_url,omitempty" db:"theme_logo_url"`
	DefaultRoleID *uuid.UUID `json:"-" db:"default_role_id"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type UserTenant struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id" index:"true" references:"users"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id" index:"true" references:"tenants"`
	RoleID    uuid.UUID `json:"role_id" db:"role_id" index:"true" references:"roles"`
	Role      string    `json:"role" db:"role"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type TenantInfo struct {
	ID            uuid.UUID `json:"tenant_id"`
	Name          string    `json:"tenant_name"`
	Role          string    `json:"role"`
	VehicleCount  int       `json:"vehicle_count"`
	ThemeColor    string    `json:"theme_color"`
}
