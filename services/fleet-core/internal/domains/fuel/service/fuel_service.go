package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	finService "github.com/flyfleet/fleet-core/internal/domains/finance/service"
	"github.com/flyfleet/fleet-core/internal/domains/fuel/models"
	"github.com/flyfleet/fleet-core/internal/domains/fuel/repository"
)

const (
	// Duplicate prevention window (minutes)
	DuplicateFuelLogWindow = 5
)

type FuelService interface {
	CreateLog(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req models.CreateFuelLogRequest) (*models.FuelLog, error)
	GetLog(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*models.FuelLog, error)
	ListLogs(ctx context.Context, filter models.FuelLogFilter) (*models.FuelLogListResponse, error)
	UpdateLog(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req models.UpdateFuelLogRequest) (*models.FuelLog, error)
	DeleteLog(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	ChangeStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req models.ChangeFuelLogStatusRequest) error
	GetStats(ctx context.Context, tenantID uuid.UUID, filter models.FuelLogFilter) (*models.FuelLogStats, error)
	GetSettings(ctx context.Context, tenantID uuid.UUID) (*models.FuelSettings, error)
	UpdateSettings(ctx context.Context, tenantID uuid.UUID, req models.UpdateFuelSettingsRequest) (*models.FuelSettings, error)
}

type fuelService struct {
	repo         repository.FuelLogRepository
	settingsRepo repository.SettingsRepository
	anomalySvc   *AnomalyDetectionService
	taxService   *finService.TaxCalculationService
	logger       *zap.Logger
}

type fuelAuditWriter interface {
	CreateAuditLog(ctx context.Context, entry *models.FuelAuditLog) error
}

func NewFuelService(
	repo repository.FuelLogRepository,
	settingsRepo repository.SettingsRepository,
	anomalySvc *AnomalyDetectionService,
	taxService *finService.TaxCalculationService,
	logger *zap.Logger,
) FuelService {
	if anomalySvc == nil {
		anomalySvc = NewAnomalyDetectionService(nil, logger)
	}
	return &fuelService{
		repo:         repo,
		settingsRepo: settingsRepo,
		anomalySvc:   anomalySvc,
		taxService:   taxService,
		logger:       logger,
	}
}

