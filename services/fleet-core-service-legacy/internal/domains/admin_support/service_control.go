package admin_support

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *service) SetSystemFlag(ctx context.Context, key, value string, actorID uuid.UUID) error {
	flag := &SystemFlag{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}

	if err := s.repo.SetSystemFlag(ctx, flag); err != nil {
		s.logger.Error("failed to set system flag", zap.String("key", key), zap.Error(err))
		return err
	}

	s.auditLog(ctx, actorID, "system.flag_change", "system_flag", key, nil, value)
	s.logger.Info("system flag updated", zap.String("key", key), zap.String("value", value), zap.String("actor_id", actorID.String()))
	return nil
}

func (s *service) StartMaintenance(ctx context.Context, serviceName string, duration time.Duration, description string) (*MaintenanceWindow, error) {
	window := &MaintenanceWindow{
		ID:          uuid.New(),
		Service:     serviceName,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(duration),
		Description: description,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateMaintenanceWindow(ctx, window); err != nil {
		s.logger.Error("failed to start maintenance window", zap.Error(err))
		return nil, err
	}

	s.logger.Warn("MAINTENANCE STARTED",
		zap.String("service", serviceName),
		zap.Time("end_time", window.EndTime),
		zap.String("description", description),
	)

	return window, nil
}

func (s *service) ListMaintenanceWindows(ctx context.Context) ([]*MaintenanceWindow, error) {
	return s.repo.ListMaintenanceWindows(ctx)
}

func (s *service) GetSystemFlag(ctx context.Context, key string) (*SystemFlag, error) {
	return s.repo.GetSystemFlag(ctx, key)
}

func (s *service) ListSystemFlags(ctx context.Context) ([]*SystemFlag, error) {
	return s.repo.ListSystemFlags(ctx)
}

// ─── Disaster Recovery & Two-Person Approvals (Phase 3) ──────────────────────

func (s *service) RequestFailover(ctx context.Context, region string, requestedBy uuid.UUID) (*AdminOperationApproval, error) {
	payload := []byte(`{"target_region": "` + region + `"}`)
	app := &AdminOperationApproval{
		ID:            uuid.New(),
		OperationType: "dr.failover",
		Payload:       payload,
		RequestedBy:   requestedBy,
		Status:        ApprovalPending,
		CreatedAt:     time.Time{}, // handled by DB default, but set here if needed: time.Now()
	}
	app.CreatedAt = time.Now()

	if err := s.repo.CreateAdminOperationApproval(ctx, app); err != nil {
		s.logger.Error("failed to create failover approval request", zap.Error(err))
		return nil, err
	}

	s.auditLog(ctx, requestedBy, "approval.request", "admin_operation_approval", app.ID.String(), nil, app.Status)
	s.logger.Info("FAILOVER REQUESTED", zap.String("approval_id", app.ID.String()), zap.String("requested_by", requestedBy.String()))
	return app, nil
}

func (s *service) ApproveFailover(ctx context.Context, approvalID uuid.UUID, approvedBy uuid.UUID) error {
	app, err := s.repo.GetAdminOperationApproval(ctx, approvalID)
	if err != nil {
		return err
	}

	if app.Status != ApprovalPending {
		return context.DeadlineExceeded // Or a custom error like "not pending"
	}

	if app.RequestedBy == approvedBy {
		return context.DeadlineExceeded // Prevent self-approval (using a dummy error for now, better to have a custom `ErrSelfApproval`)
	}

	now := time.Now()
	app.Status = ApprovalApproved
	app.ApprovedBy = &approvedBy
	app.ResolvedAt = &now

	if err := s.repo.UpdateAdminOperationApproval(ctx, app); err != nil {
		s.logger.Error("failed to approve failover", zap.Error(err))
		return err
	}

	s.auditLog(ctx, approvedBy, "approval.approve", "admin_operation_approval", app.ID.String(), ApprovalPending, app.Status)
	s.logger.Info("FAILOVER APPROVED & EXECUTING", zap.String("approval_id", app.ID.String()), zap.String("approved_by", approvedBy.String()))
	
	// Stub: Actual Failover Execution Logic
	s.auditLog(ctx, approvedBy, "dr.failover.execute", "system", "global", nil, "EXECUTED")
	
	// Mark as executed
	app.Status = ApprovalExecuted
	s.repo.UpdateAdminOperationApproval(ctx, app)

	return nil
}

func (s *service) RejectFailover(ctx context.Context, approvalID uuid.UUID, rejectedBy uuid.UUID) error {
	app, err := s.repo.GetAdminOperationApproval(ctx, approvalID)
	if err != nil {
		return err
	}

	if app.Status != ApprovalPending {
		return context.DeadlineExceeded // Or custom error
	}

	now := time.Now()
	app.Status = ApprovalRejected
	app.ApprovedBy = &rejectedBy // Using ApprovedBy field to track who resolved it
	app.ResolvedAt = &now

	if err := s.repo.UpdateAdminOperationApproval(ctx, app); err != nil {
		s.logger.Error("failed to reject failover", zap.Error(err))
		return err
	}

	s.auditLog(ctx, rejectedBy, "approval.reject", "admin_operation_approval", app.ID.String(), ApprovalPending, app.Status)
	s.logger.Info("FAILOVER REJECTED", zap.String("approval_id", app.ID.String()), zap.String("rejected_by", rejectedBy.String()))
	return nil
}

func (s *service) ListPendingApprovals(ctx context.Context) ([]*AdminOperationApproval, error) {
	return s.repo.ListPendingApprovals(ctx)
}

func (s *service) ExecuteFailover(ctx context.Context, region string, actorID uuid.UUID) error {
	// Deprecated: replaced by RequestFailover/ApproveFailover
	return nil
}

func (s *service) TriggerBackup(ctx context.Context, actorID uuid.UUID) error {
	s.logger.Info("BACKUP TRIGGERED (STUB)", zap.String("actor_id", actorID.String()))
	return nil
}
