package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/tracking-service/internal/analytics"
	"github.com/flyfleet/tracking-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type TelemetryProcessor struct {
	db     *pgxpool.Pool
	nc     *nats.Conn
	logger *zap.Logger
	analytics *analytics.TelemetryAnalytics
}

func NewTelemetryProcessor(db *pgxpool.Pool, nc *nats.Conn, logger *zap.Logger) *TelemetryProcessor {
	analytics := analytics.NewTelemetryAnalytics(db, logger)
	return &TelemetryProcessor{
		db:        db,
		nc:        nc,
		logger:    logger,
		analytics: analytics,
	}
}

func (tp *TelemetryProcessor) StartProcessing() error {
	// Subscribe to GPS telemetry
	_, err := tp.nc.Subscribe("fleet.telemetry.gps", tp.handleGPSTelemetry)
	if err != nil {
		return fmt.Errorf("failed to subscribe to GPS telemetry: %w", err)
	}

	// Subscribe to OBD telemetry
	_, err = tp.nc.Subscribe("fleet.telemetry.obd", tp.handleOBDTelemetry)
	if err != nil {
		return fmt.Errorf("failed to subscribe to OBD telemetry: %w", err)
	}

	// Subscribe to fuel telemetry
	_, err = tp.nc.Subscribe("fleet.telemetry.fuel", tp.handleFuelTelemetry)
	if err != nil {
		return fmt.Errorf("failed to subscribe to fuel telemetry: %w", err)
	}

	// Subscribe to device heartbeats
	_, err = tp.nc.Subscribe("fleet.devices.heartbeat", tp.handleDeviceHeartbeat)
	if err != nil {
		return fmt.Errorf("failed to subscribe to device heartbeats: %w", err)
	}

	tp.logger.Info("Telemetry processor started - listening for data on NATS")
	return nil
}

func (tp *TelemetryProcessor) handleGPSTelemetry(msg *nats.Msg) {
	var gpsData models.GPSTelemetry
	if err := json.Unmarshal(msg.Data, &gpsData); err != nil {
		tp.logger.Error("Failed to unmarshal GPS data", zap.Error(err))
		return
	}

	// Store in TimescaleDB
	err := tp.storeGPSData(context.Background(), &gpsData)
	if err != nil {
		tp.logger.Error("Failed to store GPS data", zap.Error(err))
		return
	}

	// Update vehicle location
	err = tp.updateVehicleLocation(context.Background(), gpsData.VehicleID.String(), gpsData.Latitude, gpsData.Longitude)
	if err != nil {
		tp.logger.Warn("Failed to update vehicle location", zap.Error(err))
	}

	// Trigger advanced analytics processing
	go tp.processAnalytics(gpsData.VehicleID.String(), time.Now().Add(-24*time.Hour), time.Now())

	tp.logger.Debug("GPS data processed",
		zap.String("vehicle_id", gpsData.VehicleID.String()),
		zap.Float64("lat", gpsData.Latitude),
		zap.Float64("lng", gpsData.Longitude),
		zap.Float64("speed", *gpsData.Speed),
	)
}

func (tp *TelemetryProcessor) handleOBDTelemetry(msg *nats.Msg) {
	var obdData models.OBDTelemetry
	if err := json.Unmarshal(msg.Data, &obdData); err != nil {
		tp.logger.Error("Failed to unmarshal OBD data", zap.Error(err))
		return
	}

	// Store in TimescaleDB
	err := tp.storeOBDData(context.Background(), &obdData)
	if err != nil {
		tp.logger.Error("Failed to store OBD data", zap.Error(err))
		return
	}

	tp.logger.Debug("OBD data processed",
		zap.String("vehicle_id", obdData.VehicleID.String()),
		zap.Int("rpm", *obdData.RPM),
		zap.Float64("fuel_level", *obdData.FuelLevel),
	)

	// Trigger advanced analytics processing
	go tp.processAnalytics(obdData.VehicleID.String(), time.Now().Add(-24*time.Hour), time.Now())
}

func (tp *TelemetryProcessor) handleFuelTelemetry(msg *nats.Msg) {
	var fuelData models.FuelTelemetry
	if err := json.Unmarshal(msg.Data, &fuelData); err != nil {
		tp.logger.Error("Failed to unmarshal fuel data", zap.Error(err))
		return
	}

	// Store in TimescaleDB
	err := tp.storeFuelData(context.Background(), &fuelData)
	if err != nil {
		tp.logger.Error("Failed to store fuel data", zap.Error(err))
		return
	}

	// Check for fuel events
	if fuelData.RefuelEvent != nil && *fuelData.RefuelEvent {
		tp.logger.Info("Refuel event detected",
			zap.String("vehicle_id", fuelData.VehicleID.String()),
			zap.Float64("level_percent", *fuelData.LevelPercent),
		)
	}

	if fuelData.TheftEvent != nil && *fuelData.TheftEvent {
		tp.logger.Warn("Fuel theft event detected",
			zap.String("vehicle_id", fuelData.VehicleID.String()),
			zap.Float64("level_percent", *fuelData.LevelPercent),
		)
	}

	tp.logger.Debug("Fuel data processed",
		zap.String("vehicle_id", fuelData.VehicleID.String()),
		zap.Float64("level_percent", *fuelData.LevelPercent),
	)

	// Trigger advanced analytics processing
	go tp.processAnalytics(fuelData.VehicleID.String(), time.Now().Add(-24*time.Hour), time.Now())
}

