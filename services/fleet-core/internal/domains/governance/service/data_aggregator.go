package service

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/governance/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type GovernanceDataAggregator interface {
	AggregateData(ctx context.Context, userID, tenantID uuid.UUID) (*models.DSARData, error)
}

type governanceDataAggregator struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewGovernanceDataAggregator(db *pgxpool.Pool, logger *zap.Logger) GovernanceDataAggregator {
	return &governanceDataAggregator{db: db, logger: logger}
}

func (a *governanceDataAggregator) AggregateData(ctx context.Context, userID, tenantID uuid.UUID) (*models.DSARData, error) {
	data := &models.DSARData{}

	// 1. User Profile
	var userProfile map[string]interface{}
	err := a.db.QueryRow(ctx, `SELECT row_to_json(u) FROM (
		SELECT users.id, uo.tenant_id, email, first_name, last_name, phone_number, role, users.created_at 
		FROM users 
		JOIN user_organizations uo ON uo.user_id = users.id
		WHERE users.id = $1 AND uo.tenant_id = $2
	) u`, userID, tenantID).Scan(&userProfile)
	if err == nil {
		data.UserProfile = userProfile
	}

	// 2. Vehicles (Related to user's organization)
	data.Vehicles = a.queryAsList(ctx, "SELECT row_to_json(v) FROM vehicles v WHERE tenant_id = $1", tenantID)

	// 3. Trip History
	data.TripHistory = a.queryAsList(ctx, "SELECT row_to_json(t) FROM trips t WHERE tenant_id = $1 AND (driver_id = $2 OR EXISTS (SELECT 1 FROM drivers d WHERE d.id = t.driver_id AND d.user_id = $2))", tenantID, userID)

	// 4. Telemetry (Related to user's trips)
	data.Telemetry = a.queryAsList(ctx, "SELECT row_to_json(tel) FROM telemetry tel WHERE EXISTS (SELECT 1 FROM trips t WHERE t.id = tel.trip_id AND t.tenant_id = $1 AND (t.driver_id = $2 OR EXISTS (SELECT 1 FROM drivers d WHERE d.id = t.driver_id AND d.user_id = $2))) LIMIT 1000", tenantID, userID)

	// 5. Finance Records (Expense reports)
	data.FinanceRecords = a.queryAsList(ctx, "SELECT row_to_json(e) FROM expense_reports e WHERE tenant_id = $1 AND user_id = $2", tenantID, userID)

	// 6. Fuel Logs
	data.FuelLogs = a.queryAsList(ctx, "SELECT row_to_json(f) FROM fuel_logs f WHERE tenant_id = $1 AND (driver_id = $2 OR EXISTS (SELECT 1 FROM drivers d WHERE d.id = f.driver_id AND d.user_id = $2))", tenantID, userID)

	// 7. Geofence Events
	data.GeofenceEvents = a.queryAsList(ctx, "SELECT row_to_json(g) FROM geofence_events g WHERE tenant_id = $1 AND EXISTS (SELECT 1 FROM trips t JOIN drivers d ON t.driver_id = d.id WHERE d.user_id = $2 AND t.vehicle_id = g.vehicle_id AND g.timestamp BETWEEN t.start_time AND COALESCE(t.end_time, NOW()))", tenantID, userID)

	// 8. Governance Audit Logs
	data.AuditLogs = a.queryAsList(ctx, "SELECT row_to_json(al) FROM governance_audit_logs al WHERE tenant_id = $1 AND user_id = $2", tenantID, userID)

	return data, nil
}

func (a *governanceDataAggregator) queryAsList(ctx context.Context, query string, args ...interface{}) []interface{} {
	rows, err := a.db.Query(ctx, query, args...)
	if err != nil {
		a.logger.Error("failed to query as list", zap.String("query", query), zap.Error(err))
		return nil
	}
	defer rows.Close()

	var result []interface{}
	for rows.Next() {
		var item map[string]interface{}
		if err := rows.Scan(&item); err != nil {
			a.logger.Error("failed to scan row as map", zap.Error(err))
			continue
		}
		result = append(result, item)
	}
	return result
}
