package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/assignment/models"
)

type EntityType int
const (
	TypeVehicle EntityType = 1
	TypeDriver  EntityType = 2
)

type LockTarget struct {
	Type EntityType
	ID   uuid.UUID
}

type AssignmentService struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewAssignmentService(db *pgxpool.Pool, logger *zap.Logger) *AssignmentService {
	return &AssignmentService{
		db:     db,
		logger: logger,
	}
}

// AssignmentContext (Enriched V4)
type AssignmentContext struct {
	TenantID       uuid.UUID
	ActorID        uuid.UUID
	ActorType      string
	SourceIP       string
	RequestID      uuid.UUID
	CorrelationID  uuid.UUID
	IdempotencyKey string
	VehicleVersion int
	DriverVersion  int
	BypassPolicy   bool
	BypassReason   string
}

type ConflictError struct {
	Message         string
	CurrentVersions map[string]int
	VehicleSnapshot json.RawMessage
	DriverSnapshot  json.RawMessage
}
func (e *ConflictError) Error() string { return e.Message }

func (s *AssignmentService) Assign(ctx context.Context, opCtx AssignmentContext, vehicleID, driverID uuid.UUID, remarks string) error {
	// 1. Policy & Eligibility Check (Pre-Transaction)
	if !opCtx.BypassPolicy {
		if err := s.checkPolicy(ctx, opCtx.TenantID, vehicleID, driverID); err != nil {
			return fmt.Errorf("policy violation: %w", err)
		}
	}

	// 2. Idempotency V2 (Payload Hash)
	payload := fmt.Sprintf("%s:%s:%s", opCtx.TenantID, vehicleID, driverID)
	hash := sha256.Sum256([]byte(payload))
	payloadHash := hex.EncodeToString(hash[:])

	// Check Idempotency before locking
	if opCtx.IdempotencyKey != "" {
		var existingHash string
		err := s.db.QueryRow(ctx, `SELECT payload_hash FROM assignment_operations WHERE idempotency_key = $1 AND tenant_id = $2`, opCtx.IdempotencyKey, opCtx.TenantID).Scan(&existingHash)
		if err == nil {
			if existingHash != payloadHash {
				return errors.New("idempotency key reused with different payload")
			}
			return nil // Already processed
		}
	}

	// 3. Prioritized Tuple Locking Orchestration
	targets := []LockTarget{
		{Type: TypeVehicle, ID: vehicleID},
		{Type: TypeDriver, ID: driverID},
	}
	sort.Slice(targets, func(i, j int) bool {
		if targets[i].Type != targets[j].Type { return targets[i].Type < targets[j].Type }
		return targets[i].ID.String() < targets[j].ID.String()
	})

	s.logger.Debug("starting assignment transaction", 
		zap.String("vehicle_id", vehicleID.String()), 
		zap.String("driver_id", driverID.String()),
		zap.Int("req_vehicle_version", opCtx.VehicleVersion),
		zap.Int("req_driver_version", opCtx.DriverVersion))

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil { 
		s.logger.Error("failed to begin transaction", zap.Error(err))
		return err 
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			s.logger.Debug("rollback error (expected if committed)", zap.Error(err))
		}
	}()

	// 4. Acquisition & OCC Validation
	var vBefore, dBefore []byte
	var vVer, dVer int

	for _, target := range targets {
		if target.Type == TypeVehicle {
			err = tx.QueryRow(ctx, `SELECT version, row_to_json(v) FROM vehicles v WHERE id = $1 AND tenant_id = $2 FOR UPDATE`, target.ID, opCtx.TenantID).Scan(&vVer, &vBefore)
			if err != nil { 
				s.logger.Error("failed to lock vehicle", zap.String("id", target.ID.String()), zap.Error(err))
				return err 
			}
			s.logger.Debug("vehicle locked", zap.Int("version", vVer))
			if opCtx.VehicleVersion > 0 && vVer != opCtx.VehicleVersion {
				return &ConflictError{Message: "Vehicle version conflict", CurrentVersions: map[string]int{"vehicle": vVer}, VehicleSnapshot: vBefore}
			}
		} else {
			err = tx.QueryRow(ctx, `SELECT version, row_to_json(d) FROM drivers d WHERE id = $1 AND tenant_id = $2 FOR UPDATE`, target.ID, opCtx.TenantID).Scan(&dVer, &dBefore)
			if err != nil { 
				s.logger.Error("failed to lock driver", zap.String("id", target.ID.String()), zap.Error(err))
				return err 
			}
			s.logger.Debug("driver locked", zap.Int("version", dVer))
			if opCtx.DriverVersion > 0 && dVer != opCtx.DriverVersion {
				return &ConflictError{Message: "Driver version conflict", CurrentVersions: map[string]int{"driver": dVer}, DriverSnapshot: dBefore}
			}
		}
	}

	// 5. Atomic Update with Hybrid Snapshot Marker
	isFullSnapshot := vVer % 10 == 0 // Every 10 versions take a full snapshot

	_, err = tx.Exec(ctx, `UPDATE vehicles SET current_driver_id = $1, version = version + 1, updated_at = NOW() WHERE id = $2`, driverID, vehicleID)
	if err != nil { 
		s.logger.Error("failed to update vehicle", zap.Error(err))
		return err 
	}
	_, err = tx.Exec(ctx, `UPDATE drivers SET current_vehicle_id = $1, version = version + 1, updated_at = NOW() WHERE id = $2`, vehicleID, driverID)
	if err != nil { 
		s.logger.Error("failed to update driver", zap.Error(err))
		return err 
	}

	// 6. Audit Logging (V4 Enriched)
	_, err = tx.Exec(ctx, `
		INSERT INTO assignment_audit_logs (
			tenant_id, vehicle_id, driver_id, from_state, to_state, event, event_type, 
			actor_id, request_id, before_snapshot, is_full_snapshot, remarks
		) VALUES ($1, $2, $3, $4, $5, 'DRIVER_ASSIGNED', 'Assignment.Created', $6, $7, $8, $9, $10)`,
		opCtx.TenantID, vehicleID, driverID, models.StateUnassigned, models.StateAssigned,
		opCtx.ActorID, opCtx.RequestID, vBefore, isFullSnapshot, remarks)
	if err != nil {
		s.logger.Error("failed to log assignment audit", zap.Error(err))
		return fmt.Errorf("failed to log audit: %w", err)
	}

	s.logger.Debug("committing transaction")
	if err := tx.Commit(ctx); err != nil {
		s.logger.Error("failed to commit transaction", zap.Error(err))
		return fmt.Errorf("commit unexpectedly resulted in rollback: %w", err)
	}
	s.logger.Info("assignment successful", zap.String("vehicle_id", vehicleID.String()), zap.String("driver_id", driverID.String()))
	return nil
}

