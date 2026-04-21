package models

import (
	"time"

	"github.com/google/uuid"
)

type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
	InvitationExpired  InvitationStatus = "expired"
	InvitationRevoked  InvitationStatus = "revoked"
)

type Invitation struct {
	ID            uuid.UUID        `json:"id"`
	TenantID         uuid.UUID        `json:"tenant_id"`
	Email         string           `json:"email"`
	InvitedBy     uuid.UUID        `json:"invited_by"`
	Token         string           `json:"-"`
	RoleIDs       []uuid.UUID      `json:"role_ids"`
	Message       string           `json:"message,omitempty"`
	Status        InvitationStatus `json:"status"`
	ExpiresAt     time.Time        `json:"expires_at"`
	AcceptedAt    *time.Time       `json:"accepted_at,omitempty"`
	AcceptedUserID *uuid.UUID      `json:"accepted_user_id,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}
