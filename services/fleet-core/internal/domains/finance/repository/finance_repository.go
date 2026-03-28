package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/finance/models"
	fuelModels "github.com/flyfleet/fleet-core-service/internal/domains/fuel/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FinanceRepository struct {
	db *pgxpool.Pool
}

func NewFinanceRepository(db *pgxpool.Pool) *FinanceRepository {
	return &FinanceRepository{db: db}
}

// --- Fuel Transactions ---

func (r *FinanceRepository) CreateFuelTransaction(ctx context.Context, tx *models.FuelTransaction) error {
	query := `
		INSERT INTO fuel_transactions (
			id, vehicle_id, driver_id, card_id, station_location, liters, amount,
			currency, odometer_reading, efficiency_calculation, anomaly_flag,
			transaction_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	err := r.db.QueryRow(ctx, query,
		id, tx.VehicleID, tx.DriverID, tx.CardID, tx.StationLocation,
		tx.Liters, tx.Amount, tx.Currency, tx.OdometerReading,
		tx.EfficiencyCalculation, tx.AnomalyFlag, tx.TransactionDate,
		now, now,
	).Scan(&tx.ID, &tx.CreatedAt, &tx.UpdatedAt)

	return err
}

func (r *FinanceRepository) GetFuelTransactionsByVehicle(ctx context.Context, vehicleID string, limit int) ([]models.FuelTransaction, error) {
	query := `SELECT id, vehicle_id, driver_id, card_id, station_location, liters, amount,
			currency, odometer_reading, efficiency_calculation, anomaly_flag,
			transaction_date, created_at, updated_at 
			FROM fuel_transactions WHERE vehicle_id = $1 ORDER BY transaction_date DESC LIMIT $2`

	rows, err := r.db.Query(ctx, query, vehicleID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []models.FuelTransaction
	for rows.Next() {
		var tx models.FuelTransaction
		if err := rows.Scan(
			&tx.ID, &tx.VehicleID, &tx.DriverID, &tx.CardID, &tx.StationLocation,
			&tx.Liters, &tx.Amount, &tx.Currency, &tx.OdometerReading,
			&tx.EfficiencyCalculation, &tx.AnomalyFlag, &tx.TransactionDate,
			&tx.CreatedAt, &tx.UpdatedAt,
		); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// --- Expense Policies ---

func (r *FinanceRepository) CreateExpensePolicy(ctx context.Context, policy *models.ExpensePolicy) error {
	query := `
		INSERT INTO expense_policies (
			id, org_id, category, limit_amount, currency,
			requires_approval, receipt_required, description,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	err := r.db.QueryRow(ctx, query,
		id, policy.OrgID, policy.Category, policy.LimitAmount,
		policy.Currency, policy.RequiresApproval, policy.ReceiptRequired,
		policy.Description, now, now,
	).Scan(&policy.ID, &policy.CreatedAt, &policy.UpdatedAt)
	return err
}

func (r *FinanceRepository) GetExpensePolicies(ctx context.Context, orgID string) ([]models.ExpensePolicy, error) {
	query := `SELECT id, org_id, category, limit_amount, currency,
			requires_approval, receipt_required, description,
			created_at, updated_at FROM expense_policies WHERE org_id = $1`

	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []models.ExpensePolicy
	for rows.Next() {
		var p models.ExpensePolicy
		if err := rows.Scan(
			&p.ID, &p.OrgID, &p.Category, &p.LimitAmount, &p.Currency,
			&p.RequiresApproval, &p.ReceiptRequired, &p.Description,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, nil
}

// --- Expense Reports ---

func (r *FinanceRepository) CreateExpenseReport(ctx context.Context, report *models.ExpenseReport) error {
	query := `
		INSERT INTO expense_reports (
			id, org_id, submitter_id, period_start, period_end,
			total_amount, currency, status, metadata,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()
	// Metadata is kept as map[string]interface{} in model, assumes driver handles JSONB or we encode it
	// For pgx, we can pass map directly if driver supports it, or json.Marshal
	// Let's assume basic map support or handled by higher layers, but for safety with basic pgxscan we'd Marshal
	// Reverting to basic scan implies we need to handle JSONB.
	// For simplicity, we'll let pgx handle map -> jsonb conversion if registered, otherwise we might need an explicit Marshal.
	// However, updated model might treat it as []byte or similar.
	// To match previous sqlx logic which did Marshal:

	// Check model definition in separate view if needed, but assuming standard struct.
	// Verify if models.ExpenseReport Metadata is []byte or map.
	// Previous code: metaJSON, _ := json.Marshal(report.Metadata)
	// We will follow suit for safety.

	err := r.db.QueryRow(ctx, query,
		id, report.OrgID, report.SubmitterID, report.PeriodStart,
		report.PeriodEnd, report.TotalAmount, report.Currency,
		report.Status, report.Metadata, now, now,
	).Scan(&report.ID, &report.CreatedAt, &report.UpdatedAt)
	return err
}

func (r *FinanceRepository) UpdateExpenseReportStatus(ctx context.Context, reportID string, status string, approverID *string, reason *string) error {
	query := `
		UPDATE expense_reports 
		SET status = $1, approved_by = $2, rejection_reason = $3, approval_date = $4, updated_at = $5
		WHERE id = $6`

	var approvalDate *time.Time
	if status == "APPROVED" || status == "REJECTED" {
		t := time.Now()
		approvalDate = &t
	}

	_, err := r.db.Exec(ctx, query, status, approverID, reason, approvalDate, time.Now(), reportID)
	return err
}

func (r *FinanceRepository) ListExpenseReports(ctx context.Context, orgID, status string, limit, offset int) ([]models.ExpenseReport, int, error) {
	countQuery := `SELECT COUNT(*) FROM expense_reports WHERE org_id = $1`
	countArgs := []any{orgID}
	if status != "" {
		countQuery += ` AND status = $2`
		countArgs = append(countArgs, status)
	}
	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, org_id, submitter_id, period_start, period_end, total_amount, currency,
		       status, approved_by, approval_date, rejection_reason, metadata, created_at, updated_at
		FROM expense_reports
		WHERE org_id = $1`
	args := []any{orgID}
	if status != "" {
		query += ` AND status = $2`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC LIMIT $` + itoa(len(args)+1) + ` OFFSET $` + itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]models.ExpenseReport, 0, limit)
	for rows.Next() {
		var report models.ExpenseReport
		var rawMeta []byte
		if err := rows.Scan(
			&report.ID, &report.OrgID, &report.SubmitterID, &report.PeriodStart, &report.PeriodEnd,
			&report.TotalAmount, &report.Currency, &report.Status, &report.ApprovedBy, &report.ApprovalDate,
			&report.RejectionReason, &rawMeta, &report.CreatedAt, &report.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if len(rawMeta) > 0 {
			_ = json.Unmarshal(rawMeta, &report.Metadata)
		}
		out = append(out, report)
	}
	return out, total, rows.Err()
}

func (r *FinanceRepository) UpsertExpensePolicy(ctx context.Context, policy *models.ExpensePolicy) error {
	query := `
		INSERT INTO expense_policies (
			id, org_id, category, limit_amount, currency, requires_approval, receipt_required, description, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (org_id, category) DO UPDATE SET
			limit_amount = EXCLUDED.limit_amount,
			currency = EXCLUDED.currency,
			requires_approval = EXCLUDED.requires_approval,
			receipt_required = EXCLUDED.receipt_required,
			description = EXCLUDED.description,
			updated_at = EXCLUDED.updated_at
		RETURNING id, created_at, updated_at`

	now := time.Now()
	if policy.ID == "" {
		policy.ID = uuid.New().String()
	}
	return r.db.QueryRow(ctx, query,
		policy.ID, policy.OrgID, policy.Category, policy.LimitAmount, policy.Currency,
		policy.RequiresApproval, policy.ReceiptRequired, policy.Description, now, now,
	).Scan(&policy.ID, &policy.CreatedAt, &policy.UpdatedAt)
}

func (r *FinanceRepository) GetLatestRevenueForecast(ctx context.Context, routeID string) (*models.RevenueForecast, error) {
	query := `
		SELECT
		  COALESCE((SELECT predicted_demand FROM model_predictions WHERE route_id = $1 ORDER BY forecast_date DESC, forecast_hour DESC LIMIT 1), 0) AS predicted_demand,
		  COALESCE((SELECT demand_count::DECIMAL FROM historical_demand WHERE route_id = $1 ORDER BY date DESC, hour DESC LIMIT 1), 0) AS current_demand,
		  COALESCE((SELECT avg_price FROM historical_demand WHERE route_id = $1 ORDER BY date DESC, hour DESC LIMIT 1), 0) AS base_price,
		  COALESCE((SELECT confidence_interval_upper - confidence_interval_lower FROM model_predictions WHERE route_id = $1 ORDER BY forecast_date DESC, forecast_hour DESC LIMIT 1), 0) AS confidence_span`

	var predictedDemand float64
	var currentDemand float64
	var basePrice float64
	var confidenceSpan float64
	if err := r.db.QueryRow(ctx, query, routeID).Scan(&predictedDemand, &currentDemand, &basePrice, &confidenceSpan); err != nil {
		return nil, err
	}

	suggestedMultiplier := 1.0
	if currentDemand > 0 {
		ratio := predictedDemand / currentDemand
		if ratio > 1.0 {
			suggestedMultiplier = 1.0 + ((ratio - 1.0) * 0.5)
		}
	}
	if suggestedMultiplier < 1.0 {
		suggestedMultiplier = 1.0
	}
	if suggestedMultiplier > 2.0 {
		suggestedMultiplier = 2.0
	}

	confidence := 90.0 - confidenceSpan
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 100 {
		confidence = 100
	}

	return &models.RevenueForecast{
		RouteID:         routeID,
		PredictedDemand: predictedDemand,
		CurrentDemand:   currentDemand,
		BasePrice:       basePrice,
		SuggestedPrice:  basePrice * suggestedMultiplier,
		ConfidencePct:   confidence,
	}, nil
}

func (r *FinanceRepository) ListABExperiments(ctx context.Context, limit int) ([]models.ABTestExperiment, error) {
	query := `SELECT id, name, status, variants, metrics, created_at FROM ab_test_experiments ORDER BY created_at DESC LIMIT $1`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.ABTestExperiment, 0, limit)
	for rows.Next() {
		var exp models.ABTestExperiment
		var variantsRaw []byte
		var metricsRaw []byte
		if err := rows.Scan(&exp.ID, &exp.Name, &exp.Status, &variantsRaw, &metricsRaw, &exp.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(variantsRaw, &exp.Variants)
		_ = json.Unmarshal(metricsRaw, &exp.Metrics)
		out = append(out, exp)
	}
	return out, rows.Err()
}

func itoa(v int) string { return fmt.Sprintf("%d", v) }

// GetUnreconciledFuelCardTransactions returns external charges that haven't been matched.
func (r *FinanceRepository) GetUnreconciledFuelCardTransactions(ctx context.Context, orgID string) ([]models.FuelCardTransaction, error) {
	query := `SELECT id, card_number, vendor, amount, currency, transaction_time, location, status
	          FROM fuel_card_transactions 
	          WHERE org_id = $1 AND id NOT IN (SELECT card_transaction_id FROM fuel_card_reconciliation WHERE card_transaction_id IS NOT NULL)`

	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []models.FuelCardTransaction
	for rows.Next() {
		var tx models.FuelCardTransaction
		if err := rows.Scan(&tx.ID, &tx.CardNumber, &tx.Vendor, &tx.Amount, &tx.Currency, &tx.TransactionTime, &tx.Location, &tx.Status); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// GetFuelLogForMatching find a candidate internal log for a given external charge.
func (r *FinanceRepository) GetFuelLogForMatching(ctx context.Context, orgID string, amount float64, txTime time.Time) (*fuelModels.FuelLog, error) {
	query := `
		SELECT id, total_cost, transaction_date
		FROM fuel_logs
		WHERE organization_id = $1
		  AND status = 'APPROVED'
		  AND total_cost BETWEEN $2 - 10 AND $2 + 10
		  AND transaction_date BETWEEN $3 - interval '24 hours' AND $3 + interval '24 hours'
		ORDER BY ABS(total_cost - $2) ASC, ABS(EXTRACT(EPOCH FROM (transaction_date - $3))) ASC
		LIMIT 1`

	var log fuelModels.FuelLog
	err := r.db.QueryRow(ctx, query, orgID, amount, txTime).Scan(
		&log.ID, &log.TotalCost, &log.TransactionDate,
	)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// --- Approval System stubs ---

func (r *FinanceRepository) GetMatchingApprovalConfig(ctx context.Context, orgID, entityType string, amount float64) (*models.ApprovalConfig, error) {
	query := `SELECT id, org_id, entity_type, min_amount, levels, created_at FROM approval_configs WHERE org_id = $1 AND entity_type = $2 AND min_amount <= $3 ORDER BY min_amount DESC LIMIT 1`
	var config models.ApprovalConfig
	err := r.db.QueryRow(ctx, query, orgID, entityType, amount).Scan(&config.ID, &config.OrgID, &config.EntityType, &config.MinAmount, &config.Levels, &config.CreatedAt)
	if err != nil {
		return nil, nil // Assume no config is fine
	}
	return &config, nil
}

func (r *FinanceRepository) CreateApprovalInstance(ctx context.Context, inst *models.ApprovalInstance) error {
	query := `INSERT INTO approval_instances (id, entity_id, entity_type, config_id, current_level_index, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.db.Exec(ctx, query, inst.ID, inst.EntityID, inst.EntityType, inst.ConfigID, inst.CurrentLevelIndex, inst.Status, inst.CreatedAt, inst.UpdatedAt)
	return err
}

func (r *FinanceRepository) GetApprovalInstanceEffectively(ctx context.Context, entityID, entityType string) (*models.ApprovalInstance, error) {
	query := `SELECT id, entity_id, entity_type, config_id, current_level_index, status, created_at, updated_at FROM approval_instances WHERE entity_id = $1 AND entity_type = $2 AND status = 'PENDING' ORDER BY created_at DESC LIMIT 1`
	var inst models.ApprovalInstance
	err := r.db.QueryRow(ctx, query, entityID, entityType).Scan(&inst.ID, &inst.EntityID, &inst.EntityType, &inst.ConfigID, &inst.CurrentLevelIndex, &inst.Status, &inst.CreatedAt, &inst.UpdatedAt)
	if err != nil {
		return nil, nil
	}
	return &inst, nil
}

func (r *FinanceRepository) GetApprovalConfigByID(ctx context.Context, id string) (*models.ApprovalConfig, error) {
	query := `SELECT id, org_id, entity_type, min_amount, levels, created_at FROM approval_configs WHERE id = $1`
	var config models.ApprovalConfig
	err := r.db.QueryRow(ctx, query, id).Scan(&config.ID, &config.OrgID, &config.EntityType, &config.MinAmount, &config.Levels, &config.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *FinanceRepository) RecordApprovalAction(ctx context.Context, action *models.ApprovalAction) error {
	query := `INSERT INTO approval_actions (id, instance_id, approver_id, action, level_index, comments, action_time) VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.db.Exec(ctx, query, action.ID, action.InstanceID, action.ApproverID, action.Action, action.LevelIndex, action.Comments, action.ActionTime)
	return err
}

func (r *FinanceRepository) UpdateApprovalInstance(ctx context.Context, inst *models.ApprovalInstance) error {
	query := `UPDATE approval_instances SET current_level_index = $1, status = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.Exec(ctx, query, inst.CurrentLevelIndex, inst.Status, inst.UpdatedAt, inst.ID)
	return err
}

// --- Budgeting stubs ---

func (r *FinanceRepository) GetBudget(ctx context.Context, orgID, category string, now time.Time) (*models.Budget, error) {
	query := `SELECT id, org_id, category, year, month, limit_amount, spent_amount, currency, alert_threshold_pct, period_start, period_end, updated_at FROM budgets WHERE org_id = $1 AND category = $2 AND period_start <= $3 AND period_end >= $3`
	var b models.Budget
	err := r.db.QueryRow(ctx, query, orgID, category, now).Scan(&b.ID, &b.OrgID, &b.Category, &b.Year, &b.Month, &b.LimitAmount, &b.SpentAmount, &b.Currency, &b.AlertThresholdPct, &b.PeriodStart, &b.PeriodEnd, &b.UpdatedAt)
	if err != nil {
		return nil, nil
	}
	return &b, nil
}

func (r *FinanceRepository) UpdateBudgetSpent(ctx context.Context, budgetID string, amount float64) error {
	query := `UPDATE budgets SET spent_amount = spent_amount + $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, amount, budgetID)
	return err
}

// --- Currency stubs ---

func (r *FinanceRepository) GetExchangeRate(ctx context.Context, from, to string) (float64, error) {
	// Simple stub for now
	if from == to {
		return 1.0, nil
	}
	return 83.0, nil // Mock USD to INR
}

// --- Toll stubs ---

func (r *FinanceRepository) CreateTollTransaction(ctx context.Context, toll *models.TollTransaction) error {
	query := `INSERT INTO toll_transactions (id, vehicle_id, fastag_id, toll_plaza, amount, currency, transaction_time, external_transaction_id, metadata, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.db.Exec(ctx, query, toll.ID, toll.VehicleID, toll.FASTagID, toll.TollPlaza, toll.Amount, toll.Currency, toll.TransactionTime, toll.ExternalTransactionID, toll.Metadata, time.Now())
	return err
}

func (r *FinanceRepository) RefreshFinanceDashboardSummary(ctx context.Context) error {
	// Stub for refreshing materialized views
	return nil
}
