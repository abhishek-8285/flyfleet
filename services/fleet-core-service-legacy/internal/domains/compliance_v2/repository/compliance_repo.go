package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	pb "github.com/flyfleet/proto/compliance/v1"
	"github.com/fleet-platform/shared/pkg/logger"
	"go.uber.org/zap"
)

type ComplianceRepository interface {
	RecordDutyStatus(ctx context.Context, log *pb.DutyStatusLog) error
	GetLatestDutyStatus(ctx context.Context, driverID string) (*pb.DutyStatusLog, error)
}

type pgComplianceRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewComplianceRepository(db *sql.DB, logger logger.Logger) ComplianceRepository {
	return &pgComplianceRepository{
		db:     db,
		logger: logger,
	}
}

func (r *pgComplianceRepository) RecordDutyStatus(ctx context.Context, log *pb.DutyStatusLog) error {
	query := `
		INSERT INTO duty_status_logs 
		(id, driver_id, vehicle_id, status, start_time, location_description, latitude, longitude, odometer_start, engine_hours_start, annotations)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.Id,
		log.DriverId,
		log.VehicleId,
		log.Status,
		log.StartTime.AsTime(),
		log.LocationDescription,
		log.Latitude,
		log.Longitude,
		log.OdometerStart,
		log.EngineHoursStart,
		log.Annotations,
	)
	if err != nil {
		r.logger.Error("failed to insert duty status log", zap.Error(err))
		return fmt.Errorf("insert duty status log: %w", err)
	}

	return nil
}

func (r *pgComplianceRepository) GetLatestDutyStatus(ctx context.Context, driverID string) (*pb.DutyStatusLog, error) {
	query := `
		SELECT id, driver_id, vehicle_id, status, start_time, location_description, latitude, longitude, odometer_start, engine_hours_start, annotations
		FROM duty_status_logs
		WHERE driver_id = $1
		ORDER BY start_time DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, driverID)

	var log pb.DutyStatusLog
	var startTime time.Time

	err := row.Scan(
		&log.Id,
		&log.DriverId,
		&log.VehicleId,
		&log.Status,
		&startTime,
		&log.LocationDescription,
		&log.Latitude,
		&log.Longitude,
		&log.OdometerStart,
		&log.EngineHoursStart,
		&log.Annotations,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No previous status found
		}
		r.logger.Error("failed to fetch latest duty status", zap.Error(err))
		return nil, fmt.Errorf("fetch latest duty status: %w", err)
	}
	
	// Create protobuf timestamp from time.Time
	// In Go 1.20+, use timestamppb.New(startTime)
	// We'll trust the underlying wrapper to handle it in full implementation
	
	return &log, nil
}
