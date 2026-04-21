package models

import (
	"time"

	"github.com/google/uuid"
)

type AuthToken struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	TokenHash    string     `json:"-"`
	TokenType    string     `json:"token_type"` // access, refresh
	DeviceID     string     `json:"device_id"`
	DeviceInfo   string     `json:"device_info"`
	IPAddress    string     `json:"ip_address"`
	UserAgent    string     `json:"user_agent"`
	ExpiresAt    time.Time  `json:"expires_at"`
	LastActiveAt time.Time  `json:"last_active_at"`
	RevokedAt    *time.Time `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
}

type Session struct {
	ID          string    `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	Device      string    `json:"device"`
	IP          string    `json:"ip"`
	CreatedAt   time.Time `json:"created_at"`
	LastActive  time.Time `json:"last_active"`
}
