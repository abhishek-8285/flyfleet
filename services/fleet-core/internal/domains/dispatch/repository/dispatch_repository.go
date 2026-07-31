package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ─── Types ───────────────────────────────────────────────────────────────────

// DispatchAction categorizes every dispatch decision for the audit trail.
type DispatchAction string

const (
	ActionAutoSuggest  DispatchAction = "auto_suggest"
	ActionManualAssign DispatchAction = "manual_assign"
	ActionAutoAssign   DispatchAction = "auto_assign"
	ActionBulkAssign   DispatchAction = "bulk_assign"
	ActionRedispatch   DispatchAction = "redispatch"
	ActionOverride     DispatchAction = "override"
)

// AssignmentTx carries the data needed to atomically link a shipment to a trip
// and record an audit entry in the dispatch_actions table.
type AssignmentTx struct {
	ShipmentID uuid.UUID
	TripID     string
	VehicleID  uuid.UUID
	DriverID   uuid.UUID
	TenantID   uuid.UUID
	ActionType DispatchAction
}

// AssignmentResult is the outcome of ExecuteAssignment.
type AssignmentResult struct {
	TripID         string
	Status         string
	ShipmentStatus string
}

// DispatchSuggestion stores a ranked vehicle/driver recommendation.
type DispatchSuggestion struct {
	TenantID   uuid.UUID
	ShipmentID uuid.UUID
	Rank       int
	VehicleID  *uuid.UUID
	DriverID   *uuid.UUID
	Score      float64
	DistanceKm float64
	EtaMin     int
	Reasons    []string
}

// ─── Interface ───────────────────────────────────────────────────────────────

// DispatchRepository defines the persistence contract for dispatch operations.
type DispatchRepository interface {
	ExecuteAssignment(ctx context.Context, tx *AssignmentTx) (*AssignmentResult, error)
	SaveSuggestion(ctx context.Context, s *DispatchSuggestion) error
	GetActiveTripsForVehicle(ctx context.Context, vehicleID uuid.UUID) ([]uuid.UUID, error)
}

// ─── PostgreSQL Implementation ───────────────────────────────────────────────

type PostgresDispatchRepository struct {
	db *pgxpool.Pool
}

func NewDispatchRepository(db *pgxpool.Pool) DispatchRepository {
	return &PostgresDispatchRepository{db: db}
}

// ExecuteAssignment atomically links a shipment to a trip and records an audit
// entry. Both operations run in a single transaction so the audit trail and
// shipment link stay consistent.
func (r *PostgresDispatchRepository) ExecuteAssignment(ctx context.Context, tx *AssignmentTx) (*AssignmentResult, error) {
	if tx == nil {
		return nil, fmt.Errorf("assignment transaction is nil")
	}

	actionType := tx.ActionType
	if actionType == "" {
		actionType = ActionManualAssign
	}

	const updateShipment = `
		UPDATE customer_v2_shipments
		   SET trip_id = $1, status = 'ASSIGNED', updated_at = NOW()
		 WHERE id = $2 AND tenant_id = $3`
	const insertAction = `
		INSERT INTO dispatch_actions
		    (tenant_id, action_type, shipment_id, trip_id, vehicle_id, driver_id, performed_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`

	dbTx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = dbTx.Rollback(ctx) }()

	if _, err := dbTx.Exec(ctx, updateShipment, tx.TripID, tx.ShipmentID, tx.TenantID); err != nil {
		return nil, fmt.Errorf("failed to link shipment to trip: %w", err)
	}

	if _, err := dbTx.Exec(ctx, insertAction,
		tx.TenantID, actionType, tx.ShipmentID, tx.TripID, tx.VehicleID, tx.DriverID,
	); err != nil {
		return nil, fmt.Errorf("failed to insert dispatch action: %w", err)
	}

	if err := dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &AssignmentResult{
		TripID:         tx.TripID,
		Status:         "SCHEDULED",
		ShipmentStatus: "ASSIGNED",
	}, nil
}

// SaveSuggestion persists a ranked dispatch suggestion for auditing and
// historical analysis.
func (r *PostgresDispatchRepository) SaveSuggestion(ctx context.Context, s *DispatchSuggestion) error {
	reasonsJSON, _ := json.Marshal(s.Reasons)
	const query = `
		INSERT INTO dispatch_suggestions
		    (tenant_id, shipment_id, suggestion_rank, vehicle_id, driver_id, score, distance_km, eta_min, reasons)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	startTime := time.Now()
	_, err := r.db.Exec(ctx, query,
		s.TenantID, s.ShipmentID, s.Rank, s.VehicleID, s.DriverID,
		s.Score, s.DistanceKm, s.EtaMin, reasonsJSON,
	)
	_ = startTime // query timing placeholder
	return err
}

// GetActiveTripsForVehicle returns the IDs of trips that are not in a terminal
// state for the given vehicle (used for availability checks).
func (r *PostgresDispatchRepository) GetActiveTripsForVehicle(ctx context.Context, vehicleID uuid.UUID) ([]uuid.UUID, error) {
	const query = `
		SELECT id FROM trips
		 WHERE vehicle_id = $1
		   AND status IN ('CREATED', 'STARTED', 'PAUSED', 'RESUMED')
		   AND is_deleted = FALSE`

	rows, err := r.db.Query(ctx, query, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active trips: %w", err)
	}
	defer rows.Close()

	var tripIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan trip id: %w", err)
		}
		tripIDs = append(tripIDs, id)
	}
	return tripIDs, rows.Err()
}
