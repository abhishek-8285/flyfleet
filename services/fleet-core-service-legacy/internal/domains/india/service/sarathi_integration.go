package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/india/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/india/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// SarathiIntegrationService defines the interface for SARATHI driver license verification integration
type SarathiIntegrationService interface {
	VerifyDriver(ctx context.Context, orgID uuid.UUID, driverID uuid.UUID, licenseNumber string, dob string) (*models.DriverVerification, error)
	BulkVerifyDriver(ctx context.Context, orgID uuid.UUID, requests []models.VerifyDriverRequest) ([]*models.DriverVerification, error)
}

type SarathiIntegrationServiceImpl struct {
	repo    repository.IndiaIntegrationRepository
	redis   *redis.Client
	logger  *zap.Logger
	limiter *rate.Limiter
}

func NewSarathiIntegrationService(repo repository.IndiaIntegrationRepository, redis *redis.Client, logger *zap.Logger) *SarathiIntegrationServiceImpl {
	return &SarathiIntegrationServiceImpl{
		repo:    repo,
		redis:   redis,
		logger:  logger,
		limiter: rate.NewLimiter(rate.Every(time.Minute/50), 1),
	}
}

func (s *SarathiIntegrationServiceImpl) VerifyDriver(ctx context.Context, orgID uuid.UUID, driverID uuid.UUID, licenseNumber string, dob string) (*models.DriverVerification, error) {
	// Try Redis cache first
	if s.redis != nil {
		cacheKey := fmt.Sprintf("sarathi:verify:%s", licenseNumber)
		val, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var cached models.DriverVerification
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				s.logger.Info("sarathi cache hit", zap.String("license_number", licenseNumber))
				return &cached, nil
			}
		}
	}

	if err := s.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}
	s.logger.Info("verifying driver with SARATHI", zap.String("license_number", licenseNumber))

	// Mock SARATHI API Call
	parsedDOB, _ := time.Parse("2006-01-02", dob)
	verification := &models.DriverVerification{
		ID:              uuid.New().String(),
		DriverID:        driverID,
		OrganizationID:  orgID,
		LicenseNumber:   licenseNumber,
		Status:          "verified",
		VerificationDate: time.Now(),
		Name:            "Mock Driver Name",
		DateOfBirth:     parsedDOB,
		ValidityTo:      time.Now().Add(5 * 365 * 24 * time.Hour),
		CreatedAt:       time.Now(),
	}

	log := &models.IntegrationLog{
		ServiceType:     "sarathi",
		RequestPayload:  map[string]string{"license_number": licenseNumber, "dob": dob},
		ResponsePayload: verification,
		Status:          200,
		OrganizationID:  &orgID,
	}
	_ = s.repo.CreateIntegrationLog(ctx, log)

	if err := s.repo.CreateDriverVerification(ctx, verification); err != nil {
		return nil, err
	}

	// Store in Redis cache with 24h TTL
	if s.redis != nil {
		cacheKey := fmt.Sprintf("sarathi:verify:%s", licenseNumber)
		if data, err := json.Marshal(verification); err == nil {
			s.redis.Set(ctx, cacheKey, data, 24*time.Hour)
		}
	}

	return verification, nil
}

func (s *SarathiIntegrationServiceImpl) BulkVerifyDriver(ctx context.Context, orgID uuid.UUID, requests []models.VerifyDriverRequest) ([]*models.DriverVerification, error) {
	if len(requests) > 100 {
		return nil, fmt.Errorf("batch size exceeds maximum of 100")
	}

	results := make([]*models.DriverVerification, 0, len(requests))
	for _, req := range requests {
		// In a real scenario, we might want to do this in parallel or use a bulk API if available.
		// For now, we'll iterate and collect results, then bulk save.
		v, err := s.VerifyDriver(ctx, orgID, req.DriverID, req.LicenseNumber, req.DateOfBirth)
		if err != nil {
			s.logger.Error("failed to verify driver in bulk", zap.String("license_number", req.LicenseNumber), zap.Error(err))
			continue
		}
		results = append(results, v)
	}

	return results, nil
}
