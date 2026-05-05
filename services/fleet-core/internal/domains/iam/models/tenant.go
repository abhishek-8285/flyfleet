package models

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	ThemeColor  string     `json:"theme_color,omitempty"`
	ThemeLogoURL string    `json:"theme_logo_url,omitempty"`
	DefaultRoleID *uuid.UUID `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type UserTenant struct {
	UserID     uuid.UUID  `json:"user_id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	RoleID     uuid.UUID  `json:"role_id"`
	Role       string     `json:"role"`
	CreatedAt  time.Time  `json:"created_at"`
}

type TenantInfo struct {
	ID            uuid.UUID `json:"tenant_id"`
	Name          string    `json:"tenant_name"`
	Role          string    `json:"role"`
	VehicleCount  int       `json:"vehicle_count"`
	ThemeColor    string    `json:"theme_color"`
}
