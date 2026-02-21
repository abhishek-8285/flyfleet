package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/vehicle-service/internal/models"
	"github.com/flyfleet/vehicle-service/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TripService struct {
	repo   *repository.TripRepository
	logger *zap.Logger
}

func NewTripService(repo *repository.TripRepository, logger *zap.Logger) *TripService {
	return &TripService{repo: repo, logger: logger}
}

func (s *TripService) generateTripID() string {
	year := time.Now().Year()
	seq := time.Now().UnixNano() % 10000
	return fmt.Sprintf("TR-%d-%04d", year, seq)
}

func (s *TripService) CreateTrip(ctx context.Context, orgID uuid.UUID, req *models.CreateTripRequest) (*models.Trip, error) {
	userID := s.getUserID(ctx)
	trip := &models.Trip{
		ID:             uuid.New(),
		TripID:         s.generateTripID(),
		VehicleID:      req.VehicleID,
		DriverID:       req.DriverID,
		StartLocation:  req.StartLocation,
		EndLocation:    req.EndLocation,
		StartTime:      req.StartTime,
		Status:         "scheduled",
		OrganizationID: orgID,
		CreatedBy:      userID,
		UpdatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, trip); err != nil {
		s.logger.Error("Failed to create trip", zap.Error(err))
		return nil, err
	}

	return trip, nil
}

func (s *TripService) getUserID(ctx context.Context) uuid.UUID {
	val := ctx.Value("user_id")
	if val == nil {
		return uuid.Nil
	}
	uidStr, ok := val.(string)
	if !ok {
		return uuid.Nil
	}
	uid, _ := uuid.Parse(uidStr)
	return uid
}

func (s *TripService) GetTrip(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TripService) GetTripByTripID(ctx context.Context, tripID string) (*models.Trip, error) {
	return s.repo.GetByTripID(ctx, tripID)
}

func (s *TripService) ListTrips(ctx context.Context, params models.TripListParams) ([]models.Trip, int, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}
	return s.repo.List(ctx, params)
}

func (s *TripService) UpdateTrip(ctx context.Context, id uuid.UUID, req *models.UpdateTripRequest) (*models.Trip, error) {
	// We need to update the repo method to accept updated_by or extract it internally.
	// Since I fixed the repo to use uuid.Nil as placeholder, let's fix it properly.
	// But first finish the service logic.
	if err := s.repo.Update(ctx, id, req); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *TripService) DeleteTrip(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// Telemetry
func (s *TripService) GetTelemetry(ctx context.Context, tripID uuid.UUID, limit int) ([]models.Telemetry, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetTelemetry(ctx, tripID, limit)
}

func (s *TripService) AddTelemetry(ctx context.Context, t *models.Telemetry) error {
	t.ID = uuid.New()
	if t.EventTime.IsZero() {
		t.EventTime = time.Now()
	}
	return s.repo.AddTelemetry(ctx, t)
}

// Settings
func (s *TripService) GetSettings(ctx context.Context, orgID uuid.UUID) (*models.TripSettings, error) {
	return s.repo.GetSettings(ctx, orgID)
}

func (s *TripService) UpdateSettings(ctx context.Context, settings *models.TripSettings) error {
	return s.repo.UpsertSettings(ctx, settings)
}
