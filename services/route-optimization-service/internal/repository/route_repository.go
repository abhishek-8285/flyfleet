package repository

import (
	"context"
	"time"

	"github.com/flyfleet/route-optimization-service/internal/models"
	"github.com/google/uuid"
)

// RouteOptimizationRepository defines the interface for route optimization data access
type RouteOptimizationRepository interface {
	CreateOptimizationRequest(ctx context.Context, request *models.RouteOptimizationRequest) error
	GetOptimizationRequest(ctx context.Context, id string) (*models.RouteOptimizationRequest, error)
	UpdateOptimizationRequest(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteOptimizationRequest(ctx context.Context, id string) error
	ListOptimizationRequests(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.RouteOptimizationRequest, int, error)
	
	CreateOptimizationResult(ctx context.Context, result *models.RouteOptimizationResult) error
	GetOptimizationResult(ctx context.Context, id string) (*models.RouteOptimizationResult, error)
	GetOptimizationResultByRequestID(ctx context.Context, requestID string) (*models.RouteOptimizationResult, error)
	UpdateOptimizationResult(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteOptimizationResult(ctx context.Context, id string) error
	ListOptimizationResults(ctx context.Context, requestID string, limit, offset int) ([]*models.RouteOptimizationResult, int, error)
	
	CreateOptimizationJob(ctx context.Context, job *models.OptimizationJob) error
	GetOptimizationJob(ctx context.Context, id string) (*models.OptimizationJob, error)
	UpdateOptimizationJob(ctx context.Context, id string, updates map[string]interface{}) error
	UpdateOptimizationJobStatus(ctx context.Context, id, status string, completedAt *time.Time, error *string) error
	ListOptimizationJobs(ctx context.Context, requestID string, status string, limit, offset int) ([]*models.OptimizationJob, int, error)
	
	HealthCheck(ctx context.Context) error
	Ping(ctx context.Context) error
}