package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/governance/metrics"
	"github.com/flyfleet/fleet-core/internal/domains/governance/models"
	"github.com/flyfleet/fleet-core/internal/domains/governance/repository"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
)

var (
	ErrRequestNotFound = errors.New("governance request not found")
	ErrInvalidStatus   = errors.New("invalid request status transition")
)

type GovernanceService interface {
	CreateRequest(ctx context.Context, tenantID, userID uuid.UUID, req models.CreateGovernanceRequest) (*models.GovernanceRequest, error)
	GetRequest(ctx context.Context, id, tenantID uuid.UUID) (*models.GovernanceRequest, error)
	ListRequests(ctx context.Context, filter models.GovernanceRequestFilter) (*models.GovernanceRequestListResponse, error)
	ApproveRequest(ctx context.Context, id, tenantID uuid.UUID, notes string) error
	VerifyRequest(ctx context.Context, id, tenantID, userID uuid.UUID, status string) error
	RejectRequest(ctx context.Context, id, tenantID uuid.UUID, notes string) error

	GetSettings(ctx context.Context, tenantID uuid.UUID) (*models.GovernanceSettings, error)
	UpdateSettings(ctx context.Context, tenantID, userID uuid.UUID, settings *models.GovernanceSettings) error

	ExpireRequests(ctx context.Context) (int, error)
	EnforceRetention(ctx context.Context) error
}

type governanceService struct {
	repo   repository.GovernanceRepository
	nats   *nats.Conn
	logger *zap.Logger
}

func NewGovernanceService(repo repository.GovernanceRepository, nats *nats.Conn, logger *zap.Logger) GovernanceService {
	return &governanceService{repo: repo, nats: nats, logger: logger}
}

func (s *governanceService) CreateRequest(ctx context.Context, tenantID, userID uuid.UUID, req models.CreateGovernanceRequest) (*models.GovernanceRequest, error) {
	_, span := otel.Tracer("governance").Start(ctx, "create_governance_request")
	defer span.End()

	startTime := time.Now()
	defer func() {
		metrics.GovernanceRequestProcessingSeconds.WithLabelValues(string(req.RequestType)).Observe(time.Since(startTime).Seconds())
	}()

	gReq := &models.GovernanceRequest{
		ID:             uuid.New(),
		OrganizationID: tenantID,
		RequesterID:    userID,
		RequestType:    req.RequestType,
		DataScope:      req.DataScope,
		Purpose:        req.Purpose,
		Format:         req.Format,
		DeliveryMethod: req.DeliveryMethod,
		Status:         models.RequestStatusPending,
		CreatedBy:      userID,
		UpdatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateRequest(ctx, gReq); err != nil {
		s.logger.Error("failed to create governance request", zap.Error(err))
		return nil, err
	}

	// Audit Log
	s.repo.CreateAuditLog(ctx, &models.GovernanceAuditLog{
		ID:             uuid.New(),
		OrganizationID: tenantID,
		UserID:         userID,
		Action:         "CREATE_REQUEST",
		Module:         "DATA_REQUESTS",
		CreatedAt:      time.Now(),
	})

	s.publishEvent("governance.request.created", gReq)
	metrics.GovernanceRequestsTotal.WithLabelValues(string(gReq.RequestType), "PENDING").Inc()
	metrics.GovernancePendingRequests.Inc()

	return gReq, nil
}

func (s *governanceService) GetRequest(ctx context.Context, id, tenantID uuid.UUID) (*models.GovernanceRequest, error) {
	_, span := otel.Tracer("governance").Start(ctx, "get_governance_request")
	defer span.End()
	req, err := s.repo.GetRequestByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, ErrRequestNotFound
	}
	return req, nil
}

func (s *governanceService) ListRequests(ctx context.Context, filter models.GovernanceRequestFilter) (*models.GovernanceRequestListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	return s.repo.ListRequests(ctx, filter)
}