func (s *fuelService) CreateLog(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req models.CreateFuelLogRequest) (*models.FuelLog, error) {
	// Validate odometer
	lastOdo, err := s.repo.GetLastOdometer(ctx, req.VehicleID, tenantID)
	if err == nil && lastOdo > 0 {
		if req.Odometer < lastOdo {
			return nil, ErrInvalidOdometer
		}
	}

	// Validate against tank capacity
	specs, err := s.repo.GetVehicleSpecs(ctx, req.VehicleID)
	if err == nil && specs != nil {
		if req.Quantity > specs.TankCapacityLiters {
			s.logger.Warn("Fuel quantity exceeds vehicle tank capacity",
				zap.String("vehicle_id", req.VehicleID.String()),
				zap.Float64("quantity", req.Quantity),
				zap.Float64("capacity", specs.TankCapacityLiters))
			return nil, ErrFuelQuantityExceedsCapacity
		}
	}

	// Check for duplicate transaction within window
	isDuplicate, err := s.repo.CheckDuplicate(ctx, req.VehicleID, tenantID, req.TransactionDate, DuplicateFuelLogWindow)
	if err != nil {
		s.logger.Error("failed to check for duplicate", zap.Error(err))
	}
	if isDuplicate {
		return nil, ErrDuplicateFuelLog
	}

	// Validate cost
	expectedCost := req.Quantity * req.UnitPrice
	if req.TotalCost < expectedCost-0.05 || req.TotalCost > expectedCost+0.05 {
		return nil, ErrInvalidCost
	}

	// GST Validation
	if req.VendorGSTIN != nil && *req.VendorGSTIN != "" {
		if !s.taxService.ValidateGSTIN(*req.VendorGSTIN) {
			s.logger.Warn("Invalid Vendor GSTIN provided", zap.String("gstin", *req.VendorGSTIN))
		}
	}
	if req.GSTNumber != nil && *req.GSTNumber != "" {
		if !s.taxService.ValidateGSTIN(*req.GSTNumber) {
			s.logger.Warn("Invalid Purchaser GSTIN provided", zap.String("gstin", *req.GSTNumber))
		}
	}

	// Build fuel log
	var gstAmount *float64
	if req.GSTRate != nil && *req.GSTRate > 0 {
		res := s.taxService.CalculateGST(req.TotalCost, *req.GSTRate, false)
		gstAmount = &res.GSTAmount
	}

	log := &models.FuelLog{
		ID:              uuid.New(),
		TenantID:  tenantID,
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
		GSTNumber:       req.GSTNumber,
		GSTRate:         req.GSTRate,
		GSTAmount:       gstAmount,
		InvoiceNumber:   req.InvoiceNumber,
		VendorGSTIN:     req.VendorGSTIN,
		ReceiptURL:      req.ReceiptURL,
		Notes:           req.Notes,
		CreatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Anomaly check
	var previousLog *models.FuelLog
	if lastOdo > 0 {
		previousLog = &models.FuelLog{Odometer: lastOdo}
	}
	if anomalyResult := s.anomalySvc.DetectFuelAnomalies(log, previousLog); anomalyResult != nil && anomalyResult.IsAnomaly {
		log.AnomalyFlag = true
		log.AnomalyReason = &anomalyResult.AnomalyType
	}

	// SECURITY FIX: Atomic check-and-insert to prevent TOCTOU race condition
	// This method handles limit validation and insertion in a single transaction
	if err := s.repo.CreateWithLimitValidation(ctx, log, req.Quantity, req.TotalCost); err != nil {
		// Check if it's a known limit error
		if strings.Contains(err.Error(), "limit exceeded") {
			s.logger.Warn("fuel limit validation failed",
				zap.String("driver_id", req.DriverID.String()),
				zap.Float64("quantity", req.Quantity),
				zap.Float64("amount", req.TotalCost),
				zap.Error(err))
			return nil, err
		}
		// Other errors
		s.logger.Error("failed to create fuel log with limit validation", zap.Error(err))
		return nil, err
	}

	// Audit log
	s.writeAuditLog(ctx, &models.FuelAuditLog{
		FuelLogID: log.ID,
		Action:    "create",
		ChangedBy: userID,
		NewValues: mustJSONMarshal(map[string]interface{}{
			"status":   log.Status,
			"quantity": log.Quantity,
		}),
	})

	return log, nil
}

func (s *fuelService) GetLog(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*models.FuelLog, error) {
	log, err := s.repo.GetByID(ctx, id, tenantID)
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

func (s *fuelService) UpdateLog(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req models.UpdateFuelLogRequest) (*models.FuelLog, error) {
	log, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	if log == nil {
		return nil, ErrLogNotFound
	}

	if log.Status == models.FuelLogStatusApproved || log.Status == models.FuelLogStatusRejected {
		return nil, ErrLogImmutable
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

	// GST Updates
	if req.GSTNumber != nil {
		log.GSTNumber = req.GSTNumber
		if *req.GSTNumber != "" && !s.taxService.ValidateGSTIN(*req.GSTNumber) {
			s.logger.Warn("Invalid Purchaser GSTIN provided in update", zap.String("gstin", *req.GSTNumber))
		}
	}
	if req.VendorGSTIN != nil {
		log.VendorGSTIN = req.VendorGSTIN
		if *req.VendorGSTIN != "" && !s.taxService.ValidateGSTIN(*req.VendorGSTIN) {
			s.logger.Warn("Invalid Vendor GSTIN provided in update", zap.String("gstin", *req.VendorGSTIN))
		}
	}
	if req.GSTRate != nil {
		log.GSTRate = req.GSTRate
		res := s.taxService.CalculateGST(log.TotalCost, *req.GSTRate, false)
		log.GSTAmount = &res.GSTAmount
	}
	if req.InvoiceNumber != nil {
		log.InvoiceNumber = req.InvoiceNumber
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
	s.writeAuditLog(ctx, &models.FuelAuditLog{
		FuelLogID: log.ID,
		Action:    "update",
		ChangedBy: log.CreatedBy,
		NewValues: mustJSONMarshal(map[string]interface{}{
			"status":       log.Status,
			"quantity":     log.Quantity,
			"total_cost":   log.TotalCost,
			"station_name": log.StationName,
		}),
	})

	return log, nil
}

func (s *fuelService) DeleteLog(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	log, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return err
	}
	if log == nil {
		return ErrLogNotFound
	}
	if log.Status == models.FuelLogStatusApproved || log.Status == models.FuelLogStatusRejected {
		return ErrLogImmutable
	}
	if err := s.repo.Delete(ctx, id, tenantID); err != nil {
		return err
	}
	s.writeAuditLog(ctx, &models.FuelAuditLog{
		FuelLogID: id,
		Action:    "delete",
		ChangedBy: log.CreatedBy,
		OldValues: mustJSONMarshal(map[string]interface{}{
			"status":       log.Status,
			"quantity":     log.Quantity,
			"total_cost":   log.TotalCost,
			"station_name": log.StationName,
		}),
	})
	return nil
}

func (s *fuelService) ChangeStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req models.ChangeFuelLogStatusRequest) error {
	log, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return err
	}
	if log == nil {
		return ErrLogNotFound
	}

	if log.Status != models.FuelLogStatusPending {
		return ErrStatusNotPending
	}

	// Prepare audit log data
	oldVal, _ := json.Marshal(map[string]interface{}{"status": string(log.Status)})
	newVal, _ := json.Marshal(map[string]interface{}{"status": string(req.Status)})

	_ = &models.FuelAuditLog{
		FuelLogID: id,
		Action:    string(req.Status),
		ChangedBy: uuid.Nil, // Would be populated from context
		ChangedAt: time.Now(),
		OldValues: oldVal,
		NewValues: newVal,
	}

	// Update status
	if err := s.repo.UpdateStatus(ctx, id, tenantID, req.Status); err != nil {
		return err
	}
	s.writeAuditLog(ctx, &models.FuelAuditLog{
		FuelLogID: id,
		Action:    strings.ToLower(string(req.Status)),
		ChangedBy: log.CreatedBy,
		OldValues: oldVal,
		NewValues: newVal,
	})
	s.logger.Info("Fuel log status changed",
		zap.String("log_id", id.String()),
		zap.String("old_status", string(log.Status)),
		zap.String("new_status", string(req.Status)),
	)

	return nil
}

func (s *fuelService) GetStats(ctx context.Context, tenantID uuid.UUID, filter models.FuelLogFilter) (*models.FuelLogStats, error) {
	return s.repo.GetStats(ctx, tenantID, filter)
}

// CalculateFuelEfficiency calculates fuel efficiency (km per liter) for a vehicle
func (s *fuelService) CalculateFuelEfficiency(ctx context.Context, vehicleID, tenantID uuid.UUID) (float64, error) {
	// Get last two approved fuel logs for the vehicle
	// This would need a new repository method to get last N logs
	// For now, return 0 as placeholder
	return 0.0, nil
}

// GetFuelEfficiencyHistory gets fuel efficiency history for a vehicle
func (s *fuelService) GetFuelEfficiencyHistory(ctx context.Context, vehicleID, tenantID uuid.UUID, limit int) ([]FuelEfficiencyRecord, error) {
	// This would need a new repository method
	// For now, return empty slice
	return []FuelEfficiencyRecord{}, nil
}

// FuelEfficiencyRecord represents a fuel efficiency calculation record
type FuelEfficiencyRecord struct {
	Date       time.Time `json:"date"`
	Odometer   float64   `json:"odometer"`
	FuelQty    float64   `json:"fuel_quantity"`
	Distance   float64   `json:"distance"`
	Efficiency float64   `json:"efficiency_km_per_liter"`
}

func (s *fuelService) GetSettings(ctx context.Context, tenantID uuid.UUID) (*models.FuelSettings, error) {
	return s.settingsRepo.GetOrDefault(ctx, tenantID)
}

func (s *fuelService) UpdateSettings(ctx context.Context, tenantID uuid.UUID, req models.UpdateFuelSettingsRequest) (*models.FuelSettings, error) {
	settings, err := s.settingsRepo.GetOrDefault(ctx, tenantID)
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

func (s *fuelService) validateFuelLimits(ctx context.Context, tenantID, driverID uuid.UUID, quantity, amount float64) error {
	// 1. Get Driver Specific Limits
	limit, err := s.repo.GetDriverFuelLimit(ctx, driverID, tenantID)
	if err != nil {
		s.logger.Error("failed to get driver fuel limit", zap.Error(err))
	}

	var dailyLiters, monthlyLiters, dailyAmount, monthlyAmount *float64
	if limit != nil {
		dailyLiters = limit.DailyLitersLimit
		monthlyLiters = limit.MonthlyLitersLimit
		dailyAmount = limit.DailyAmountLimit
		monthlyAmount = limit.MonthlyAmountLimit
	} else {
		// 2. Fallback to org settings
		orgSettings, err := s.repo.GetOrgFuelSettings(ctx, tenantID)
		if err != nil {
			s.logger.Error("failed to get org fuel settings", zap.Error(err))
		}
		if orgSettings != nil {
			dailyLiters = orgSettings.DefaultDailyLitersLimit
			monthlyLiters = orgSettings.DefaultMonthlyLitersLimit
			dailyAmount = orgSettings.DefaultDailyAmountLimit
			monthlyAmount = orgSettings.DefaultMonthlyAmountLimit
		}
	}

	if dailyLiters == nil && monthlyLiters == nil && dailyAmount == nil && monthlyAmount == nil {
		return nil // No limits set
	}

	// 3. Check Daily Usage
	startOfDay := time.Now().Truncate(24 * time.Hour)
	usedDailyLiters, usedDailyAmount, err := s.repo.GetFuelUsage(ctx, driverID, tenantID, startOfDay)
	if err == nil {
		if dailyLiters != nil && usedDailyLiters+quantity > *dailyLiters {
			return ErrDailyFuelLimitExceeded
		}
		if dailyAmount != nil && usedDailyAmount+amount > *dailyAmount {
			return ErrDailyFuelLimitExceeded
		}
	}

	// 4. Check Monthly Usage
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	usedMonthlyLiters, usedMonthlyAmount, err := s.repo.GetFuelUsage(ctx, driverID, tenantID, startOfMonth)
	if err == nil {
		if monthlyLiters != nil && usedMonthlyLiters+quantity > *monthlyLiters {
			return ErrMonthlyFuelLimitExceeded
		}
		if monthlyAmount != nil && usedMonthlyAmount+amount > *monthlyAmount {
			return ErrMonthlyFuelLimitExceeded
		}
	}

	return nil
}

func (s *fuelService) writeAuditLog(ctx context.Context, entry *models.FuelAuditLog) {
	writer, ok := s.repo.(fuelAuditWriter)
	if !ok || entry == nil {
		return
	}
	if entry.ChangedAt.IsZero() {
		entry.ChangedAt = time.Now().UTC()
	}
	if err := writer.CreateAuditLog(ctx, entry); err != nil {
		s.logger.Warn("failed to persist fuel audit log", zap.Error(err))
	}
}

func mustJSONMarshal(v map[string]interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
