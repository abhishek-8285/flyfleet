package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID        uuid.UUID         `json:"id"`
	UserID    uuid.UUID         `json:"user_id"`
	TenantID     uuid.UUID         `json:"tenant_id"`
	Action    string            `json:"action"`
	Resource  string            `json:"resource"`
	Data      map[string]interface{} `json:"data,omitempty"`
	IP        string            `json:"ip"`
	UserAgent string            `json:"user_agent"`
	CreatedAt time.Time         `json:"created_at"`
}
