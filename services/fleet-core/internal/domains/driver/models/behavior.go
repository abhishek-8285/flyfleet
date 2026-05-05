package models

import (
	"time"

	"github.com/google/uuid"
)

// DriverBehaviorLog represents a log of changes to a driver's behavior score
type DriverBehaviorLog struct {
	ID        uuid.UUID `json:"id" db:"id"`
	DriverID  uuid.UUID `json:"driver_id" db:"driver_id" index:"true" references:"drivers"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id" index:"true" references:"tenants"`
	OldScore  int       `json:"old_score" db:"old_score"`
	NewScore  int       `json:"new_score" db:"new_score"`
	Reason    string    `json:"reason" db:"reason"`
	EventType string    `json:"event_type" db:"event_type"`
	UserID    uuid.UUID `json:"user_id" db:"user_id" references:"users"`
	UserEmail string    `json:"user_email" db:"user_email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
