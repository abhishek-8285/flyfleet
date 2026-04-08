package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/finance/models"
	"github.com/flyfleet/fleet-core/internal/domains/finance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TDS section rates per Indian Income Tax Act
var tdsRates = map[string]float64{
	"194C":  2.0,  // Contractors & sub-contractors
	"194I":  10.0, // Rent
	"194J":  10.0, // Professional/technical services
	"194H":  5.0,  // Commission/brokerage
	"194Q":  0.1,  // Purchase of goods > 50L
	"194IA": 1.0,  // Transfer of immovable property
}

// TDSTaxService handles TDS deduction for vendor payments.
type TDSTaxService struct {
	repo   *repository.FinanceRepository
	logger *zap.Logger
}

func NewTDSTaxService(repo *repository.FinanceRepository, logger *zap.Logger) *TDSTaxService {
	return &TDSTaxService{repo: repo, logger: logger}
}

// DeductTDS calculates and records TDS for a vendor payment.
func (s *TDSTaxService) DeductTDS(ctx context.Context, req models.DeductTDSRequest) (*models.TDSRecord, error) {
	rate, ok := tdsRates[req.TDSSection]
	if !ok {
		return nil, fmt.Errorf("unknown TDS section: %s", req.TDSSection)
	}

	tdsAmount := (req.GrossAmount * rate) / 100
	netAmount := req.GrossAmount - tdsAmount

	// Determine financial year and quarter
	now := time.Now()
	fy := financialYear(now)
	quarter := taxQuarter(now)

	record := &models.TDSRecord{
		ID:            uuid.New(),
		EntityType:    req.EntityType,
		EntityID:      req.EntityID,
		VendorID:      req.VendorID,
		OrgID:         req.OrgID,
		TDSSection:    req.TDSSection,
		GrossAmount:   req.GrossAmount,
		TDSRate:       rate,
		TDSAmount:     tdsAmount,
		NetAmount:     netAmount,
		DeductedAt:    now,
		FinancialYear: fy,
		Quarter:       quarter,
	}

	if err := s.repo.CreateTDSRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to save TDS record: %w", err)
	}

	s.logger.Info("TDS deducted",
		zap.String("section", req.TDSSection),
		zap.Float64("rate", rate),
		zap.Float64("tds_amount", tdsAmount),
	)
	return record, nil
}

// GetTDSReport returns TDS records for an org in a financial year/quarter.
func (s *TDSTaxService) GetTDSReport(ctx context.Context, orgID uuid.UUID, fy, quarter string) ([]models.TDSRecord, error) {
	return s.repo.GetTDSRecords(ctx, orgID, fy, quarter)
}

// financialYear returns the Indian FY string (e.g. "2024-25").
func financialYear(t time.Time) string {
	year := t.Year()
	if t.Month() < 4 {
		return fmt.Sprintf("%d-%02d", year-1, year%100)
	}
	return fmt.Sprintf("%d-%02d", year, (year+1)%100)
}

// taxQuarter returns Q1–Q4 for the Indian fiscal year (Apr–Mar).
func taxQuarter(t time.Time) string {
	m := t.Month()
	switch {
	case m >= 4 && m <= 6:
		return "Q1"
	case m >= 7 && m <= 9:
		return "Q2"
	case m >= 10 && m <= 12:
		return "Q3"
	default:
		return "Q4"
	}
}
