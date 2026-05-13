package models

import (
	"time"

	"github.com/google/uuid"
)

type AssignmentState string

const (
	StateUnassigned       AssignmentState = "UNASSIGNED"
	StatePending          AssignmentState = "PENDING_ASSIGNMENT"
	StateAssigned         AssignmentState = "ASSIGNED"
	StateTransferPending  AssignmentState = "TRANSFER_PENDING"
	StateSuspended        AssignmentState = "SUSPENDED"
	StateRevoked          AssignmentState = "REVOKED"
	StateSoftBlocked      AssignmentState = "SOFT_BLOCKED"
)

type AssignmentEvent string

const (
	EventDriverAssigned          AssignmentEvent = "DRIVER_ASSIGNED"
	EventDriverUnassigned        AssignmentEvent = "DRIVER_UNASSIGNED"
	EventVehicleDeleted          AssignmentEvent = "VEHICLE_DELETED"
	EventDriverDeleted           AssignmentEvent = "DRIVER_DELETED"
	EventConflictDetected        AssignmentEvent = "CONFLICT_DETECTED"
	EventIntegrityRepairExecuted AssignmentEvent = "INTEGRITY_REPAIR_EXECUTED"
)

type AssignmentAuditLog struct {
	ID         uuid.UUID       `json:"id"`
	TenantID   uuid.UUID       `json:"tenant_id"`
	VehicleID  uuid.UUID       `json:"vehicle_id"`
	DriverID   uuid.UUID       `json:"driver_id"`
	FromState  AssignmentState `json:"from_state"`
	ToState    AssignmentState `json:"to_state"`
	Event      AssignmentEvent `json:"event"`
	ActorID    uuid.UUID       `json:"actor_id"`
	Remarks    string          `json:"remarks"`
	Timestamp  time.Time       `json:"timestamp"`
}