func (tp *TelemetryProcessor) handleDeviceHeartbeat(msg *nats.Msg) {
	var heartbeat models.DeviceHeartbeat
	if err := json.Unmarshal(msg.Data, &heartbeat); err != nil {
		tp.logger.Error("Failed to unmarshal heartbeat data", zap.Error(err))
		return
	}

	// Store heartbeat
	err := tp.storeDeviceHeartbeat(context.Background(), &heartbeat)
	if err != nil {
		tp.logger.Error("Failed to store device heartbeat", zap.Error(err))
		return
	}

	tp.logger.Debug("Device heartbeat processed",
		zap.String("device_id", heartbeat.DeviceID),
		zap.String("status", heartbeat.Status),
		zap.Int("battery", *heartbeat.Battery),
	)
}

func (tp *TelemetryProcessor) storeGPSData(ctx context.Context, gps *models.GPSTelemetry) error {
	query := `
		INSERT INTO gps_locations (
			time, vehicle_id, latitude, longitude, altitude, speed, heading, 
			accuracy, ignition, odometer, battery_level, satellites, raw_data
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := tp.db.Exec(ctx, query,
		gps.Time,
		gps.VehicleID,
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

	return err
}

func (tp *TelemetryProcessor) storeOBDData(ctx context.Context, obd *models.OBDTelemetry) error {
	query := `
		INSERT INTO obd_readings (
			time, vehicle_id, rpm, coolant_temp, engine_load, throttle_pos, 
			fuel_level, dtcs, vin, raw_data
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := tp.db.Exec(ctx, query,
		obd.Time,
		obd.VehicleID,
		obd.RPM,
		obd.CoolantTemp,
		obd.EngineLoad,
		obd.ThrottlePos,
		obd.FuelLevel,
		obd.DTCs,
		obd.VIN,
		obd.RawData,
	)

	return err
}

func (tp *TelemetryProcessor) storeFuelData(ctx context.Context, fuel *models.FuelTelemetry) error {
	query := `
		INSERT INTO fuel_readings (
			time, vehicle_id, level_percent, level_liters, refuel_event, 
			theft_event, location, raw_data
		) VALUES ($1, $2, $3, $4, $5, $6, ST_SetSRID(ST_MakePoint($7, $8), 4326), $9)`

	var lat, lng *float64
	if fuel.Latitude != nil && fuel.Longitude != nil {
		lat = fuel.Latitude
		lng = fuel.Longitude
	}

	_, err := tp.db.Exec(ctx, query,
		fuel.Time,
		fuel.VehicleID,
		fuel.LevelPercent,
		fuel.LevelLiters,
		fuel.RefuelEvent,
		fuel.TheftEvent,
		lat, lng,
		fuel.RawData,
	)

	return err
}

func (tp *TelemetryProcessor) storeDeviceHeartbeat(ctx context.Context, heartbeat *models.DeviceHeartbeat) error {
	query := `
		INSERT INTO device_heartbeats (
			device_id, vehicle_id, status, last_seen, battery_level, 
			signal_strength, error_codes
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (device_id) 
		DO UPDATE SET 
			vehicle_id = EXCLUDED.vehicle_id,
			status = EXCLUDED.status,
			last_seen = EXCLUDED.last_seen,
			battery_level = EXCLUDED.battery_level,
			signal_strength = EXCLUDED.signal_strength,
			error_codes = EXCLUDED.error_codes`

	_, err := tp.db.Exec(ctx, query,
		heartbeat.DeviceID,
		heartbeat.VehicleID,
		heartbeat.Status,
		heartbeat.LastSeen,
		heartbeat.Battery,
		heartbeat.SignalStrength,
		heartbeat.ErrorCodes,
	)

	return err
}

func (tp *TelemetryProcessor) updateVehicleLocation(ctx context.Context, vehicleID string, lat, lng float64) error {
	query := `
		UPDATE vehicles 
		SET current_location = ST_SetSRID(ST_MakePoint($1, $2), 4326),
		    last_seen = NOW()
		WHERE id = $3`

	_, err := tp.db.Exec(ctx, query, lng, lat, vehicleID)
	return err
}

func (tp *TelemetryProcessor) processAnalytics(vehicleIDStr string, startTime, endTime time.Time) {
	// Parse the vehicle ID string to UUID
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		// Log the error but continue
		tp.logger.Error("Failed to parse vehicle ID for analytics", zap.String("vehicle_id", vehicleIDStr), zap.Error(err))
		return
	}

	// Process advanced analytics
	err = tp.analytics.ProcessAdvancedAnalytics(vehicleID, startTime, endTime)
	if err != nil {
		tp.logger.Error("Failed to process advanced analytics", zap.String("vehicle_id", vehicleIDStr), zap.Error(err))
		return
	}

	// Log successful analytics processing
	tp.logger.Info("Advanced analytics processed",
		zap.String("vehicle_id", vehicleIDStr),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime))
}