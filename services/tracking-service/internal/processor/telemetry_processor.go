package processor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/fleet-platform/shared/pkg/database"
	"github.com/flyfleet/tracking-service/internal/analytics"
	"github.com/flyfleet/tracking-service/internal/models"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type TelemetryProcessor struct {
	db        database.Database
	analytics *analytics.TelemetryAnalytics
	logger    *zap.Logger
}

func NewTelemetryProcessor(db database.Database, analytics *analytics.TelemetryAnalytics, logger *zap.Logger) *TelemetryProcessor {
	return &TelemetryProcessor{
		db:        db,
		analytics: analytics,
		logger:    logger,
	}
}

func (tp *TelemetryProcessor) HandleTelemetry(msg *nats.Msg) {
	// Identify data type from subject
	subject := msg.Subject
	switch {
	case subject == "telemetry.gps":
		tp.handleGPSTelemetry(msg)
	case subject == "telemetry.obd":
		tp.handleOBDTelemetry(msg)
	case subject == "telemetry.fuel":
		tp.handleFuelTelemetry(msg)
	case subject == "telemetry.heartbeat":
		tp.handleDeviceHeartbeat(msg)
	default:
		tp.logger.Warn("Unknown telemetry subject", zap.String("subject", subject))
	}
}

func (tp *TelemetryProcessor) handleGPSTelemetry(msg *nats.Msg) {
	var gpsData models.GPSTelemetry
	if err := json.Unmarshal(msg.Data, &gpsData); err != nil {
		tp.logger.Error("Failed to unmarshal GPS telemetry", zap.Error(err))
		return
	}

	// Store GPS data (RLS-aware)
	err := tp.storeGPSData(context.Background(), &gpsData)
	if err != nil {
		tp.logger.Error("Failed to store GPS data", zap.Error(err))
		return
	}

	// Update vehicle location
	err = tp.updateVehicleLocation(context.Background(), gpsData.OrgID.String(), gpsData.VehicleID.String(), gpsData.Latitude, gpsData.Longitude)
	if err != nil {
		tp.logger.Warn("Failed to update vehicle location", zap.Error(err))
	}

	// Trigger advanced analytics processing
	go tp.processAnalytics(gpsData.OrgID.String(), gpsData.VehicleID.String(), time.Now().Add(-24*time.Hour), time.Now())

	tp.logger.Debug("GPS data processed",
		zap.String("org_id", gpsData.OrgID.String()),
		zap.String("vehicle_id", gpsData.VehicleID.String()),
		zap.Float64("lat", gpsData.Latitude),
		zap.Float64("lng", gpsData.Longitude))
}

func (tp *TelemetryProcessor) handleOBDTelemetry(msg *nats.Msg) {
	var obdData models.OBDTelemetry
	if err := json.Unmarshal(msg.Data, &obdData); err != nil {
		tp.logger.Error("Failed to unmarshal OBD telemetry", zap.Error(err))
		return
	}

	// Store OBD data (RLS-aware)
	err := tp.storeOBDData(context.Background(), &obdData)
	if err != nil {
		tp.logger.Error("Failed to store OBD data", zap.Error(err))
		return
	}

	// Trigger advanced analytics processing
	go tp.processAnalytics(obdData.OrgID.String(), obdData.VehicleID.String(), time.Now().Add(-24*time.Hour), time.Now())
}

func (tp *TelemetryProcessor) handleFuelTelemetry(msg *nats.Msg) {
	var fuelData models.FuelTelemetry
	if err := json.Unmarshal(msg.Data, &fuelData); err != nil {
		tp.logger.Error("Failed to unmarshal fuel telemetry", zap.Error(err))
		return
	}

	// Store fuel data (RLS-aware)
	err := tp.storeFuelData(context.Background(), &fuelData)
	if err != nil {
		tp.logger.Error("Failed to store fuel data", zap.Error(err))
		return
	}

	// Trigger advanced analytics processing
	go tp.processAnalytics(fuelData.OrgID.String(), fuelData.VehicleID.String(), time.Now().Add(-24*time.Hour), time.Now())
}

func (tp *TelemetryProcessor) handleDeviceHeartbeat(msg *nats.Msg) {
	var heartbeat models.DeviceHeartbeat
	if err := json.Unmarshal(msg.Data, &heartbeat); err != nil {
		tp.logger.Error("Failed to unmarshal device heartbeat", zap.Error(err))
		return
	}

	// Store heartbeat (RLS-aware)
	err := tp.storeDeviceHeartbeat(context.Background(), &heartbeat)
	if err != nil {
		tp.logger.Error("Failed to store device heartbeat", zap.Error(err))
		return
	}
}

