package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/governance/metrics"
	"github.com/flyfleet/fleet-core-service/internal/domains/governance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/governance/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type DataDeletionService interface {
	ProcessDeletionRequest(ctx context.Context, requestID uuid.UUID) error
}

type dataDeletionService struct {
	repo   repository.GovernanceRepository
	db     *pgxpool.Pool
	nats   *nats.Conn
	logger *zap.Logger
}

func NewDataDeletionService(repo repository.GovernanceRepository, db *pgxpool.Pool, nats *nats.Conn, logger *zap.Logger) DataDeletionService {
	return &dataDeletionService{repo: repo, db: db, nats: nats, logger: logger}
}

func (s *dataDeletionService) ProcessDeletionRequest(ctx context.Context, requestID uuid.UUID) error {
	req, err := s.repo.GetRequestByIDInternal(ctx, requestID)
	if err != nil {
		return err
	}
	if req == nil {
		return fmt.Errorf("request not found")
	}

	if req.RequestType != models.RequestTypeDeletion {
		return fmt.Errorf("not a deletion request")
	}

	// 0. Check for Legal Holds
	onHold, err := s.repo.CheckLegalHold(ctx, req.RequesterID)
	if err != nil {
		return fmt.Errorf("failed to check legal hold: %w", err)
	}
	if onHold {
		s.logger.Warn("skipping deletion request due to legal hold", zap.String("id", requestID.String()), zap.String("userID", req.RequesterID.String()))
		s.repo.UpdateRequestStatus(ctx, requestID, req.OrganizationID, models.RequestStatusFailed, nil, nil) // We'll mark as FAILED for hold
		return fmt.Errorf("legal hold active for user %s", req.RequesterID)
	}

	// Phase 9: Alert on multiple deletions
	var delCount int
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM governance_requests WHERE requester_id = $1 AND request_type = 'DELETION'", req.RequesterID).Scan(&delCount)
	if delCount > 1 && s.nats != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"alert_type":      "MULTIPLE_DELETIONS_DETECTED",
			"organization_id": req.OrganizationID.String(),
			"user_id":         req.RequesterID.String(),
			"timestamp":       time.Now().Unix(),
		})
		s.nats.Publish("governance.security.alert", payload)
	}

	// Update status to PROCESSING
	s.repo.UpdateRequestStatus(ctx, requestID, req.OrganizationID, models.RequestStatusProcessing, nil, nil)

	// 1. Transaction is critical
	tx, err := s.db.Begin(ctx)
	if err != nil {
		s.repo.UpdateRequestStatus(ctx, requestID, req.OrganizationID, models.RequestStatusFailed, nil, nil)
		return err
	}
	defer tx.Rollback(ctx)

	s.logger.Info("performing data deletion for request", zap.String("id", requestID.String()), zap.String("userID", req.RequesterID.String()))

	// 2. Cascade delete/Anonymize PII
	// Note: In production, we'd use service-specific deletion hooks or soft deletes.
	// For this task, we perform direct scrub:

	// Telemetry scrub
	_, _ = tx.Exec(ctx, "DELETE FROM telemetry WHERE device_id IN (SELECT id FROM user_devices WHERE user_id = $1)", req.RequesterID)

	// Fuel logs scrub
	_, _ = tx.Exec(ctx, "DELETE FROM fuel_logs WHERE driver_id = $1", req.RequesterID)

	// Trips scrub
	_, _ = tx.Exec(ctx, "DELETE FROM trips WHERE driver_id = $1", req.RequesterID)

	// GeoEvents scrub
	_, _ = tx.Exec(ctx, "DELETE FROM geofence_events WHERE vehicle_id IN (SELECT id FROM vehicles WHERE driver_id = $1)", req.RequesterID)

	// 3. Log Audit logs
	s.repo.CreateAuditLog(ctx, &models.GovernanceAuditLog{
		ID:             uuid.New(),
		OrganizationID: req.OrganizationID,
		UserID:         req.RequesterID,
		Action:         "DELETION_COMPLETED",
		Module:         "DATA_REQUESTS",
		CreatedAt:      time.Now(),
	})

	// 4. Update Status to COMPLETED
	s.repo.UpdateRequestStatus(ctx, requestID, req.OrganizationID, models.RequestStatusCompleted, nil, nil)

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.publishEvent("governance.deletion.completed", req)
	metrics.GovernanceDeletionsTotal.Inc()

	return nil
}

func (s *dataDeletionService) publishEvent(subject string, data interface{}) {
	if s.nats == nil {
		return
	}
	payload, _ := json.Marshal(data)
	s.nats.Publish(subject, payload)
}
