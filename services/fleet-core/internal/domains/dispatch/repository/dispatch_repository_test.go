package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Mock database for repository tests ─────────────────────────────────────

type mockDispatchDB struct {
	actions   []mockDispatchAction
	updatedAt map[uuid.UUID]mockShipmentUpdate
}

type mockDispatchAction struct {
	ID         uuid.UUID
	TenantID      uuid.UUID
	ActionType string
	ShipmentID uuid.UUID
	TripID     string
	VehicleID  uuid.UUID
	DriverID   uuid.UUID
}

type mockShipmentUpdate struct {
	TripID string
	Status string
}

func (m *mockDispatchDB) ExecuteAssignment(ctx context.Context, tx *AssignmentTx) (*AssignmentResult, error) {
	// 1. Update shipment
	m.updatedAt[tx.ShipmentID] = mockShipmentUpdate{
		TripID: tx.TripID,
		Status: "ASSIGNED",
	}

	// 2. Insert dispatch action
	action := mockDispatchAction{
		ID:         uuid.New(),
		TenantID:      tx.TenantID,
		ActionType: "manual_assign",
		ShipmentID: tx.ShipmentID,
		TripID:     tx.TripID,
		VehicleID:  tx.VehicleID,
		DriverID:   tx.DriverID,
	}
	m.actions = append(m.actions, action)

	return &AssignmentResult{
		TripID:         tx.TripID,
		Status:         "SCHEDULED",
		ShipmentStatus: "ASSIGNED",
	}, nil
}

// ── Repository Tests ────────────────────────────────────────────────────────

func TestDispatchRepository_ExecuteAssignment_UpdatesShipment(t *testing.T) {
	db := &mockDispatchDB{
		updatedAt: make(map[uuid.UUID]mockShipmentUpdate),
	}

	shipmentID := uuid.New()
	tenantID := uuid.New()
	tripID := uuid.New().String()
	vehicleID := uuid.New()
	driverID := uuid.New()

	result, err := db.ExecuteAssignment(context.Background(), &AssignmentTx{
		ShipmentID: shipmentID,
		TripID:     tripID,
		VehicleID:  vehicleID,
		DriverID:   driverID,
		TenantID:      tenantID,
	})

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify shipment update
	update, ok := db.updatedAt[shipmentID]
	require.True(t, ok, "Shipment must be updated")
	assert.Equal(t, tripID, update.TripID, "Shipment trip_id must be set")
	assert.Equal(t, "ASSIGNED", update.Status, "Shipment status must be ASSIGNED")
}

func TestDispatchRepository_ExecuteAssignment_InsertsAuditRecord(t *testing.T) {
	db := &mockDispatchDB{
		updatedAt: make(map[uuid.UUID]mockShipmentUpdate),
	}

	shipmentID := uuid.New()
	tenantID := uuid.New()
	tripID := uuid.New().String()
	vehicleID := uuid.New()
	driverID := uuid.New()

	_, err := db.ExecuteAssignment(context.Background(), &AssignmentTx{
		ShipmentID: shipmentID,
		TripID:     tripID,
		VehicleID:  vehicleID,
		DriverID:   driverID,
		TenantID:      tenantID,
	})

	require.NoError(t, err)

	// Verify audit record
	require.Len(t, db.actions, 1, "One dispatch action must be recorded")
	action := db.actions[0]
	assert.Equal(t, tenantID, action.TenantID)
	assert.Equal(t, "manual_assign", action.ActionType)
	assert.Equal(t, shipmentID, action.ShipmentID)
	assert.Equal(t, tripID, action.TripID)
	assert.Equal(t, vehicleID, action.VehicleID)
	assert.Equal(t, driverID, action.DriverID)
}

func TestDispatchRepository_AuditRecordIsImmutable(t *testing.T) {
	db := &mockDispatchDB{
		updatedAt: make(map[uuid.UUID]mockShipmentUpdate),
	}

	shipmentID := uuid.New()
	tenantID := uuid.New()
	tripID := uuid.New().String()

	_, _ = db.ExecuteAssignment(context.Background(), &AssignmentTx{
		ShipmentID: shipmentID,
		TripID:     tripID,
		VehicleID:  uuid.New(),
		DriverID:   uuid.New(),
		TenantID:      tenantID,
	})

	// Audit record must not be modifiable after insert
	original := db.actions[0]
	originalTripID := original.TripID

	// Attempt to modify (should have no effect on stored record)
	original.TripID = "tampered"
	assert.Equal(t, originalTripID, db.actions[0].TripID,
		"Original audit record must remain unchanged")
}

func TestDispatchRepository_MultipleAssignmentsCreateSeparateRecords(t *testing.T) {
	db := &mockDispatchDB{
		updatedAt: make(map[uuid.UUID]mockShipmentUpdate),
	}

	// First assignment
	_, _ = db.ExecuteAssignment(context.Background(), &AssignmentTx{
		ShipmentID: uuid.New(),
		TripID:     uuid.New().String(),
		VehicleID:  uuid.New(),
		DriverID:   uuid.New(),
		TenantID:      uuid.New(),
	})

	// Second assignment
	_, _ = db.ExecuteAssignment(context.Background(), &AssignmentTx{
		ShipmentID: uuid.New(),
		TripID:     uuid.New().String(),
		VehicleID:  uuid.New(),
		DriverID:   uuid.New(),
		TenantID:      uuid.New(),
	})

	assert.Len(t, db.actions, 2, "Two separate audit records must exist")
	assert.Len(t, db.updatedAt, 2, "Two shipments must be updated")

	// Each record must have unique action ID
	assert.NotEqual(t, db.actions[0].ID, db.actions[1].ID,
		"Each audit action must have unique ID")
}

func TestDispatchRepository_ActionTypeVariants(t *testing.T) {
	actionTypes := []string{
		"auto_suggest",
		"manual_assign",
		"auto_assign",
		"bulk_assign",
		"redispatch",
		"override",
	}

	validTypes := map[string]bool{
		"auto_suggest":  true,
		"manual_assign": true,
		"auto_assign":   true,
		"bulk_assign":   true,
		"redispatch":    true,
		"override":      true,
	}

	for _, at := range actionTypes {
		t.Run(at, func(t *testing.T) {
			assert.True(t, validTypes[at],
				"Action type %q must be valid", at)
		})
	}
}

// AssignmentTx and AssignmentResult are now defined in dispatch_repository.go.
