package models

import (
	"time"
)

type EmergencyHaltRequest struct {
	Reason       string `json:"reason" binding:"required"`
	AuthorizedBy string `json:"authorized_by"`
}

type EmergencyStatus struct {
	Active     bool       `json:"active"`
	LastHaltAt *time.Time `json:"last_halt_at,omitempty"`
	Reason     string     `json:"reason,omitempty"`
	IssuedBy   string     `json:"issued_by,omitempty"`
	CommandID  string     `json:"command_id,omitempty"`
}

type HaltCommand struct {
	CommandID string     `json:"command_id"`
	Type      string     `json:"type"` // e.g. "GLOBAL_HALT"
	IssuedAt  time.Time  `json:"issued_at"`
	Reason    string     `json:"reason"`
	Issuer    string     `json:"issuer"`
	ClearedAt *time.Time `json:"cleared_at,omitempty"`
	ClearedBy string     `json:"cleared_by,omitempty"`
}