func (s *governanceService) ApproveRequest(ctx context.Context, id, tenantID uuid.UUID, notes string) error {
	ctx, span := otel.Tracer("governance").Start(ctx, "approve_governance_request")
	defer span.End()

	req, err := s.GetRequest(ctx, id, tenantID)
	if err != nil {
		return err
	}

	if req.Status != models.RequestStatusPending {
		return ErrInvalidStatus
	}

	if req.VerificationStatus != "VERIFIED" {
		return errors.New("only verified requests can be approved")
	}

	// Set expiry date based on settings (default 48h)
	settings, err := s.repo.GetSettings(ctx, tenantID)
	if err != nil {
		s.logger.Warn("failed to get governance settings, using default TTL", zap.Error(err))
	}
	ttl := 48
	if settings != nil {
		ttl = settings.DownloadTTLHours
	}
	expiry := time.Now().Add(time.Duration(ttl) * time.Hour)

	s.logger.Info("approving request", zap.String("id", id.String()), zap.Time("expiry", expiry))

	// Update status with expiry date
	if err := s.repo.UpdateRequestStatus(ctx, id, tenantID, models.RequestStatusApproved, &notes, &expiry); err != nil {
		return err
	}

	// Audit Log
	userID := s.getUserID(ctx)
	s.repo.CreateAuditLog(ctx, &models.GovernanceAuditLog{
		ID:             uuid.New(),
		OrganizationID: tenantID,
		UserID:         userID,
		Action:         "APPROVE_REQUEST",
		Module:         "DATA_REQUESTS",
		ChangeDetail:   &notes,
		CreatedAt:      time.Now(),
	})

	s.publishEvent("governance.request.approved", req)
	metrics.GovernanceRequestsTotal.WithLabelValues(string(req.RequestType), "APPROVED").Inc()
	metrics.GovernancePendingRequests.Dec()

	return nil
}

func (s *governanceService) VerifyRequest(ctx context.Context, id, tenantID, userID uuid.UUID, status string) error {
	ctx, span := otel.Tracer("governance").Start(ctx, "verify_governance_request")
	defer span.End()

	req, err := s.GetRequest(ctx, id, tenantID)
	if err != nil {
		return err
	}

	if req.Status != models.RequestStatusPending {
		return ErrInvalidStatus
	}

	if err := s.repo.UpdateVerificationStatus(ctx, id, tenantID, status, userID); err != nil {
		return err
	}

	// Audit Log
	s.repo.CreateAuditLog(ctx, &models.GovernanceAuditLog{
		ID:             uuid.New(),
		OrganizationID: tenantID,
		UserID:         userID,
		Action:         "VERIFY_REQUEST",
		Module:         "DATA_REQUESTS",
		ChangeDetail:   &status,
		CreatedAt:      time.Now(),
	})

	req.VerificationStatus = status
	s.publishEvent("governance.request.verified", req)

	return nil
}

func (s *governanceService) RejectRequest(ctx context.Context, id, tenantID uuid.UUID, notes string) error {
	ctx, span := otel.Tracer("governance").Start(ctx, "reject_governance_request")
	defer span.End()

	req, err := s.GetRequest(ctx, id, tenantID)
	if err != nil {
		return err
	}

	if req.Status != models.RequestStatusPending {
		return ErrInvalidStatus
	}

	// Update status (no expiry needed for rejected requests)
	if err := s.repo.UpdateRequestStatus(ctx, id, tenantID, models.RequestStatusRejected, &notes, nil); err != nil {
		return err
	}

	// Audit Log
	userID := s.getUserID(ctx)
	s.repo.CreateAuditLog(ctx, &models.GovernanceAuditLog{
		ID:             uuid.New(),
		OrganizationID: tenantID,
		UserID:         userID,
		Action:         "REJECT_REQUEST",
		Module:         "DATA_REQUESTS",
		ChangeDetail:   &notes,
		CreatedAt:      time.Now(),
	})

	metrics.GovernanceRequestsTotal.WithLabelValues(string(req.RequestType), "REJECTED").Inc()
	metrics.GovernancePendingRequests.Dec()

	return nil
}

