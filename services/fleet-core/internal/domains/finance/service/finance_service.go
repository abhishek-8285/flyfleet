package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core/internal/domains/finance/repository"
	"go.uber.org/zap"
)

type FinanceService struct {
	repo   *repository.FinanceRepository
	logger *zap.Logger
}

func NewFinanceService(repo *repository.FinanceRepository, logger *zap.Logger) *FinanceService {
	return &FinanceService{repo: repo, logger: logger}
}

// IngestFuelData processes a fuel transaction and checks for anomalies
func (s *FinanceService) IngestFuelData(ctx context.Context, tx *models.FuelTransaction) error {
	if tx.VehicleID == "" {
		return errors.New("vehicle_id is required")
	}
	if tx.Liters <= 0 {
		return errors.New("liters must be greater than zero")
	}
	if tx.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if tx.Currency == "" {
		tx.Currency = "USD"
	}
	tx.Currency = strings.ToUpper(tx.Currency)
	if tx.TransactionDate.IsZero() {
		tx.TransactionDate = time.Now().UTC()
	}

	// 1. Calculate Efficiency if Odometer is present
	// simplified: in a real system we'd need the *previous* odometer reading to calculate distance
	// For now, we'll just check if the amount/liters ratio is within reasonable bounds (e.g., price per liter)

	const MaxPricePerLiter = 2.0 // USD
	const MinPricePerLiter = 0.5 // USD

	pricePerLiter := tx.Amount / tx.Liters
	if pricePerLiter > MaxPricePerLiter || pricePerLiter < MinPricePerLiter {
		tx.AnomalyFlag = true
		s.logger.Warn("Fuel anomaly detected: abnormal price per liter",
			zap.Float64("price", pricePerLiter),
			zap.String("vehicle_id", tx.VehicleID))
	}

	// Conservative safety checks until full vehicle/tank metadata and geofencing are integrated.
	if tx.Liters > 300 {
		tx.AnomalyFlag = true
		s.logger.Warn("Fuel anomaly detected: liters exceed fleet safety threshold",
			zap.Float64("liters", tx.Liters),
			zap.String("vehicle_id", tx.VehicleID))
	}
	if tx.StationLocation != nil && strings.TrimSpace(*tx.StationLocation) == "" {
		tx.AnomalyFlag = true
	}

	return s.repo.CreateFuelTransaction(ctx, tx)
}

// SubmitExpenseReport validation and submission
func (s *FinanceService) SubmitExpenseReport(ctx context.Context, report *models.ExpenseReport) error {
	// 1. Validate Totals
	if report.TotalAmount <= 0 {
		return errors.New("total_amount must be greater than zero")
	}
	if report.OrgID == "" || report.SubmitterID == "" {
		return errors.New("org_id and submitter_id are required")
	}
	if report.Currency == "" {
		report.Currency = "USD"
	}
	report.Currency = strings.ToUpper(report.Currency)
	if report.Metadata == nil {
		report.Metadata = map[string]interface{}{}
	}

	report.Status = "SUBMITTED"
	return s.repo.CreateExpenseReport(ctx, report)
}

// ApproveExpenseReport handles approval logic
func (s *FinanceService) ApproveExpenseReport(ctx context.Context, reportID, approverID string) error {
	return s.repo.UpdateExpenseReportStatus(ctx, reportID, "APPROVED", &approverID, nil)
}

// RejectExpenseReport handles rejection logic
func (s *FinanceService) RejectExpenseReport(ctx context.Context, reportID, approverID, reason string) error {
	return s.repo.UpdateExpenseReportStatus(ctx, reportID, "REJECTED", &approverID, &reason)
}

func (s *FinanceService) ListExpenseReports(ctx context.Context, orgID, status string, limit, offset int) (*models.ExpenseReportListResponse, error) {
	if strings.TrimSpace(orgID) == "" {
		return nil, errors.New("org_id is required")
	}
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	status = strings.ToUpper(strings.TrimSpace(status))

	reports, total, err := s.repo.ListExpenseReports(ctx, orgID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	return &models.ExpenseReportListResponse{
		Reports: reports,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

func (s *FinanceService) ListExpensePolicies(ctx context.Context, orgID string) ([]models.ExpensePolicy, error) {
	if strings.TrimSpace(orgID) == "" {
		return nil, errors.New("org_id is required")
	}
	return s.repo.GetExpensePolicies(ctx, orgID)
}

func (s *FinanceService) UpsertExpensePolicy(ctx context.Context, policy *models.ExpensePolicy) error {
	if policy == nil {
		return errors.New("policy is required")
	}
	if strings.TrimSpace(policy.OrgID) == "" {
		return errors.New("org_id is required")
	}
	if strings.TrimSpace(policy.Category) == "" {
		return errors.New("category is required")
	}
	policy.Category = strings.ToUpper(strings.TrimSpace(policy.Category))
	if policy.Currency == "" {
		policy.Currency = "USD"
	}
	policy.Currency = strings.ToUpper(strings.TrimSpace(policy.Currency))
	if policy.LimitAmount != nil && *policy.LimitAmount < 0 {
		return errors.New("limit_amount must be >= 0")
	}
	return s.repo.UpsertExpensePolicy(ctx, policy)
}

func (s *FinanceService) GetRevenueForecast(ctx context.Context, routeID string) (*models.RevenueForecast, error) {
	routeID = strings.TrimSpace(routeID)
	if routeID == "" {
		return nil, fmt.Errorf("route_id is required")
	}
	return s.repo.GetLatestRevenueForecast(ctx, routeID)
}

func (s *FinanceService) ListABExperiments(ctx context.Context, limit int) ([]models.ABTestExperiment, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListABExperiments(ctx, limit)
}
