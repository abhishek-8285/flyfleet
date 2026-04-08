package models

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	ThemeColor  string     `json:"theme_color,omitempty"`
	ThemeLogoURL string    `json:"theme_logo_url,omitempty"`
	DefaultRoleID *uuid.UUID `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type UserOrganization struct {
	UserID     uuid.UUID  `json:"user_id"`
	OrgID      uuid.UUID  `json:"org_id"`
	RoleID     uuid.UUID  `json:"role_id"`
	Role       string     `json:"role"`
	CreatedAt  time.Time  `json:"created_at"`
}

type OrgInfo struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Role          string    `json:"role"`
	VehicleCount  int       `json:"vehicle_count"`
	ThemeColor    string    `json:"theme_color"`
}
