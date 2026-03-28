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

// VahanIntegrationService defines the interface for VAHAN vehicle verification integration
type VahanIntegrationService interface {
	VerifyVehicle(ctx context.Context, orgID uuid.UUID, vehicleID uuid.UUID, regNumber string) (*models.VehicleVerification, error)
	BulkVerifyVehicle(ctx context.Context, orgID uuid.UUID, requests []models.VerifyVehicleRequest) ([]*models.VehicleVerification, error)
}

type VahanIntegrationServiceImpl struct {
	repo    repository.IndiaIntegrationRepository
	redis   *redis.Client
	logger  *zap.Logger
	limiter *rate.Limiter
}

func NewVahanIntegrationService(repo repository.IndiaIntegrationRepository, redis *redis.Client, logger *zap.Logger) *VahanIntegrationServiceImpl {
	return &VahanIntegrationServiceImpl{
		repo:    repo,
		redis:   redis,
		logger:  logger,
		limiter: rate.NewLimiter(rate.Every(time.Minute/50), 1),
	}
}

func (s *VahanIntegrationServiceImpl) VerifyVehicle(ctx context.Context, orgID uuid.UUID, vehicleID uuid.UUID, regNumber string) (*models.VehicleVerification, error) {
	// Try Redis cache first
	if s.redis != nil {
		// SECURITY FIX: Include orgID in cache key to prevent cross-tenant cache collision
		cacheKey := fmt.Sprintf("vahan:verify:%s:%s", orgID.String(), regNumber)
		val, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var cached models.VehicleVerification
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				s.logger.Info("vahan cache hit", zap.String("reg_number", regNumber), zap.String("org_id", orgID.String()))
				// SECURITY FIX: Remove PII fields from cached response to avoid exposure in Redis
				cached.OwnerName = ""
				cached.RCStatus = ""
				return &cached, nil
			}
		}
	}

	if err := s.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}
	s.logger.Info("verifying vehicle with VAHAN", zap.String("reg_number", regNumber))

	// Mock VAHAN API Call
	verification := &models.VehicleVerification{
		ID:                 uuid.New().String(),
		VehicleID:          vehicleID,
		OrganizationID:     orgID,
		RegistrationNumber: regNumber,
		Status:             "verified",
		VerificationDate:   time.Now(),
		OwnerName:          "Mock Owner Name",
		RCStatus:           "active",
		InsuranceExpiry:    time.Now().Add(30 * 24 * time.Hour),
		PUCExpiry:          time.Now().Add(15 * 24 * time.Hour),
		CreatedAt:          time.Now(),
	}

	log := &models.IntegrationLog{
		ServiceType:     "vahan",
		RequestPayload:  map[string]string{"reg_number": regNumber},
		ResponsePayload: verification,
		Status:          200,
		OrganizationID:  &orgID,
	}
	_ = s.repo.CreateIntegrationLog(ctx, log)

	if err := s.repo.CreateVehicleVerification(ctx, verification); err != nil {
		return nil, err
	}

	// Store in Redis cache with 24h TTL
	if s.redis != nil {
		// SECURITY FIX: Include orgID in cache key to prevent cross-tenant cache collision
		cacheKey := fmt.Sprintf("vahan:verify:%s:%s", orgID.String(), regNumber)
		// SECURITY FIX: Remove PII before caching to prevent exposure if Redis is compromised
		verification.OwnerName = ""
		verification.RCStatus = ""
		if data, err := json.Marshal(verification); err == nil {
			s.redis.Set(ctx, cacheKey, data, 24*time.Hour)
		}

		// Also set compliance flag if insurance is expired
		if verification.InsuranceExpiry.Before(time.Now()) {
			complianceKey := fmt.Sprintf("compliance:expired_insurance:vehicle:%s", vehicleID.String())
			s.redis.Set(ctx, complianceKey, "true", 0)
		} else {
			// Ensure it's cleared if now renewed
			complianceKey := fmt.Sprintf("compliance:expired_insurance:vehicle:%s", vehicleID.String())
			s.redis.Del(ctx, complianceKey)
		}
	}

	return verification, nil
}

func (s *VahanIntegrationServiceImpl) BulkVerifyVehicle(ctx context.Context, orgID uuid.UUID, requests []models.VerifyVehicleRequest) ([]*models.VehicleVerification, error) {
	if len(requests) > 100 {
		return nil, fmt.Errorf("batch size exceeds maximum of 100")
	}

	results := make([]*models.VehicleVerification, 0, len(requests))
	for _, req := range requests {
		// In a real scenario, we might want to do this in parallel or use a bulk API if available.
		// For now, we'll iterate and collect results, then bulk save.
		v, err := s.VerifyVehicle(ctx, orgID, req.VehicleID, req.RegistrationNumber)
		if err != nil {
			s.logger.Error("failed to verify vehicle in bulk", zap.String("reg_number", req.RegistrationNumber), zap.Error(err))
			continue
		}
		results = append(results, v)
	}

	// The individual VerifyVehicle already saves to DB, but if we wanted to true bulk save:
	// err := s.repo.BulkCreateVehicleVerification(ctx, results)

	return results, nil
}
