package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	OrgID        uuid.UUID `json:"org_id"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    uuid.UUID `json:"created_by"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    uuid.UUID `json:"updated_by"`
}

type SecurityEvent struct {
	UserID    uuid.UUID
	EventType string
	IPAddress string
	UserAgent string
	Metadata  string
}

type AuthResponse struct {
	AccessToken  string
	RefreshToken string
	User         User
}