func (s *AssignmentService) Unassign(ctx context.Context, opCtx AssignmentContext, vehicleID uuid.UUID, remarks string) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil { return err }
	defer tx.Rollback(ctx)

	var currentDriverID *uuid.UUID
	var vBefore []byte
	err = tx.QueryRow(ctx, `SELECT current_driver_id, row_to_json(v) FROM vehicles v WHERE id = $1 AND tenant_id = $2 FOR UPDATE`, vehicleID, opCtx.TenantID).Scan(&currentDriverID, &vBefore)
	if err != nil { return err }

	if currentDriverID == nil {
		return nil // Already unassigned
	}

	driverID := *currentDriverID

	// Sort locks (Vehicle priority 1, Driver priority 2)
	// We already have vehicle lock. Now lock driver.
	_, err = tx.Exec(ctx, `SELECT 1 FROM drivers WHERE id = $1 AND tenant_id = $2 FOR UPDATE`, driverID, opCtx.TenantID)
	if err != nil { return err }

	_, err = tx.Exec(ctx, `UPDATE vehicles SET current_driver_id = NULL, version = version + 1, updated_at = NOW() WHERE id = $1`, vehicleID)
	if err != nil { return err }

	_, err = tx.Exec(ctx, `UPDATE drivers SET current_vehicle_id = NULL, version = version + 1, updated_at = NOW() WHERE id = $1`, driverID)
	if err != nil { return err }

	_, err = tx.Exec(ctx, `
		INSERT INTO assignment_audit_logs (
			tenant_id, vehicle_id, driver_id, from_state, to_state, event, event_type, 
			actor_id, request_id, before_snapshot, is_full_snapshot, remarks
		) VALUES ($1, $2, $3, $4, $5, 'DRIVER_UNASSIGNED', 'Assignment.Released', $6, $7, $8, $9, $10)`,
		opCtx.TenantID, vehicleID, driverID, models.StateAssigned, models.StateUnassigned,
		opCtx.ActorID, opCtx.RequestID, vBefore, false, remarks)
	if err != nil {
		s.logger.Error("failed to log unassignment audit", zap.Error(err))
		return fmt.Errorf("failed to log audit: %w", err)
	}

	s.logger.Debug("committing transaction")
	if err := tx.Commit(ctx); err != nil {
		s.logger.Error("failed to commit transaction", zap.Error(err))
		return fmt.Errorf("commit unexpectedly resulted in rollback: %w", err)
	}
	return nil
}

func (s *AssignmentService) checkPolicy(ctx context.Context, tenantID uuid.UUID, vehicleID, driverID uuid.UUID) error {
	// Simple Eligibility Check (Compliance Status)
	var vStatus, dStatus string
	err := s.db.QueryRow(ctx, `SELECT compliance_status FROM vehicles WHERE id = $1 AND tenant_id = $2`, vehicleID, tenantID).Scan(&vStatus)
	if err != nil { return err }
	if vStatus == "NON_COMPLIANT" { return errors.New("vehicle is non-compliant (maintenance/permit block)") }

	err = s.db.QueryRow(ctx, `SELECT compliance_status FROM drivers WHERE id = $1 AND tenant_id = $2`, driverID, tenantID).Scan(&dStatus)
	if err != nil { return err }
	if dStatus == "NON_COMPLIANT" { return errors.New("driver is non-compliant (license/fatigue block)") }

	return nil
}