func (tp *TelemetryProcessor) storeGPSData(ctx context.Context, gps *models.GPSTelemetry) error {
	query := `
		INSERT INTO gps_locations (
			time, vehicle_id, org_id, latitude, longitude, altitude, speed, heading, 
			accuracy, ignition, odometer, battery_level, satellites, raw_data
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	return tp.db.WithTenant(ctx, gps.OrgID.String(), func(ctx context.Context, tx database.Transaction) error {
		return tx.Execute(ctx, query,
			gps.Time,
			gps.VehicleID,
			gps.OrgID,
			gps.Latitude,
			gps.Longitude,
			gps.Altitude,
			gps.Speed,
			gps.Heading,
			gps.Accuracy,
			gps.Ignition,
			gps.Odometer,
			gps.Battery,
			gps.Satellites,
			gps.RawData,
		)
	})
}

func (tp *TelemetryProcessor) storeOBDData(ctx context.Context, obd *models.OBDTelemetry) error {
	query := `
		INSERT INTO obd_readings (
			time, vehicle_id, org_id, rpm, coolant_temp, engine_load, throttle_pos, 
			fuel_level, dtcs, vin, raw_data
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	return tp.db.WithTenant(ctx, obd.OrgID.String(), func(ctx context.Context, tx database.Transaction) error {
		return tx.Execute(ctx, query,
			obd.Time,
			obd.VehicleID,
			obd.OrgID,
			obd.RPM,
			obd.CoolantTemp,
			obd.EngineLoad,
			obd.ThrottlePos,
			obd.FuelLevel,
			obd.DTCs,
			obd.VIN,
			obd.RawData,
		)
	})
}

func (tp *TelemetryProcessor) storeFuelData(ctx context.Context, fuel *models.FuelTelemetry) error {
	query := `
		INSERT INTO fuel_readings (
			time, vehicle_id, org_id, level_percent, level_liters, refuel_event, 
			theft_event, location, raw_data
		) VALUES ($1, $2, $3, $4, $5, $6, $7, ST_SetSRID(ST_MakePoint($8, $9), 4326), $10)`

	var lat, lng *float64
	if fuel.Latitude != nil && fuel.Longitude != nil {
		lat = fuel.Latitude
		lng = fuel.Longitude
	}

	return tp.db.WithTenant(ctx, fuel.OrgID.String(), func(ctx context.Context, tx database.Transaction) error {
		return tx.Execute(ctx, query,
			fuel.Time,
			fuel.VehicleID,
			fuel.OrgID,
			fuel.LevelPercent,
			fuel.LevelLiters,
			fuel.RefuelEvent,
			fuel.TheftEvent,
			lat, lng,
			fuel.RawData,
		)
	})
}

func (tp *TelemetryProcessor) storeDeviceHeartbeat(ctx context.Context, heartbeat *models.DeviceHeartbeat) error {
	query := `
		INSERT INTO device_heartbeats (
			device_id, vehicle_id, org_id, status, last_seen, battery_level, 
			signal_strength, error_codes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (device_id) 
		DO UPDATE SET 
			vehicle_id = EXCLUDED.vehicle_id,
			org_id = EXCLUDED.org_id,
			status = EXCLUDED.status,
			last_seen = EXCLUDED.last_seen,
			battery_level = EXCLUDED.battery_level,
			signal_strength = EXCLUDED.signal_strength,
			error_codes = EXCLUDED.error_codes`

	return tp.db.WithTenant(ctx, heartbeat.OrgID.String(), func(ctx context.Context, tx database.Transaction) error {
		return tx.Execute(ctx, query,
			heartbeat.DeviceID,
			heartbeat.VehicleID,
			heartbeat.OrgID,
			heartbeat.Status,
			heartbeat.LastSeen,
			heartbeat.Battery,
			heartbeat.SignalStrength,
			heartbeat.ErrorCodes,
		)
	})
}

func (tp *TelemetryProcessor) updateVehicleLocation(ctx context.Context, orgID, vehicleID string, lat, lng float64) error {
	query := `
		UPDATE vehicles 
		SET current_location = ST_SetSRID(ST_MakePoint($1, $2), 4326),
		    last_seen = NOW()
		WHERE id = $3`

	return tp.db.WithTenant(ctx, orgID, func(ctx context.Context, tx database.Transaction) error {
		return tx.Execute(ctx, query, lng, lat, vehicleID)
	})
}

func (tp *TelemetryProcessor) processAnalytics(orgID, vehicleIDStr string, startTime, endTime time.Time) {
	// Parse the vehicle ID string to UUID
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		// Log the error but continue
		tp.logger.Error("Failed to parse vehicle ID for analytics", zap.String("org_id", orgID), zap.String("vehicle_id", vehicleIDStr), zap.Error(err))
		return
	}

	// Process advanced analytics (passing context and OrgID)
	err = tp.analytics.ProcessAdvancedAnalytics(context.Background(), orgID, vehicleID, startTime, endTime)
	if err != nil {
		tp.logger.Error("Failed to process advanced analytics", zap.String("org_id", orgID), zap.String("vehicle_id", vehicleIDStr), zap.Error(err))
		return
	}

	// Log successful analytics processing
	tp.logger.Info("Advanced analytics processed",
		zap.String("org_id", orgID),
		zap.String("vehicle_id", vehicleIDStr),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime))
}