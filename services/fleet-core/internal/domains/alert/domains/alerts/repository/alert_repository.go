package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/flyfleet/fleet-core/internal/domains/alert/domains/alerts/models"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type AlertRepository struct {
	db *pgxpool.Pool
}

func NewAlertRepository(db *pgxpool.Pool) *AlertRepository {
	return &AlertRepository{db: db}
}

// CreateRule inserts a new alert rule into the database
func (r *AlertRepository) CreateRule(ctx context.Context, rule *models.AlertRule) error {
	query := `
		INSERT INTO alert_rules (
			id, tenant_id, name, type, description, conditions, severity,
			actions, vehicles, driver_ids, enabled, created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	// Convert structs to JSON for storage
	conditionsJSON, _ := json.Marshal(rule.Conditions)
	actionsJSON, _ := json.Marshal(rule.Actions)
	vehiclesJSON, _ := json.Marshal(rule.Vehicles)
	driverIDsJSON, _ := json.Marshal(rule.DriverIDs)

	err := r.db.QueryRow(ctx, query,
		id, rule.OrgID, rule.Name, rule.Type, rule.Description,
		conditionsJSON, rule.Severity, actionsJSON, vehiclesJSON,
		driverIDsJSON, rule.Enabled, now, now, rule.CreatedBy, rule.UpdatedBy,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	return err
}

// GetRuleByID retrieves an alert rule by its ID
func (r *AlertRepository) GetRuleByID(ctx context.Context, id string) (*models.AlertRule, error) {
	query := `
		SELECT id, tenant_id, name, type, description, conditions, severity,
		       actions, vehicles, driver_ids, enabled, created_at, updated_at, created_by, updated_by
		FROM alert_rules WHERE id = $1`

	var rule models.AlertRule
	var conditionsJSON, actionsJSON, vehiclesJSON, driverIDsJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&rule.ID, &rule.OrgID, &rule.Name, &rule.Type, &rule.Description,
		&conditionsJSON, &rule.Severity, &actionsJSON, &vehiclesJSON,
		&driverIDsJSON, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt, &rule.CreatedBy, &rule.UpdatedBy,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	json.Unmarshal(conditionsJSON, &rule.Conditions)
	json.Unmarshal(actionsJSON, &rule.Actions)
	json.Unmarshal(vehiclesJSON, &rule.Vehicles)
	json.Unmarshal(driverIDsJSON, &rule.DriverIDs)

	return &rule, nil
}

// ListRules retrieves alert rules with filtering and pagination
func (r *AlertRepository) ListRules(ctx context.Context, orgID uuid.UUID, filter models.AlertListFilter) ([]*models.AlertRule, int, error) {
	baseQuery := `
		SELECT id, tenant_id, name, type, description, severity, enabled, created_at, updated_at
		FROM alert_rules WHERE tenant_id = $1`

	countQuery := `SELECT COUNT(*) FROM alert_rules WHERE tenant_id = $1`

	var conditions []string
	var args []interface{}
	argIndex := 2
	args = append(args, orgID)

	if len(filter.Type) > 0 {
		typePlaceholders := make([]string, len(filter.Type))
		for i, alertType := range filter.Type {
			typePlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, alertType)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(typePlaceholders, ",")))
	}

	if len(filter.Severity) > 0 {
		severityPlaceholders := make([]string, len(filter.Severity))
		for i, severity := range filter.Severity {
			severityPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, severity)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("severity IN (%s)", strings.Join(severityPlaceholders, ",")))
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	// Count query
	var totalCount int
	countSQL := countQuery + whereClause
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// List query with pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	orderBy := "created_at DESC"
	if filter.OrderBy != "" {
		allowedColumns := map[string]bool{
			"created_at": true, "name": true, "type": true, "severity": true,
		}
		if allowedColumns[filter.OrderBy] {
			direction := "ASC"
			if strings.ToUpper(filter.OrderDir) == "DESC" {
				direction = "DESC"
			}
			orderBy = fmt.Sprintf("%s %s", filter.OrderBy, direction)
		}
	}

	listSQL := baseQuery + whereClause + " ORDER BY " + orderBy + limitClause

	rows, err := r.db.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rules []*models.AlertRule
	for rows.Next() {
		var rule models.AlertRule
		err := rows.Scan(
			&rule.ID, &rule.OrgID, &rule.Name, &rule.Type, &rule.Description,
			&rule.Severity, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		rules = append(rules, &rule)
	}

	return rules, totalCount, nil
}

// UpdateRule modifies an existing alert rule
func (r *AlertRepository) UpdateRule(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for key, value := range updates {
		// Security check against SQL injection
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(key) {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))

		// Handle special JSON fields
		switch key {
		case "conditions", "actions", "vehicles", "driver_ids":
			jsonData, _ := json.Marshal(value)
			args = append(args, jsonData)
		default:
			args = append(args, value)
		}
		argIndex++
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE alert_rules SET %s, updated_at = $%d WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex, argIndex+1)

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// DeleteRule removes an alert rule
func (r *AlertRepository) DeleteRule(ctx context.Context, id string) error {
	query := "DELETE FROM alert_rules WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// CreateAlert inserts a new alert into the database
func (r *AlertRepository) CreateAlert(ctx context.Context, alert *models.Alert) error {
	query := `
		INSERT INTO alerts (
			id, tenant_id, rule_id, type, severity, vehicle_id, driver_id,
			location, timestamp, details, status, dedup_key, 
			escalation_level, sla_deadline, created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	// Convert structs to JSON for storage
	locationJSON, _ := json.Marshal(alert.Location)
	detailsJSON, _ := json.Marshal(alert.Details)

	err := r.db.QueryRow(ctx, query,
		id, alert.OrgID, alert.RuleID, alert.Type, alert.Severity,
		alert.VehicleID, alert.DriverID, locationJSON, alert.Timestamp,
		detailsJSON, alert.Status, alert.DedupKey,
		alert.EscalationLevel, alert.SLADeadline, now, now, alert.CreatedBy, alert.UpdatedBy,
	).Scan(&alert.ID, &alert.CreatedAt, &alert.UpdatedAt)

	return err
}

// GetAlertByID retrieves an alert by its ID
func (r *AlertRepository) GetAlertByID(ctx context.Context, id string) (*models.Alert, error) {
	query := `
		SELECT id, tenant_id, rule_id, type, severity, vehicle_id, driver_id,
		       location, timestamp, details, status, dedup_key, 
		       escalation_level, last_escalated_at, sla_deadline, 
		       acknowledged_by, acknowledged_at, resolved_by, resolved_at, 
		       escalated_to, notes, created_at, updated_at
		FROM alerts WHERE id = $1`

	var alert models.Alert
	var locationJSON, detailsJSON, escalatedToJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&alert.ID, &alert.OrgID, &alert.RuleID, &alert.Type, &alert.Severity,
		&alert.VehicleID, &alert.DriverID, &locationJSON, &alert.Timestamp,
		&detailsJSON, &alert.Status, &alert.DedupKey,
		&alert.EscalationLevel, &alert.LastEscalatedAt, &alert.SLADeadline,
		&alert.AcknowledgedBy, &alert.AcknowledgedAt,
		&alert.ResolvedBy, &alert.ResolvedAt, &escalatedToJSON, &alert.Notes,
		&alert.CreatedAt, &alert.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	json.Unmarshal(locationJSON, &alert.Location)
	json.Unmarshal(detailsJSON, &alert.Details)
	json.Unmarshal(escalatedToJSON, &alert.EscalatedTo)

	return &alert, nil
}

// ListAlerts retrieves alerts with filtering and pagination
func (r *AlertRepository) ListAlerts(ctx context.Context, filter models.AlertListFilter) ([]*models.Alert, int, error) {
	baseQuery := `
		SELECT id, tenant_id, rule_id, type, severity, vehicle_id, driver_id,
		       timestamp, status, created_at, updated_at
		FROM alerts WHERE tenant_id = $1`

	countQuery := `SELECT COUNT(*) FROM alerts WHERE tenant_id = $1`

	var conditions []string
	var args []interface{}
	argIndex := 2
	args = append(args, filter.OrgID)

	// Add date range filter
	if !filter.FromDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIndex))
		args = append(args, filter.FromDate)
		argIndex++
	}

	if !filter.ToDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argIndex))
		args = append(args, filter.ToDate)
		argIndex++
	}

	// Add status filter
	if len(filter.Status) > 0 {
		statusPlaceholders := make([]string, len(filter.Status))
		for i, status := range filter.Status {
			statusPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(statusPlaceholders, ",")))
	}

	// Add severity filter
	if len(filter.Severity) > 0 {
		severityPlaceholders := make([]string, len(filter.Severity))
		for i, severity := range filter.Severity {
			severityPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, severity)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("severity IN (%s)", strings.Join(severityPlaceholders, ",")))
	}

	// Add vehicle filter
	if filter.VehicleID != nil {
		conditions = append(conditions, fmt.Sprintf("vehicle_id = $%d", argIndex))
		args = append(args, filter.VehicleID)
		argIndex++
	}

	// Add driver filter
	if filter.DriverID != nil {
		conditions = append(conditions, fmt.Sprintf("driver_id = $%d", argIndex))
		args = append(args, filter.DriverID)
		argIndex++
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	// Count query
	var totalCount int
	countSQL := countQuery + whereClause
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// List query with pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	orderBy := "timestamp DESC"
	if filter.OrderBy != "" {
		allowedColumns := map[string]bool{
			"timestamp": true, "type": true, "severity": true, "status": true,
		}
		if allowedColumns[filter.OrderBy] {
			direction := "ASC"
			if strings.ToUpper(filter.OrderDir) == "DESC" {
				direction = "DESC"
			}
			orderBy = fmt.Sprintf("%s %s", filter.OrderBy, direction)
		}
	}

	listSQL := baseQuery + whereClause + " ORDER BY " + orderBy + limitClause

	rows, err := r.db.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var alerts []*models.Alert
	for rows.Next() {
		var alert models.Alert
		err := rows.Scan(
			&alert.ID, &alert.OrgID, &alert.RuleID, &alert.Type, &alert.Severity,
			&alert.VehicleID, &alert.DriverID, &alert.Timestamp, &alert.Status,
			&alert.CreatedAt, &alert.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		alerts = append(alerts, &alert)
	}

	return alerts, totalCount, nil
}

// UpdateAlertStatus updates the status of an alert
func (r *AlertRepository) UpdateAlertStatus(ctx context.Context, id string, status string, userID *uuid.UUID) error {
	var query string
	var args []interface{}

	switch status {
	case "acknowledged":
		query = "UPDATE alerts SET status = $1, acknowledged_by = $2, acknowledged_at = $3, updated_at = $4 WHERE id = $5"
		args = []interface{}{status, userID, time.Now(), time.Now(), id}
	case "resolved":
		query = "UPDATE alerts SET status = $1, resolved_by = $2, resolved_at = $3, updated_at = $4 WHERE id = $5"
		args = []interface{}{status, userID, time.Now(), time.Now(), id}
	default:
		query = "UPDATE alerts SET status = $1, updated_at = $2 WHERE id = $3"
		args = []interface{}{status, time.Now(), id}
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetActiveRulesForVehicle retrieves active alert rules for a specific vehicle
func (r *AlertRepository) GetActiveRulesForVehicle(ctx context.Context, vehicleID uuid.UUID) ([]*models.AlertRule, error) {
	query := `
		SELECT id, tenant_id, name, type, description, conditions, severity,
		       actions, vehicles, driver_ids, enabled, created_at, updated_at, created_by
		FROM alert_rules 
		WHERE enabled = TRUE 
		  AND (vehicles = '{}' OR $1 = ANY(vehicles))
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.AlertRule
	for rows.Next() {
		var rule models.AlertRule
		var conditionsJSON, actionsJSON, vehiclesJSON, driverIDsJSON []byte

		err := rows.Scan(
			&rule.ID, &rule.OrgID, &rule.Name, &rule.Type, &rule.Description,
			&conditionsJSON, &rule.Severity, &actionsJSON, &vehiclesJSON,
			&driverIDsJSON, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt, &rule.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		json.Unmarshal(conditionsJSON, &rule.Conditions)
		json.Unmarshal(actionsJSON, &rule.Actions)
		json.Unmarshal(vehiclesJSON, &rule.Vehicles)
		json.Unmarshal(driverIDsJSON, &rule.DriverIDs)

		rules = append(rules, &rule)
	}

	return rules, nil
}

// GetActiveRulesForDriver retrieves active alert rules for a specific driver
func (r *AlertRepository) GetActiveRulesForDriver(ctx context.Context, driverID uuid.UUID) ([]*models.AlertRule, error) {
	query := `
		SELECT id, tenant_id, name, type, description, conditions, severity,
		       actions, vehicles, driver_ids, enabled, created_at, updated_at, created_by
		FROM alert_rules 
		WHERE enabled = TRUE 
		  AND (driver_ids = '{}' OR $1 = ANY(driver_ids))
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.AlertRule
	for rows.Next() {
		var rule models.AlertRule
		var conditionsJSON, actionsJSON, vehiclesJSON, driverIDsJSON []byte

		err := rows.Scan(
			&rule.ID, &rule.OrgID, &rule.Name, &rule.Type, &rule.Description,
			&conditionsJSON, &rule.Severity, &actionsJSON, &vehiclesJSON,
			&driverIDsJSON, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt, &rule.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		json.Unmarshal(conditionsJSON, &rule.Conditions)
		json.Unmarshal(actionsJSON, &rule.Actions)
		json.Unmarshal(vehiclesJSON, &rule.Vehicles)
		json.Unmarshal(driverIDsJSON, &rule.DriverIDs)

		rules = append(rules, &rule)
	}

	return rules, nil
}

// GetAlertSummary retrieves alert statistics for an organization
func (r *AlertRepository) GetAlertSummary(ctx context.Context, orgID uuid.UUID, fromDate, toDate time.Time) (*models.AlertSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_alerts,
			COUNT(CASE WHEN status = 'new' THEN 1 END) as new_alerts,
			COUNT(CASE WHEN status = 'acknowledged' THEN 1 END) as acknowledged_alerts,
			COUNT(CASE WHEN status = 'resolved' THEN 1 END) as resolved_alerts,
			COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical_alerts,
			COUNT(CASE WHEN severity = 'warning' THEN 1 END) as warning_alerts,
			COUNT(CASE WHEN severity = 'info' THEN 1 END) as info_alerts
		FROM alerts 
		WHERE tenant_id = $1 
		  AND timestamp >= $2 
		  AND timestamp <= $3`

	var summary models.AlertSummary
	var criticalCount, warningCount, infoCount int
	err := r.db.QueryRow(ctx, query, orgID, fromDate, toDate).Scan(
		&summary.TotalAlerts, &summary.NewAlerts, &summary.Acknowledged,
		&summary.Resolved, &criticalCount, &warningCount, &infoCount,
	)

	if err != nil {
		return nil, err
	}

	// Initialize maps and populate counts
	summary.BySeverity = make(map[string]int)
	summary.BySeverity["critical"] = criticalCount
	summary.BySeverity["warning"] = warningCount
	summary.BySeverity["info"] = infoCount
	summary.ByType = make(map[string]int)
	summary.ByStatus = make(map[string]int)

	// Get breakdown by type
	typeQuery := `
		SELECT type, COUNT(*) 
		FROM alerts 
		WHERE tenant_id = $1 AND timestamp >= $2 AND timestamp <= $3
		GROUP BY type`

	typeRows, err := r.db.Query(ctx, typeQuery, orgID, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var alertType string
		var count int
		typeRows.Scan(&alertType, &count)
		summary.ByType[alertType] = count
	}

	// Get breakdown by status
	statusQuery := `
		SELECT status, COUNT(*) 
		FROM alerts 
		WHERE tenant_id = $1 AND timestamp >= $2 AND timestamp <= $3
		GROUP BY status`

	statusRows, err := r.db.Query(ctx, statusQuery, orgID, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status string
		var count int
		statusRows.Scan(&status, &count)
		summary.ByStatus[status] = count
	}

	return &summary, nil
}
func (r *AlertRepository) GetStats(ctx context.Context) (*models.AlertStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_alerts,
			COUNT(CASE WHEN status = 'new' THEN 1 END) as new_alerts,
			COUNT(CASE WHEN severity = 'critical' AND status != 'resolved' THEN 1 END) as critical_alerts,
			COUNT(CASE WHEN status != 'resolved' THEN 1 END) as open_alerts
		FROM alerts`

	var stats models.AlertStats
	err := r.db.QueryRow(ctx, query).Scan(&stats.TotalAlerts, &stats.NewAlerts, &stats.CriticalAlerts, &stats.OpenAlerts)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetTemplate retrieves a notification template by alert type, channel, language, and org.
// Returns sql.ErrNoRows if no matching template exists.
func (r *AlertRepository) GetTemplate(ctx context.Context, alertType, channel, language string, orgID uuid.UUID) (*models.Template, error) {
	query := `
		SELECT id, tenant_id, alert_type, channel, language, subject, body, created_at, updated_at
		FROM alert_templates
		WHERE alert_type = $1 AND channel = $2 AND language = $3 AND tenant_id = $4
		LIMIT 1`

	var tpl models.Template
	err := r.db.QueryRow(ctx, query, alertType, channel, language, orgID).Scan(
		&tpl.ID, &tpl.OrgID, &tpl.AlertType, &tpl.Channel, &tpl.Language,
		&tpl.Subject, &tpl.Body, &tpl.CreatedAt, &tpl.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}
