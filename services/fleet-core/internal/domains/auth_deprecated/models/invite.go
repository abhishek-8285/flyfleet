package models

import (
	"time"

	"github.com/google/uuid"
)

type Invite struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Code      string    `json:"code" db:"code"`
	OrgID     uuid.UUID `json:"org_id" db:"org_id"`
	Role      string    `json:"role" db:"role"`
	Expiry    time.Time `json:"expiry" db:"expiry"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
	Used      bool      `json:"used" db:"used"`
}
