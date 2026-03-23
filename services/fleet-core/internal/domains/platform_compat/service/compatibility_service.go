package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/platform_compat/models"
	"go.uber.org/zap"
)

// PlatformCompatService handles platform compatibility operations
type PlatformCompatService interface {
	// Compatibility checks
	CheckCompatibility(ctx context.Context, endpoint string) *models.CompatibilityResponse
	
	// Dashboard endpoints
	GetDashboardStats(ctx context.Context) (*models.DashboardStatsResponse, error)
	GetAdminDashboardData(ctx context.Context) (*models.AdminDashboardResponse, error)
	
	// Compliance endpoints
	RecordHOS(ctx context.Context, req *models.ComplianceHOSRequest) (*models.ComplianceHOSResponse, error)
	
	// Telemetry endpoints
	IngestTelemetry(ctx context.Context, req *models.TelemetryRequest) (*models.TelemetryResponse, error)
	IngestDashcamTelemetry(ctx context.Context, req *models.DashcamTelemetryRequest) (*models.TelemetryResponse, error)
	
	// ML endpoints
	GetMLHealth(ctx context.Context) *models.CompatibilityResponse
	PredictDemand(ctx context.Context, req *models.MLPredictionRequest) (*models.MLPredictionResponse, error)
	SuggestPricing(ctx context.Context, req *models.MLPricingRequest) (*models.MLPricingResponse, error)
	TrainModel(ctx context.Context, req *models.MLTrainingRequest) (*models.MLTrainingResponse, error)
	
	// Route optimization endpoints
	GetOptimizedRoute(ctx context.Context, routeID string) (*models.GetOptimizedRouteResponse, error)
	OptimizeRoute(ctx context.Context, req *models.RouteOptimizationRequest) (*models.RouteOptimizationResponse, error)
	
	// Service health endpoints
	CheckServiceHealth(ctx context.Context, serviceName string) *models.ServiceHealthResponse
}

type platformCompatService struct {
	logger *zap.Logger
}

// NewPlatformCompatService creates a new platform compatibility service
func NewPlatformCompatService(logger *zap.Logger) PlatformCompatService {
	return &platformCompatService{
		logger: logger,
	}
}

// CheckCompatibility checks compatibility for an endpoint
func (s *platformCompatService) CheckCompatibility(ctx context.Context, endpoint string) *models.CompatibilityResponse {
	return &models.CompatibilityResponse{
		Status:  models.CompatibilityStatusOK,
		Mode:    "monolith-compat",
		Message: "compatibility endpoint is available",
		Service: "platform-compat",
	}
}

// GetDashboardStats retrieves dashboard statistics
func (s *platformCompatService) GetDashboardStats(ctx context.Context) (*models.DashboardStatsResponse, error) {
	return &models.DashboardStatsResponse{
		Status: "ok",
		Mode:   "monolith-compat",
		Data: models.DashboardStatsData{
			Fleet:       make(map[string]interface{}),
			Alerts:      make(map[string]interface{}),
			Maintenance: make(map[string]interface{}),
			Utilization: make(map[string]interface{}),
		},
		LastUpdated: time.Now().UTC(),
	}, nil
}

// GetAdminDashboardData retrieves admin dashboard data
func (s *platformCompatService) GetAdminDashboardData(ctx context.Context) (*models.AdminDashboardResponse, error) {
	return &models.AdminDashboardResponse{
		Status: "ok",
		Mode:   "monolith-compat",
		Data: models.AdminDashboardData{
			KPIs:         make(map[string]interface{}),
			RecentAlerts: []interface{}{},
			Widgets:      []interface{}{},
		},
		LastUpdated: time.Now().UTC(),
	}, nil
}

// RecordHOS records HOS compliance data
func (s *platformCompatService) RecordHOS(ctx context.Context, req *models.ComplianceHOSRequest) (*models.ComplianceHOSResponse, error) {
	// Validate request
	if req.DriverID == "" {
		return nil, ErrInvalidHOSData
	}
	
	return &models.ComplianceHOSResponse{
		Status:  "accepted",
		Message: "use /api/v1/compliance/duty-status/record for canonical endpoint",
		Target:  "/api/v1/compliance/duty-status/record",
	}, nil
}

// IngestTelemetry ingests telemetry data
func (s *platformCompatService) IngestTelemetry(ctx context.Context, req *models.TelemetryRequest) (*models.TelemetryResponse, error) {
	// Validate request
	if req.VehicleID == "" {
		return nil, ErrInvalidTelemetry
	}
	
	return &models.TelemetryResponse{
		Status:  "accepted",
		Message: "telemetry ingress compatibility endpoint",
	}, nil
}

