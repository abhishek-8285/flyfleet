package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/vehicle/models"
)

type VehicleRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewVehicleRepository(db *pgxpool.Pool, logger *zap.Logger) *VehicleRepository {
	return &VehicleRepository{db: db, logger: logger}
}

func (r *VehicleRepository) Create(ctx context.Context, vehicle *models.Vehicle) error {
	// 1. Initial Compliance Evaluation
	res := vehicle.EvaluateCompliance()
	if res.IsHardBlocked {
		vehicle.IsHardBlocked = true
		vehicle.Status = models.StatusSuspended
	}

	if err := vehicle.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 2. Transactional Create
	tx, err := r.db.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO vehicles (
			tenant_id, registration_number, chassis_number, engine_number, vin_number,
			vehicle_type, capacity_tonnage, fuel_type, fuel_tank_capacity,
			ownership_type, parent_vehicle_id, branch_id, cost_center_id, status,
			gps_device_id, gps_enabled, gps_device_type, immobilizer_enabled, ais140_compliant,
			make, model, year, color, odometer, current_driver_id, speed_limit,
			is_blacklisted, is_hard_blocked, risk_score, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, NOW(), NOW(), 1)
		RETURNING id, created_at, updated_at, version`

	err = tx.QueryRow(ctx, query,
		vehicle.TenantID, vehicle.RegistrationNumber, vehicle.ChassisNumber, vehicle.EngineNumber, vehicle.VINNumber,
		vehicle.VehicleType, vehicle.CapacityTonnage, vehicle.FuelType, vehicle.FuelTankCapacity,
		vehicle.OwnershipType, vehicle.ParentVehicleID, vehicle.BranchID, vehicle.CostCenterID, vehicle.Status,
		vehicle.GPSDeviceID, vehicle.GPSEnabled, vehicle.GPSDeviceType, vehicle.ImmobilizerEnabled, vehicle.AIS140Compliant,
		vehicle.Make, vehicle.Model, vehicle.Year, vehicle.Color, vehicle.Odometer, vehicle.CurrentDriverID, vehicle.SpeedLimit,
		vehicle.IsBlacklisted, vehicle.IsHardBlocked, vehicle.RiskScore,
	).Scan(&vehicle.ID, &vehicle.CreatedAt, &vehicle.UpdatedAt, &vehicle.Version)
	if err != nil { return err }

	// 3. Side Effects (Audit + Outbox)
	r.txLogAudit(ctx, tx, vehicle.TenantID, vehicle.ID, uuid.Nil, models.ActionCreate, "ALL", "", "Vehicle Created")
	if res.IsHardBlocked {
		r.txLogAudit(ctx, tx, vehicle.TenantID, vehicle.ID, uuid.Nil, models.ActionHardBlock, "status", string(models.StatusProvisioning), string(models.StatusSuspended))
		payload, _ := json.Marshal(res.Breaches)
		r.txEmitEvent(ctx, tx, vehicle.TenantID, models.OutboxComplianceBreach, payload)
	}

	return tx.Commit(ctx)
}

func (r *VehicleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) {
	query := `
		SELECT 
			v.id, v.tenant_id, COALESCE(v.registration_number, ''), 
			COALESCE(v.chassis_number, ''), COALESCE(v.engine_number, ''), COALESCE(v.vin_number, ''),
			COALESCE(v.vehicle_type, 'TRUCK'), COALESCE(v.capacity_tonnage, 0), 
			COALESCE(v.fuel_type, 'DIESEL'), COALESCE(v.fuel_tank_capacity, 0),
			COALESCE(v.ownership_type, 'OWNED'), 
			COALESCE(v.parent_vehicle_id, '00000000-0000-0000-0000-000000000000'::uuid), 
			COALESCE(v.branch_id, '00000000-0000-0000-0000-000000000000'::uuid), 
			COALESCE(v.cost_center_id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(v.status, 'ACTIVE'), COALESCE(v.gps_device_id, ''), COALESCE(v.gps_enabled, false),
			COALESCE(v.gps_device_type, ''), COALESCE(v.immobilizer_enabled, false), COALESCE(v.ais140_compliant, false),
			COALESCE(v.make, ''), COALESCE(v.model, ''), COALESCE(v.year, 0), 
			COALESCE(v.color, ''), COALESCE(v.odometer, 0), 
			COALESCE(v.current_driver_id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(v.speed_limit, 0), COALESCE(v.is_blacklisted, false), COALESCE(v.is_hard_blocked, false), 
			COALESCE(v.risk_score, 0), v.created_at, v.updated_at, v.deleted_at, v.version,
			COALESCE(d.full_name, '') as current_driver_name,
			d.license_expiry,
			COALESCE(alerts.alert_count, 0) as active_alerts,
			COALESCE(alerts.critical_count, 0) as critical_alerts,
			COALESCE(alerts.overspeed_count, 0) as overspeed_alerts,
			COALESCE(alerts.geofence_violation, false) as geofence_violation,
			t.recorded_at as last_seen,
			COALESCE(t.ignition_status, 'UNKNOWN') as ignition_status,
			COALESCE(t.latitude, 0) as lat,
			COALESCE(t.longitude, 0) as lng,
			(SELECT id FROM trips WHERE vehicle_id = v.id AND status = 'IN_PROGRESS' LIMIT 1) IS NOT NULL as active_trip,
			next_doc.expiry_date as next_expiry_date,
			next_doc.type as next_expiry_type
		FROM vehicles v
		LEFT JOIN drivers d ON v.current_driver_id = d.id
		LEFT JOIN LATERAL (
			SELECT 
				COUNT(*) as alert_count,
				COUNT(*) FILTER (WHERE severity = 'CRITICAL') as critical_count,
				COUNT(*) FILTER (WHERE type = 'OVERSPEED') as overspeed_count,
				BOOL_OR(type = 'GEOFENCE_EXIT') as geofence_violation
			FROM alerts 
			WHERE vehicle_id = v.id AND status = 'ACTIVE'
		) alerts ON true
		LEFT JOIN LATERAL (
			SELECT recorded_at, ignition_status, latitude, longitude
			FROM telemetry_points
			WHERE vehicle_id = v.id
			ORDER BY recorded_at DESC
			LIMIT 1
		) t ON true
		LEFT JOIN LATERAL (
			SELECT expiry_date, type
			FROM vehicle_documents
			WHERE vehicle_id = v.id AND status = 'APPROVED'
			ORDER BY expiry_date ASC
			LIMIT 1
		) next_doc ON true
		WHERE v.id = $1 AND v.deleted_at IS NULL`

	v := &models.Vehicle{Summary: &models.VehicleSummary{}}
	var (
		driverName    *string
		licenseExpiry *time.Time
		alerts        int
		criticals     int
		overspeeds    int
		geofence      bool
		lastSeen      *time.Time
		ignition      string
		lat, lng      float64
		activeTrip    bool
		nextExpiry    *time.Time
		nextType      *string
	)

	err := r.db.QueryRow(ctx, query, id).Scan(
		&v.ID, &v.TenantID, &v.RegistrationNumber, &v.ChassisNumber, &v.EngineNumber, &v.VINNumber,
		&v.VehicleType, &v.CapacityTonnage, &v.FuelType, &v.FuelTankCapacity, &v.OwnershipType,
		&v.ParentVehicleID, &v.BranchID, &v.CostCenterID, &v.Status,
		&v.GPSDeviceID, &v.GPSEnabled, &v.GPSDeviceType, &v.ImmobilizerEnabled, &v.AIS140Compliant,
		&v.Make, &v.Model, &v.Year, &v.Color, &v.Odometer, &v.CurrentDriverID, &v.SpeedLimit,
		&v.IsBlacklisted, &v.IsHardBlocked, &v.RiskScore,
		&v.CreatedAt, &v.UpdatedAt, &v.DeletedAt, &v.Version,
		&driverName, &licenseExpiry,
		&alerts, &criticals, &overspeeds, &geofence, &lastSeen, &ignition, &lat, &lng, &activeTrip,
		&nextExpiry, &nextType,
	)
	if err != nil {
		if err == pgx.ErrNoRows { return nil, fmt.Errorf("vehicle not found") }
		return nil, err
	}

	r.populateSummary(v, driverName, licenseExpiry, alerts, criticals, overspeeds, geofence, lastSeen, ignition, activeTrip, lat, lng, nextExpiry, nextType)
	return v, nil
}

func (r *VehicleRepository) populateSummary(v *models.Vehicle, driverName *string, licenseExpiry *time.Time, alerts, criticals, overspeeds int, geofence bool, lastSeen *time.Time, ignition string, activeTrip bool, lat, lng float64, nextExpiry *time.Time, nextType *string) {
	if v.CurrentDriverID != nil && driverName != nil {
		v.CurrentDriver = &models.DriverInfo{ID: *v.CurrentDriverID, Name: *driverName}
		v.Summary.DriverAssigned = true
		v.Summary.DriverLicenseStatus = models.LicenseMissing
		if licenseExpiry != nil {
			if licenseExpiry.Before(time.Now()) {
				v.Summary.DriverLicenseStatus = models.LicenseExpired
			} else {
				v.Summary.DriverLicenseStatus = models.LicenseValid
			}
		}
	}

	v.Summary.AlertCount = alerts
	v.Summary.CriticalAlertCount = criticals
	v.Summary.OverSpeedCount = overspeeds
	v.Summary.GeoFenceViolation = geofence
	v.Summary.ActiveTrip = activeTrip
	v.Summary.LastSeen = lastSeen
	v.Summary.LastLocationLat = lat
	v.Summary.LastLocationLng = lng
	v.Summary.RiskScore = v.RiskScore
	v.Summary.RiskCategory = models.GetRiskCategory(v.RiskScore)
	v.Summary.IgnitionStatus = models.IgnitionStatus(ignition)
	
	// Runtime Evaluation
	res := v.EvaluateCompliance()
	v.Summary.IsHardBlocked = res.IsHardBlocked || v.IsHardBlocked

	// GPS Status & Freshness
	v.Summary.GPSStatus = models.GPSOffline
	if lastSeen != nil {
		v.Summary.TelemetryDelayMs = time.Since(*lastSeen).Milliseconds()
		v.Summary.DataFreshnessScore = models.GetDataFreshnessScore(v.Summary.TelemetryDelayMs)
		
		if v.Summary.TelemetryDelayMs < 60000 { // 1 min
			v.Summary.GPSStatus = models.GPSOnline
		} else if v.Summary.TelemetryDelayMs < 300000 { // 5 mins
			v.Summary.GPSStatus = models.GPSStale
		}
	}

	// Compliance Status based on next expiry
	v.Summary.ComplianceStatus = models.ComplianceGreen
	v.Summary.ComplianceScore = 100
	if nextExpiry != nil {
		v.Summary.NextExpiryDate = nextExpiry
		if nextType != nil {
			v.Summary.NextExpiryType = models.DocumentType(*nextType)
		}

		now := time.Now()
		expiryThreshold := now.AddDate(0, 0, 30)

		if nextExpiry.Before(now) {
			v.Summary.ComplianceStatus = models.ComplianceRed
			v.Summary.ComplianceScore = 0
		} else if nextExpiry.Before(expiryThreshold) {
			v.Summary.ComplianceStatus = models.ComplianceYellow
			v.Summary.ComplianceScore = 70
		}
	}

	v.Summary.MaintenanceStatus = models.MaintHealthy
	if v.Status == models.StatusMaintenance { v.Summary.MaintenanceStatus = models.MaintDue }
	v.Summary.FuelEfficiency30d = 8.5
}

func (r *VehicleRepository) Update(ctx context.Context, vehicle *models.Vehicle) error {
	// 1. Compliance Evaluation
	res := vehicle.EvaluateCompliance()
	if res.IsHardBlocked {
		vehicle.IsHardBlocked = true
		vehicle.Status = models.StatusSuspended
	}

	if err := vehicle.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 2. Transactional Update
	tx, err := r.db.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)

	query := `
		UPDATE vehicles SET
			registration_number  = COALESCE($1, registration_number),
			chassis_number       = COALESCE($2, chassis_number),
			engine_number        = COALESCE($3, engine_number),
			vin_number           = COALESCE($4, vin_number),
			vehicle_type         = COALESCE($5, vehicle_type),
			status               = COALESCE($6, status),
			is_blacklisted       = $7,
			is_hard_blocked      = $8,
			updated_at           = NOW(),
			version              = version + 1
		WHERE id = $9 AND deleted_at IS NULL AND version = $10
		RETURNING updated_at, version`

	err = tx.QueryRow(ctx, query,
		vehicle.RegistrationNumber, vehicle.ChassisNumber, vehicle.EngineNumber, vehicle.VINNumber,
		vehicle.VehicleType, vehicle.Status, vehicle.IsBlacklisted, vehicle.IsHardBlocked, vehicle.ID, vehicle.Version,
	).Scan(&vehicle.UpdatedAt, &vehicle.Version)
	if err != nil { return err }

	// 3. Side Effects
	if res.Changed && res.IsHardBlocked {
		r.txLogAudit(ctx, tx, vehicle.TenantID, vehicle.ID, uuid.Nil, models.ActionHardBlock, "status", "PREVIOUS", string(models.StatusSuspended))
		payload, _ := json.Marshal(res.Breaches)
		r.txEmitEvent(ctx, tx, vehicle.TenantID, models.OutboxComplianceBreach, payload)
	}

	return tx.Commit(ctx)
}

func (r *VehicleRepository) txLogAudit(ctx context.Context, tx pgx.Tx, tenantID, vehicleID, userID uuid.UUID, action models.AuditAction, field, old, new string) {
	query := `INSERT INTO vehicle_audit_logs (id, tenant_id, vehicle_id, user_id, action, field_name, old_value, new_value, timestamp)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`
	_, _ = tx.Exec(ctx, query, uuid.New(), tenantID, vehicleID, userID, action, field, old, new)
}

func (r *VehicleRepository) txEmitEvent(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, eventType models.OutboxEventType, payload []byte) {
	query := `INSERT INTO vehicle_outbox (id, tenant_id, event_type, payload, created_at, expires_at)
	          VALUES ($1, $2, $3, $4, NOW(), NOW() + INTERVAL '7 days')`
	_, _ = tx.Exec(ctx, query, uuid.New(), tenantID, eventType, payload)
}

func (r *VehicleRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Vehicle, int64, error) {
	return r.Search(ctx, models.VehicleSearchParams{
		TenantID: &tenantID,
		Page:     (offset / limit) + 1,
		Limit:    limit,
	})
}

func (r *VehicleRepository) Search(ctx context.Context, params models.VehicleSearchParams) ([]*models.Vehicle, int64, error) {
	var allCount int
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM vehicles").Scan(&allCount)
	r.logger.Debug("search vehicles debug count", zap.Int("all_vehicles_in_db", allCount))

	query := `
		SELECT 
			v.id, v.tenant_id, COALESCE(v.registration_number, ''), 
			COALESCE(v.chassis_number, ''), COALESCE(v.engine_number, ''), COALESCE(v.vin_number, ''),
			COALESCE(v.vehicle_type, 'TRUCK'), COALESCE(v.capacity_tonnage, 0), 
			COALESCE(v.fuel_type, 'DIESEL'), COALESCE(v.fuel_tank_capacity, 0),
			COALESCE(v.ownership_type, 'OWNED'), 
			COALESCE(v.parent_vehicle_id, '00000000-0000-0000-0000-000000000000'::uuid), 
			COALESCE(v.branch_id, '00000000-0000-0000-0000-000000000000'::uuid), 
			COALESCE(v.cost_center_id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(v.status, 'ACTIVE'), COALESCE(v.gps_device_id, ''), COALESCE(v.gps_enabled, false),
			COALESCE(v.gps_device_type, ''), COALESCE(v.immobilizer_enabled, false), COALESCE(v.ais140_compliant, false),
			COALESCE(v.make, ''), COALESCE(v.model, ''), COALESCE(v.year, 0), 
			COALESCE(v.color, ''), COALESCE(v.odometer, 0), 
			COALESCE(v.current_driver_id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(v.speed_limit, 0), COALESCE(v.is_blacklisted, false), COALESCE(v.is_hard_blocked, false), 
			COALESCE(v.risk_score, 0), v.created_at, v.updated_at, v.deleted_at, v.version,
			COALESCE(d.full_name, '') as current_driver_name,
			d.license_expiry,
			COALESCE(alerts.alert_count, 0) as active_alerts,
			COALESCE(alerts.critical_count, 0) as critical_alerts,
			COALESCE(alerts.overspeed_count, 0) as overspeed_alerts,
			COALESCE(alerts.geofence_violation, false) as geofence_violation,
			t.recorded_at as last_seen,
			COALESCE(t.ignition_status, 'UNKNOWN') as ignition_status,
			COALESCE(t.latitude, 0) as lat,
			COALESCE(t.longitude, 0) as lng,
			(SELECT id FROM trips WHERE vehicle_id = v.id AND status = 'IN_PROGRESS' LIMIT 1) IS NOT NULL as active_trip,
			next_doc.expiry_date as next_expiry_date,
			next_doc.type as next_expiry_type
		FROM vehicles v
		LEFT JOIN drivers d ON v.current_driver_id = d.id
		LEFT JOIN LATERAL (
			SELECT 
				COUNT(*) as alert_count,
				COUNT(*) FILTER (WHERE severity = 'CRITICAL') as critical_count,
				COUNT(*) FILTER (WHERE type = 'OVERSPEED') as overspeed_count,
				BOOL_OR(type = 'GEOFENCE_EXIT') as geofence_violation
			FROM alerts 
			WHERE vehicle_id = v.id AND status = 'ACTIVE'
		) alerts ON true
		LEFT JOIN LATERAL (
			SELECT recorded_at, ignition_status, latitude, longitude
			FROM telemetry_points
			WHERE vehicle_id = v.id
			ORDER BY recorded_at DESC
			LIMIT 1
		) t ON true
		LEFT JOIN LATERAL (
			SELECT expiry_date, type
			FROM vehicle_documents
			WHERE vehicle_id = v.id AND status = 'APPROVED'
			ORDER BY expiry_date ASC
			LIMIT 1
		) next_doc ON true
		WHERE v.deleted_at IS NULL`

	args := []interface{}{}
	argIdx := 1

	if params.TenantID != nil {
		query += fmt.Sprintf(" AND v.tenant_id = $%d", argIdx)
		args = append(args, *params.TenantID)
		argIdx++
	}

	if params.Status != "" {
		query += fmt.Sprintf(" AND v.status = $%d", argIdx)
		args = append(args, params.Status)
		argIdx++
	}

	if params.Query != "" {
		query += fmt.Sprintf(" AND (v.registration_number ILIKE $%d OR v.make ILIKE $%d OR v.model ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+params.Query+"%")
		argIdx++
	}

	// Pagination
	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := (params.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	query += fmt.Sprintf(" ORDER BY v.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("search vehicles query failed", zap.Error(err), zap.String("query", query))
		return nil, 0, err
	}
	r.logger.Debug("search vehicles query executed", zap.String("query", query), zap.Any("args", args))
	defer rows.Close()

	var vehicles []*models.Vehicle
	for rows.Next() {
		v := &models.Vehicle{Summary: &models.VehicleSummary{}}
		var (
			driverName    *string
			licenseExpiry *time.Time
			alerts        int
			criticals     int
			overspeeds    int
			geofence      bool
			lastSeen      *time.Time
			ignition      string
			lat, lng      float64
			activeTrip    bool
			nextExpiry    *time.Time
			nextType      *string
		)

		err := rows.Scan(
			&v.ID, &v.TenantID, &v.RegistrationNumber, &v.ChassisNumber, &v.EngineNumber, &v.VINNumber,
			&v.VehicleType, &v.CapacityTonnage, &v.FuelType, &v.FuelTankCapacity, &v.OwnershipType,
			&v.ParentVehicleID, &v.BranchID, &v.CostCenterID, &v.Status,
			&v.GPSDeviceID, &v.GPSEnabled, &v.GPSDeviceType, &v.ImmobilizerEnabled, &v.AIS140Compliant,
			&v.Make, &v.Model, &v.Year, &v.Color, &v.Odometer, &v.CurrentDriverID, &v.SpeedLimit,
			&v.IsBlacklisted, &v.IsHardBlocked, &v.RiskScore,
			&v.CreatedAt, &v.UpdatedAt, &v.DeletedAt, &v.Version,
			&driverName, &licenseExpiry,
			&alerts, &criticals, &overspeeds, &geofence, &lastSeen, &ignition, &lat, &lng, &activeTrip,
			&nextExpiry, &nextType,
		)
		if err != nil {
			r.logger.Error("failed to scan search result", zap.Error(err))
			return nil, 0, err
		}

		r.populateSummary(v, driverName, licenseExpiry, alerts, criticals, overspeeds, geofence, lastSeen, ignition, activeTrip, lat, lng, nextExpiry, nextType)
		vehicles = append(vehicles, v)
	}

	countQuery := "SELECT COUNT(*) FROM vehicles WHERE deleted_at IS NULL"
	if params.TenantID != nil {
		countQuery += fmt.Sprintf(" AND tenant_id = '%s'", params.TenantID.String())
	}
	var total int64
	err = r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		r.logger.Error("count vehicles query failed", zap.Error(err), zap.String("query", countQuery))
	} else {
		r.logger.Debug("count vehicles query executed", zap.String("query", countQuery), zap.Int64("total", total), zap.Int("results_returned", len(vehicles)))
	}

	return vehicles, total, nil
}

func (r *VehicleRepository) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	query := `UPDATE vehicles SET deleted_at = NOW() WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

func (r *VehicleRepository) AssignDriver(ctx context.Context, vehicleID, tenantID, driverID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)

	var existingVehicleID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT id FROM vehicles WHERE current_driver_id = $1 AND deleted_at IS NULL AND id != $2`, driverID, vehicleID).Scan(&existingVehicleID)
	if err == nil {
		return fmt.Errorf("driver is already assigned to vehicle %s", existingVehicleID)
	}

	_, err = tx.Exec(ctx, `UPDATE vehicles SET current_driver_id = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`, driverID, vehicleID, tenantID)
	if err != nil { return err }

	return tx.Commit(ctx)
}

func (r *VehicleRepository) UnassignDriver(ctx context.Context, vehicleID, tenantID uuid.UUID) error {
	query := `UPDATE vehicles SET current_driver_id = NULL, updated_at = NOW() WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, vehicleID, tenantID)
	return err
}

func (r *VehicleRepository) GetStats(ctx context.Context) (*models.VehicleStats, error) {
	stats := &models.VehicleStats{}
	err := r.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'ACTIVE') as active,
			COUNT(*) FILTER (WHERE status = 'MAINTENANCE') as maintenance,
			COUNT(*) FILTER (WHERE status = 'SUSPENDED') as suspended
		FROM vehicles WHERE deleted_at IS NULL`).Scan(&stats.TotalVehicles, &stats.ActiveVehicles, &stats.Maintenance, &stats.Suspended)
	return stats, err
}

func (r *VehicleRepository) GetByRegistrationNumber(ctx context.Context, tenantID uuid.UUID, registrationNumber string) (*models.Vehicle, error) {
	v := &models.Vehicle{}
	err := r.db.QueryRow(ctx, `SELECT id, tenant_id, registration_number FROM vehicles WHERE tenant_id = $1 AND registration_number = $2 AND deleted_at IS NULL`, tenantID, registrationNumber).Scan(&v.ID, &v.TenantID, &v.RegistrationNumber)
	if err != nil {
		if err == pgx.ErrNoRows { return nil, nil }
		return nil, err
	}
	return v, nil
}

func (r *VehicleRepository) IsGPSAvailable(ctx context.Context, tenantID uuid.UUID, deviceID string, excludingVehicleID *uuid.UUID) (bool, string, error) {
	query := `SELECT id FROM vehicles WHERE gps_device_id = $1 AND deleted_at IS NULL`
	args := []interface{}{deviceID}
	if excludingVehicleID != nil {
		query += " AND id != $2"
		args = append(args, *excludingVehicleID)
	}

	var id uuid.UUID
	err := r.db.QueryRow(ctx, query, args...).Scan(&id)
	if err == pgx.ErrNoRows {
		return true, "", nil
	}
	if err != nil { return false, "", err }

	return false, fmt.Sprintf("GPS device already in use by vehicle %s", id), nil
}
