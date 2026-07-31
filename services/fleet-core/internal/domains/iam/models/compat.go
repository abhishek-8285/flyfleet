package models

import "github.com/google/uuid"

// Legacy compatibility types for auth->iam domain migration

type SecurityEvent struct {
	UserID     uuid.UUID
	EventType  string
	ActorID    *uuid.UUID
	OnBehalfOf *uuid.UUID
}

type SigningKey struct {
	KID           string
	PrivateKeyEnc string
}

type Organization struct {
	ID   uuid.UUID
	Type string
	Name string
}

type RefreshToken struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Token  string
}