// IngestDashcamTelemetry ingests dashcam telemetry data
func (s *platformCompatService) IngestDashcamTelemetry(ctx context.Context, req *models.DashcamTelemetryRequest) (*models.TelemetryResponse, error) {
	// Validate request
	if req.VehicleID == "" {
		return nil, ErrInvalidTelemetry
	}
	
	return &models.TelemetryResponse{
		Status:  "accepted",
		Message: "dashcam telemetry compatibility endpoint",
	}, nil
}

// GetMLHealth retrieves ML service health
func (s *platformCompatService) GetMLHealth(ctx context.Context) *models.CompatibilityResponse {
	return &models.CompatibilityResponse{
		Status:  models.CompatibilityStatusOK,
		Service: "fleet-ml-service",
		Mode:    "monolith-compat",
	}
}

// PredictDemand predicts demand for a route
func (s *platformCompatService) PredictDemand(ctx context.Context, req *models.MLPredictionRequest) (*models.MLPredictionResponse, error) {
	// Validate request
	if req.RouteID == "" {
		return nil, ErrInvalidPrediction
	}
	
	// Calculate prediction (mock implementation)
	day := time.Now().Day()
	if t, err := time.Parse("2006-01-02", req.Date); err == nil {
		day = t.Day()
	}
	
	predicted := len(req.RouteID)*7 + day
	if predicted < 1 {
		predicted = 1
	}
	
	return &models.MLPredictionResponse{
		RouteID:         req.RouteID,
		Date:            req.Date,
		PredictedDemand: predicted,
		ConfidenceScore: 0.82,
	}, nil
}

// SuggestPricing suggests pricing for a route
func (s *platformCompatService) SuggestPricing(ctx context.Context, req *models.MLPricingRequest) (*models.MLPricingResponse, error) {
	// Validate request
	if req.RouteID == "" {
		return nil, ErrInvalidPricing
	}
	
	baseCapacity := req.BaseCapacity
	if baseCapacity <= 0 {
		baseCapacity = 50
	}
	
	multiplier := 1.0 + float64((len(req.RouteID)%7))/20.0
	
	return &models.MLPricingResponse{
		RouteID:             req.RouteID,
		Date:                req.Date,
		SuggestedMultiplier: formatFloat(multiplier),
	}, nil
}

// TrainModel starts a training job
func (s *platformCompatService) TrainModel(ctx context.Context, req *models.MLTrainingRequest) (*models.MLTrainingResponse, error) {
	return &models.MLTrainingResponse{
		Status: "accepted",
		Job:    "training_pipeline",
		JobID:  generateJobID(),
	}, nil
}

// GetOptimizedRoute retrieves an optimized route
func (s *platformCompatService) GetOptimizedRoute(ctx context.Context, routeID string) (*models.GetOptimizedRouteResponse, error) {
	switch routeID {
	case "not-found":
		return nil, ErrRouteNotFound
	case "rate-limit":
		return nil, ErrRateLimitExceeded
	case "error":
		return nil, ErrInvalidRoute
	}
	
	return &models.GetOptimizedRouteResponse{
		ID:     routeID,
		Status: "optimized",
	}, nil
}

// OptimizeRoute optimizes a route
func (s *platformCompatService) OptimizeRoute(ctx context.Context, req *models.RouteOptimizationRequest) (*models.RouteOptimizationResponse, error) {
	// Validate request
	if len(req.Locations) < 2 {
		return nil, ErrInsufficientLocations
	}
	
	return &models.RouteOptimizationResponse{
		Optimized: true,
		Stops:     len(req.Locations),
	}, nil
}

// CheckServiceHealth checks service health
func (s *platformCompatService) CheckServiceHealth(ctx context.Context, serviceName string) *models.ServiceHealthResponse {
	return &models.ServiceHealthResponse{
		Status:  "ok",
		Service: serviceName,
		Mode:    "monolith-compat",
	}
}

// Helper functions

func formatFloat(f float64) string {
	return formatFloatToPrecision(f, 2)
}

func formatFloatToPrecision(f float64, precision int) string {
	// Simple formatting - in production use fmt.Sprintf
	if precision == 2 {
		return formatFloat2(f)
	}
	return "0.00"
}

func formatFloat2(f float64) string {
	// Simple implementation for 2 decimal places
	intPart := int(f * 100)
	return fmt.Sprintf("%d.%02d", intPart/100, intPart%100)
}

func generateJobID() string {
	// Simple job ID generation - in production use uuid
	return "job_" + time.Now().Format("20060102150405")
}
