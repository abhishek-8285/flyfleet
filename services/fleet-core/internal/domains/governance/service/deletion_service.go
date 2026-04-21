package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/governance/metrics"
	"github.com/flyfleet/fleet-core/internal/domains/governance/models"
	"github.com/flyfleet/fleet-core/internal/domains/governance/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type DataDeletionService interface {
	ProcessDeletionRequest(ctx context.Context, requestID uuid.UUID) error
}

type dataDeletionService struct {
	repo   repository.GovernanceRepository
	db     *pgxpool.Pool
	nats   *nats.Conn
	redis  *redis.Client
	logger *zap.Logger
}

func NewDataDeletionService(repo repository.GovernanceRepository, db *pgxpool.Pool, nats *nats.Conn, redis *redis.Client, logger *zap.Logger) DataDeletionService {
	return &dataDeletionService{repo: repo, db: db, nats: nats, redis: redis, logger: logger}
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
			"tenant_id": req.OrganizationID.String(),
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
	// Track if any deletion error occurred
	hadError := false

	// 2. Cascade delete/Anonymize PII
	// Note: In production, we'd use service-specific deletion hooks or soft deletes.
	// For this task, we perform direct scrub:

	// Telemetry scrub
	_, err = tx.Exec(ctx, "DELETE FROM telemetry WHERE device_id IN (SELECT id FROM user_devices WHERE user_id = $1)", req.RequesterID)
	if err != nil {
		s.logger.Error("Failed to delete telemetry", zap.Error(err), zap.String("user_id", req.RequesterID.String()))
		hadError = true
	}

	// Fuel logs scrub
	_, err = tx.Exec(ctx, "DELETE FROM fuel_logs WHERE driver_id = $1", req.RequesterID)
	if err != nil {
		s.logger.Error("Failed to delete fuel_logs", zap.Error(err), zap.String("user_id", req.RequesterID.String()))
		hadError = true
	}

	// Trips scrub
	_, err = tx.Exec(ctx, "DELETE FROM trips WHERE driver_id = $1", req.RequesterID)
	if err != nil {
		s.logger.Error("Failed to delete trips", zap.Error(err), zap.String("user_id", req.RequesterID.String()))
		hadError = true
	}

	// GeoEvents scrub
	_, err = tx.Exec(ctx, "DELETE FROM geofence_events WHERE vehicle_id IN (SELECT id FROM vehicles WHERE driver_id = $1)", req.RequesterID)
	if err != nil {
		s.logger.Error("Failed to delete geofence_events", zap.Error(err), zap.String("user_id", req.RequesterID.String()))
		hadError = true
	}

	// 3. Clean up cached data to prevent PII residuals
	if s.redis != nil {
		s.cleanupUserCache(ctx, req)
	}

	// 4. Check if any deletion errors occurred
	if hadError {
		// Rollback will be handled by defer, but we need to update status to FAILED
		s.repo.UpdateRequestStatus(ctx, requestID, req.OrganizationID, models.RequestStatusFailed, nil, nil)
		s.logger.Error("Deletion failed due to errors in one or more delete operations", zap.String("request_id", requestID.String()))
		return fmt.Errorf("deletion failed due to database errors")
	}

	// 5. All deletions succeeded, commit transaction
	if err := tx.Commit(ctx); err != nil {
		s.repo.UpdateRequestStatus(ctx, requestID, req.OrganizationID, models.RequestStatusFailed, nil, nil)
		return fmt.Errorf("failed to commit deletion transaction: %w", err)
	}

	// 6. Log successful completion
	auditLog := &models.GovernanceAuditLog{
		ID:             uuid.New(),
		OrganizationID: req.OrganizationID,
		UserID:         req.RequesterID,
		Action:         "DELETION_COMPLETED",
		Module:         "DATA_REQUESTS",
		CreatedAt:      time.Now(),
	}
	s.repo.CreateAuditLog(ctx, auditLog)

	// 7. Update Status to COMPLETED
	s.repo.UpdateRequestStatus(ctx, requestID, req.OrganizationID, models.RequestStatusCompleted, nil, nil)

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

// cleanupUserCache removes all Redis cache entries related to the deleted user and their vehicles
func (s *dataDeletionService) cleanupUserCache(ctx context.Context, req *models.GovernanceRequest) {
	tenantID := req.OrganizationID
	userID := req.RequesterID

	// Fetch all vehicles associated with this user to target vehicle-specific cache
	rows, err := s.db.Query(ctx, `
		SELECT id, registration_number
		FROM vehicles
		WHERE driver_id = $1 AND tenant_id = $2`,
		userID, tenantID,
	)
	if err != nil {
		s.logger.Warn("Failed to query vehicles for cache cleanup", zap.Error(err))
		return
	}
	defer rows.Close()

	type VehicleInfo struct {
		ID                 uuid.UUID
		RegistrationNumber string
	}
	var vehicles []VehicleInfo
	for rows.Next() {
		var v VehicleInfo
		if err := rows.Scan(&v.ID, &v.RegistrationNumber); err != nil {
			s.logger.Warn("Failed to scan vehicle", zap.Error(err))
			continue
		}
		vehicles = append(vehicles, v)
	}

	// Delete cache entries for each vehicle
	for _, v := range vehicles {
		// 1. VAHAN verification cache: vahan:verify:{tenantID}:{regNumber}
		vahanKey := fmt.Sprintf("vahan:verify:%s:%s", tenantID.String(), v.RegistrationNumber)
		if err := s.redis.Del(ctx, vahanKey).Err(); err != nil {
			s.logger.Warn("Failed to delete vahan cache", zap.Error(err), zap.String("key", vahanKey))
		}

		// 2. Compliance expired insurance flag: compliance:expired_insurance:vehicle:{vehicleID}
		compKey := fmt.Sprintf("compliance:expired_insurance:vehicle:%s", v.ID.String())
		if err := s.redis.Del(ctx, compKey).Err(); err != nil {
			s.logger.Warn("Failed to delete compliance cache", zap.Error(err), zap.String("key", compKey))
		}

		// 3. Vehicle cache keys pattern: vehicle:{id}:*
		vehiclePattern := fmt.Sprintf("vehicle:%s:*", v.ID.String())
		keys, err := s.redis.Keys(ctx, vehiclePattern).Result()
		if err != nil {
			s.logger.Warn("Failed to list vehicle cache keys", zap.Error(err), zap.String("pattern", vehiclePattern))
		} else if len(keys) > 0 {
			if err := s.redis.Del(ctx, keys...).Err(); err != nil {
				s.logger.Warn("Failed to delete vehicle cache keys", zap.Error(err), zap.String("pattern", vehiclePattern))
			}
		}

		// 4. Analytics vehicle cache: analytics:{tenantID}:vehicle:{id}:*
		analyticsVehPattern := fmt.Sprintf("analytics:%s:vehicle:%s:*", tenantID.String(), v.ID.String())
		keys, err = s.redis.Keys(ctx, analyticsVehPattern).Result()
		if err != nil {
			s.logger.Warn("Failed to list analytics vehicle cache keys", zap.Error(err), zap.String("pattern", analyticsVehPattern))
		} else if len(keys) > 0 {
			if err := s.redis.Del(ctx, keys...).Err(); err != nil {
				s.logger.Warn("Failed to delete analytics vehicle cache", zap.Error(err), zap.String("pattern", analyticsVehPattern))
			}
		}
	}

	// 5. Analytics driver cache: analytics:{tenantID}:driver:{userID}:*
	analyticsDrvPattern := fmt.Sprintf("analytics:%s:driver:%s:*", tenantID.String(), userID.String())
	keys, err := s.redis.Keys(ctx, analyticsDrvPattern).Result()
	if err != nil {
		s.logger.Warn("Failed to list analytics driver cache keys", zap.Error(err), zap.String("pattern", analyticsDrvPattern))
	} else if len(keys) > 0 {
		if err := s.redis.Del(ctx, keys...).Err(); err != nil {
			s.logger.Warn("Failed to delete analytics driver cache", zap.Error(err), zap.String("pattern", analyticsDrvPattern))
		}
	}

	// 6. Driver debounce/cache: driver:{userID}
	driverKey := fmt.Sprintf("driver:%s", userID.String())
	if err := s.redis.Del(ctx, driverKey).Err(); err != nil {
		s.logger.Warn("Failed to delete driver cache key", zap.Error(err), zap.String("key", driverKey))
	}
}
