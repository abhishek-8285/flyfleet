package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/fuel/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/fuel/repository"
)

var (
	ErrLogNotFound      = errors.New("fuel log not found")
	ErrInvalidOdometer = errors.New("odometer reading cannot be lower than previous entry")
	ErrInvalidCost     = errors.New("total cost must match quantity multiplied by unit price")
)

type FuelService interface {
	CreateLog(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, req models.CreateFuelLogRequest) (*models.FuelLog, error)
	GetLog(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.FuelLog, error)
	ListLogs(ctx context.Context, filter models.FuelLogFilter) (*models.FuelLogListResponse, error)
	UpdateLog(ctx context.Context, id uuid.UUID, orgID uuid.UUID, req models.UpdateFuelLogRequest) (*models.FuelLog, error)
	DeleteLog(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	ChangeStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, req models.ChangeFuelLogStatusRequest) error
	GetStats(ctx context.Context, orgID uuid.UUID, filter models.FuelLogFilter) (*models.FuelLogStats, error)
	GetSettings(ctx context.Context, orgID uuid.UUID) (*models.FuelSettings, error)
	UpdateSettings(ctx context.Context, orgID uuid.UUID, req models.UpdateFuelSettingsRequest) (*models.FuelSettings, error)
}

type fuelService struct {
	repo         repository.FuelLogRepository
	settingsRepo repository.SettingsRepository
	logger       *zap.Logger
}

func NewFuelService(repo repository.FuelLogRepository, settingsRepo repository.SettingsRepository, logger *zap.Logger) FuelService {
	return &fuelService{
		repo:         repo,
		settingsRepo: settingsRepo,
		logger:       logger,
	}
}

func (s *fuelService) CreateLog(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, req models.CreateFuelLogRequest) (*models.FuelLog, error) {
	// Validate odometer
	lastOdo, err := s.repo.GetLastOdometer(ctx, req.VehicleID, orgID)
	if err == nil && req.Odometer < lastOdo {
		return nil, ErrInvalidOdometer
	}

	// Validate cost (with small epsilon for floating point)
	expectedCost := req.Quantity * req.UnitPrice
	if req.TotalCost < expectedCost-0.05 || req.TotalCost > expectedCost+0.05 {
		return nil, ErrInvalidCost
	}

	log := &models.FuelLog{
		ID:              uuid.New(),
		OrganizationID:  orgID,
		VehicleID:       req.VehicleID,
		DriverID:        req.DriverID,
		StationName:     req.StationName,
		FuelType:        req.FuelType,
		Quantity:        req.Quantity,
		UnitPrice:       req.UnitPrice,
		TotalCost:       req.TotalCost,
		Odometer:        req.Odometer,
		TransactionDate: req.TransactionDate,
		Status:          models.FuelLogStatusPending,
		ReceiptURL:      req.ReceiptURL,
		Notes:           req.Notes,
		CreatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.Create(ctx, log); err != nil {
		s.logger.Error("failed to create fuel log", zap.Error(err))
		return nil, err
	}

	return log, nil
}

func (s *fuelService) GetLog(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.FuelLog, error) {
	log, err := s.repo.GetByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if log == nil {
		return nil, ErrLogNotFound
	}
	return log, nil
}

func (s *fuelService) ListLogs(ctx context.Context, filter models.FuelLogFilter) (*models.FuelLogListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	return s.repo.List(ctx, filter)
}

func (s *fuelService) UpdateLog(ctx context.Context, id uuid.UUID, orgID uuid.UUID, req models.UpdateFuelLogRequest) (*models.FuelLog, error) {
	log, err := s.repo.GetByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if log == nil {
		return nil, ErrLogNotFound
	}

	if log.Status == models.FuelLogStatusApproved {
		return nil, errors.New("approved fuel logs cannot be modified")
	}

	if req.StationName != nil {
		log.StationName = *req.StationName
	}
	if req.FuelType != nil {
		log.FuelType = *req.FuelType
	}
	if req.Quantity != nil {
		log.Quantity = *req.Quantity
	}
	if req.UnitPrice != nil {
		log.UnitPrice = *req.UnitPrice
	}
	if req.TotalCost != nil {
		log.TotalCost = *req.TotalCost
	}
	if req.Odometer != nil {
		log.Odometer = *req.Odometer
	}
	if req.TransactionDate != nil {
		log.TransactionDate = *req.TransactionDate
	}
	if req.ReceiptURL != nil {
		log.ReceiptURL = req.ReceiptURL
	}
	if req.Notes != nil {
		log.Notes = req.Notes
	}

	// Re-validate cost if changed
	if req.Quantity != nil || req.UnitPrice != nil || req.TotalCost != nil {
		expectedCost := log.Quantity * log.UnitPrice
		if log.TotalCost < expectedCost-0.05 || log.TotalCost > expectedCost+0.05 {
			return nil, ErrInvalidCost
		}
	}

	if err := s.repo.Update(ctx, log); err != nil {
		return nil, err
	}

	return log, nil
}

func (s *fuelService) DeleteLog(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	log, err := s.repo.GetByID(ctx, id, orgID)
	if err != nil {
		return err
	}
	if log == nil {
		return ErrLogNotFound
	}
	if log.Status == models.FuelLogStatusApproved {
		return errors.New("approved fuel logs cannot be deleted")
	}
	return s.repo.Delete(ctx, id, orgID)
}

func (s *fuelService) ChangeStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, req models.ChangeFuelLogStatusRequest) error {
	log, err := s.repo.GetByID(ctx, id, orgID)
	if err != nil {
		return err
	}
	if log == nil {
		return ErrLogNotFound
	}

	if log.Status != models.FuelLogStatusPending {
		return errors.New("only pending fuel logs can be approved or rejected")
	}

	return s.repo.UpdateStatus(ctx, id, orgID, req.Status)
}

func (s *fuelService) GetStats(ctx context.Context, orgID uuid.UUID, filter models.FuelLogFilter) (*models.FuelLogStats, error) {
	return s.repo.GetStats(ctx, orgID, filter)
}

func (s *fuelService) GetSettings(ctx context.Context, orgID uuid.UUID) (*models.FuelSettings, error) {
	return s.settingsRepo.GetOrDefault(ctx, orgID)
}

func (s *fuelService) UpdateSettings(ctx context.Context, orgID uuid.UUID, req models.UpdateFuelSettingsRequest) (*models.FuelSettings, error) {
	settings, err := s.settingsRepo.GetOrDefault(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if req.DefaultCurrency != nil {
		settings.DefaultCurrency = *req.DefaultCurrency
	}
	if req.AllowedFuelTypes != nil {
		settings.AllowedFuelTypes = *req.AllowedFuelTypes
	}
	if req.ValidationRules != nil {
		settings.ValidationRules = *req.ValidationRules
	}
	if req.AutoSyncEnabled != nil {
		settings.AutoSyncEnabled = *req.AutoSyncEnabled
	}

	if err := s.settingsRepo.Upsert(ctx, settings); err != nil {
		return nil, err
	}

	return settings, nil
}
