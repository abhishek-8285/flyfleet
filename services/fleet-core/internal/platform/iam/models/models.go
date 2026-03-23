package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Email          string    `json:"email"`
	FullName       string    `json:"full_name"`
	PasswordHash   string    `json:"-"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type Role struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	Description    string    `json:"description"`
}

type Permission struct {
	ID       uuid.UUID `json:"id"`
	Resource string    `json:"resource"`
	Action   string    `json:"action"`
}
