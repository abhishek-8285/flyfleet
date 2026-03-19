package models

import (
	"time"

	"github.com/google/uuid"
)

// DVIRReport represents a Driver Vehicle Inspection Report
type DVIRReport struct {
	ID             uuid.UUID `json:"id" db:"id"`
	DriverID       uuid.UUID `json:"driver_id" db:"driver_id"`
	VehicleID      uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	InspectionType string    `json:"inspection_type" db:"inspection_type"` // PRE_TRIP, POST_TRIP
	Condition      string    `json:"condition" db:"condition"`             // SAFE, UNSAFE
	Notes          string    `json:"notes" db:"notes"`
	Signature      string    `json:"signature" db:"signature"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// DVIRDefect represents an individual defect filed in a DVIR Report
type DVIRDefect struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	DVIRID      uuid.UUID  `json:"dvir_id" db:"dvir_id"`
	DefectType  string     `json:"defect_type" db:"defect_type"`
	Description string     `json:"description" db:"description"`
	Resolved    bool       `json:"resolved" db:"resolved"`
	ResolvedBy  *uuid.UUID `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
}