func (s *governanceService) GetSettings(ctx context.Context, tenantID uuid.UUID) (*models.GovernanceSettings, error) {
	return s.repo.GetSettings(ctx, tenantID)
}

func (s *governanceService) UpdateSettings(ctx context.Context, tenantID, userID uuid.UUID, settings *models.GovernanceSettings) error {
	settings.OrganizationID = tenantID
	settings.CreatedBy = userID // Fallback if record is new
	settings.UpdatedBy = userID

	if err := s.repo.UpsertSettings(ctx, settings); err != nil {
		return err
	}

	// Audit Log
	s.repo.CreateAuditLog(ctx, &models.GovernanceAuditLog{
		ID:             uuid.New(),
		OrganizationID: tenantID,
		UserID:         userID,
		Action:         "UPDATE_SETTINGS",
		Module:         "GOVERNANCE_SETTINGS",
		CreatedAt:      time.Now(),
	})

	return nil
}

func (s *governanceService) ExpireRequests(ctx context.Context) (int, error) {
	_, span := otel.Tracer("governance").Start(ctx, "expire_governance_requests")
	defer span.End()

	expired, err := s.repo.ExpireRequests(ctx)
	if err != nil {
		s.logger.Error("failed to expire requests", zap.Error(err))
		return 0, err
	}

	for _, info := range expired {
		s.logger.Info("request expired", zap.String("id", info.ID.String()))
		s.repo.CreateAuditLog(ctx, &models.GovernanceAuditLog{
			ID:             uuid.New(),
			OrganizationID: info.OrganizationID,
			UserID:         uuid.Nil, // System action
			Action:         "EXPIRE_REQUEST",
			Module:         "DATA_REQUESTS",
			CreatedAt:      time.Now(),
		})
		s.publishEvent("governance.request.expired", info)
		metrics.GovernanceRequestsTotal.WithLabelValues("UNKNOWN", "EXPIRED").Inc()
	}

	return len(expired), nil
}

func (s *governanceService) getUserID(ctx context.Context) uuid.UUID {
	if id, ok := ctxkeys.GetUserID(ctx); ok {
		return id
	}
	return uuid.Nil
}

func (s *governanceService) publishEvent(subject string, data interface{}) {
	if s.nats == nil {
		return
	}
	payload, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("failed to marshal nats event", zap.Error(err))
		return
	}
	if err := s.nats.Publish(subject, payload); err != nil {
		s.logger.Error("failed to publish nats event", zap.String("subject", subject), zap.Error(err))
	}
}

func (s *governanceService) EnforceRetention(ctx context.Context) error {
	settings, err := s.repo.ListAllSettings(ctx)
	if err != nil {
		return fmt.Errorf("failed to list all governance settings: %w", err)
	}

	for _, setting := range settings {
		var policies map[string]int
		if err := json.Unmarshal(setting.RetentionPolicies, &policies); err != nil {
			s.logger.Error("failed to unmarshal retention policies", zap.String("tenantID", setting.OrganizationID.String()), zap.Error(err))
			continue
		}

		for table, days := range policies {
			count, err := s.repo.EnforceRetentionPolicy(ctx, setting.OrganizationID, table, days)
			if err != nil {
				s.logger.Error("failed to enforce retention policy",
					zap.String("tenantID", setting.OrganizationID.String()),
					zap.String("table", table),
					zap.Error(err))
				continue
			}
			if count > 0 {
				s.logger.Info("enforced retention policy",
					zap.String("tenantID", setting.OrganizationID.String()),
					zap.String("table", table),
					zap.Int64("deleted_rows", count))
				metrics.GovernanceDeletionsTotal.Add(float64(count))
			}
		}
	}

	return nil
}
