package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ReconciliationSummary struct {
	TenantID   uuid.UUID
	AutoHealed int
	Quarantined int
	Errors     int
}

const ReconciliationBatchSize = 100

// Reconcile executes a throttled, batch-aware integrity scan.
func (s *AssignmentService) Reconcile(ctx context.Context, tenantID uuid.UUID) (*ReconciliationSummary, error) {
	summary := &ReconciliationSummary{TenantID: tenantID}
	s.logger.Info("starting throttled reconciliation scan (V3)", zap.String("tenant_id", tenantID.String()))

	// 1. Process Stale Cache in Batches
	for {
		count, err := s.healBatch(ctx, tenantID)
		if err != nil {
			s.logger.Error("batch healing failed", zap.Error(err))
			summary.Errors++
			break
		}
		summary.AutoHealed += count
		if count < ReconciliationBatchSize {
			break // Finished all stale records
		}
		
		// Throttling to prevent DB storm
		s.logger.Debug("throttling reconciliation batch", zap.Int("healed", count))
		time.Sleep(300 * time.Millisecond)
	}

	// 2. Process Conflicts (Quarantine)
	qCount, err := s.quarantineConflictsV3(ctx, tenantID)
	if err != nil {
		summary.Errors++
	} else {
		summary.Quarantined = qCount
	}

	return summary, nil
}

func (s *AssignmentService) healBatch(ctx context.Context, tenantID uuid.UUID) (int, error) {
	// Atomic Batch Healing with Limit
	query := `
		WITH target_drivers AS (
			SELECT d.id
			FROM drivers d
			LEFT JOIN vehicles v ON d.current_vehicle_id = v.id
			WHERE d.tenant_id = $1
			  AND d.current_vehicle_id IS NOT NULL
			  AND (v.id IS NULL OR v.current_driver_id != d.id OR v.deleted_at IS NOT NULL)
			LIMIT $2
		)
		UPDATE drivers d
		SET current_vehicle_id = NULL, 
		    version = version + 1,
		    updated_at = NOW()
		FROM target_drivers td
		WHERE d.id = td.id
		RETURNING d.id`

	rows, err := s.db.Query(ctx, query, tenantID, ReconciliationBatchSize)
	if err != nil { return 0, err }
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	return count, nil
}

func (s *AssignmentService) quarantineConflictsV3(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `
		WITH conflicts AS (
			SELECT v.id
			FROM vehicles v
			JOIN drivers d ON v.current_driver_id = d.id
			WHERE v.tenant_id = $1
			  AND v.current_driver_id IS NOT NULL
			  AND d.current_vehicle_id != v.id
			  AND v.deleted_at IS NULL
			  AND d.deleted_at IS NULL
		)
		UPDATE vehicles v
		SET assignment_quarantine_reason = 'CONFLICT: SoT/Cache mismatch',
		    updated_at = NOW()
		FROM conflicts c
		WHERE v.id = c.id
		RETURNING v.id`

	rows, err := s.db.Query(ctx, query, tenantID)
	if err != nil { return 0, err }
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	return count, nil
}
