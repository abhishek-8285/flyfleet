package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fleet-platform/shared/pkg/cache"
	"github.com/fleet-platform/shared/pkg/database"
	"github.com/flyfleet/fleet-core-service/internal/domains/tracking/analytics"
	"github.com/flyfleet/fleet-core-service/internal/domains/tracking/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type TelemetryProcessor struct {
	db        database.Database
	redis     cache.RedisCacheInterface
	analytics *analytics.TelemetryAnalytics
	logger    *zap.Logger

	// Batching channels
	gpsChan       chan *models.GPSTelemetry
	obdChan       chan *models.OBDTelemetry
	fuelChan      chan *models.FuelTelemetry
	heartbeatChan chan *models.DeviceHeartbeat

	// Config
	batchSize     int
	flushInterval time.Duration
}

func NewTelemetryProcessor(db database.Database, red cache.RedisCacheInterface, analytics *analytics.TelemetryAnalytics, logger *zap.Logger) *TelemetryProcessor {
	tp := &TelemetryProcessor{
		db:            db,
		redis:         red,
		analytics:     analytics,
		logger:        logger,
		gpsChan:       make(chan *models.GPSTelemetry, 5000),
		obdChan:       make(chan *models.OBDTelemetry, 5000),
		fuelChan:      make(chan *models.FuelTelemetry, 2000),
		heartbeatChan: make(chan *models.DeviceHeartbeat, 1000),
		batchSize:     100, // Process 100 records at once
		flushInterval: 1 * time.Second,
	}

	// Start background batch workers
	go tp.gpsWorker()
	go tp.obdWorker()
	go tp.fuelWorker()
	go tp.heartbeatWorker()

	return tp
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

	// Queue for batch processing
	select {
	case tp.gpsChan <- &gpsData:
	default:
		tp.logger.Warn("GPS telemetry channel full, dropping message")
	}
}

func (tp *TelemetryProcessor) handleOBDTelemetry(msg *nats.Msg) {
	var obdData models.OBDTelemetry
	if err := json.Unmarshal(msg.Data, &obdData); err != nil {
		tp.logger.Error("Failed to unmarshal OBD telemetry", zap.Error(err))
		return
	}

	// Queue for batch processing
	select {
	case tp.obdChan <- &obdData:
	default:
		tp.logger.Warn("OBD telemetry channel full, dropping message")
	}
}

func (tp *TelemetryProcessor) handleFuelTelemetry(msg *nats.Msg) {
	var fuelData models.FuelTelemetry
	if err := json.Unmarshal(msg.Data, &fuelData); err != nil {
		tp.logger.Error("Failed to unmarshal fuel telemetry", zap.Error(err))
		return
	}

	// Queue for batch processing
	select {
	case tp.fuelChan <- &fuelData:
	default:
		tp.logger.Warn("Fuel telemetry channel full, dropping message")
	}
}

func (tp *TelemetryProcessor) handleDeviceHeartbeat(msg *nats.Msg) {
	var heartbeat models.DeviceHeartbeat
	if err := json.Unmarshal(msg.Data, &heartbeat); err != nil {
		tp.logger.Error("Failed to unmarshal device heartbeat", zap.Error(err))
		return
	}

	// Queue for batch processing
	select {
	case tp.heartbeatChan <- &heartbeat:
	default:
		tp.logger.Warn("Heartbeat telemetry channel full, dropping message")
	}
}

func (tp *TelemetryProcessor) gpsWorker() {
	buffer := make([]*models.GPSTelemetry, 0, tp.batchSize)
	ticker := time.NewTicker(tp.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case data := <-tp.gpsChan:
			buffer = append(buffer, data)
			if len(buffer) >= tp.batchSize {
				tp.flushGPS(buffer)
				buffer = make([]*models.GPSTelemetry, 0, tp.batchSize)
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				tp.flushGPS(buffer)
				buffer = make([]*models.GPSTelemetry, 0, tp.batchSize)
			}
		}
	}
}

func (tp *TelemetryProcessor) obdWorker() {
	buffer := make([]*models.OBDTelemetry, 0, tp.batchSize)
	ticker := time.NewTicker(tp.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case data := <-tp.obdChan:
			buffer = append(buffer, data)
			if len(buffer) >= tp.batchSize {
				tp.flushOBD(buffer)
				buffer = make([]*models.OBDTelemetry, 0, tp.batchSize)
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				tp.flushOBD(buffer)
				buffer = make([]*models.OBDTelemetry, 0, tp.batchSize)
			}
		}
	}
}

func (tp *TelemetryProcessor) fuelWorker() {
	buffer := make([]*models.FuelTelemetry, 0, tp.batchSize)
	ticker := time.NewTicker(tp.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case data := <-tp.fuelChan:
			buffer = append(buffer, data)
			if len(buffer) >= tp.batchSize {
				tp.flushFuel(buffer)
				buffer = make([]*models.FuelTelemetry, 0, tp.batchSize)
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				tp.flushFuel(buffer)
				buffer = make([]*models.FuelTelemetry, 0, tp.batchSize)
			}
		}
	}
}

func (tp *TelemetryProcessor) heartbeatWorker() {
	for data := range tp.heartbeatChan {
		err := tp.storeDeviceHeartbeat(context.Background(), data)
		if err == nil {
			tp.updateVehicleCache(data.VehicleID.String(), "heartbeat", data)
		} else {
			tp.logger.Error("Failed to store heartbeat", zap.Error(err))
		}
	}
}

func (tp *TelemetryProcessor) flushGPS(batch []*models.GPSTelemetry) {
	ctx := context.Background()
	byOrg := make(map[uuid.UUID][]*models.GPSTelemetry)
	for _, item := range batch {
		byOrg[item.OrgID] = append(byOrg[item.OrgID], item)
	}

	for orgID, items := range byOrg {
		err := tp.db.WithTenant(ctx, orgID.String(), func(ctx context.Context, tx database.Transaction) error {
			err := tp.storeGPSDataBatchTx(ctx, tx, items)
			if err != nil {
				tp.logger.Error("Failed to store GPS data in batch", zap.Error(err))
				return err
			}

			for _, item := range items {
				_ = tp.updateVehicleLocationTx(ctx, tx, item.VehicleID.String(), item.Latitude, item.Longitude)

				// Update redis cache
				tp.updateVehicleCache(item.VehicleID.String(), "gps", item)

				go tp.processAnalytics(orgID.String(), item.VehicleID.String(), time.Now().Add(-24*time.Hour), time.Now())
			}
			return nil
		})
		if err != nil {
			tp.logger.Error("Batch GPS flush failed for org", zap.String("org_id", orgID.String()), zap.Error(err))
		}
	}
}

func (tp *TelemetryProcessor) flushOBD(batch []*models.OBDTelemetry) {
	ctx := context.Background()
	byOrg := make(map[uuid.UUID][]*models.OBDTelemetry)
	for _, item := range batch {
		byOrg[item.OrgID] = append(byOrg[item.OrgID], item)
	}

	for orgID, items := range byOrg {
		_ = tp.db.WithTenant(ctx, orgID.String(), func(ctx context.Context, tx database.Transaction) error {
			_ = tp.storeOBDDataBatchTx(ctx, tx, items)
			for _, item := range items {
				// Update redis cache
				tp.updateVehicleCache(item.VehicleID.String(), "obd", item)

				go tp.processAnalytics(orgID.String(), item.VehicleID.String(), time.Now().Add(-24*time.Hour), time.Now())
			}
			return nil
		})
	}
}

func (tp *TelemetryProcessor) flushFuel(batch []*models.FuelTelemetry) {
	ctx := context.Background()
	byOrg := make(map[uuid.UUID][]*models.FuelTelemetry)
	for _, item := range batch {
		byOrg[item.OrgID] = append(byOrg[item.OrgID], item)
	}

	for orgID, items := range byOrg {
		_ = tp.db.WithTenant(ctx, orgID.String(), func(ctx context.Context, tx database.Transaction) error {
			_ = tp.storeFuelDataBatchTx(ctx, tx, items)
			for _, item := range items {
				// Update redis cache
				tp.updateVehicleCache(item.VehicleID.String(), "fuel", item)

				go tp.processAnalytics(orgID.String(), item.VehicleID.String(), time.Now().Add(-24*time.Hour), time.Now())
			}
			return nil
		})
	}
}

func (tp *TelemetryProcessor) storeGPSDataBatchTx(ctx context.Context, tx database.Transaction, items []*models.GPSTelemetry) error {
	if len(items) == 0 {
		return nil
	}
	rows := make([][]any, len(items))
	for i, gps := range items {
		rows[i] = []any{
			gps.Time, gps.VehicleID, gps.OrgID, gps.Latitude, gps.Longitude,
			gps.Altitude, gps.Speed, gps.Heading, gps.Accuracy, gps.Ignition,
			gps.Odometer, gps.Battery, gps.Satellites, gps.RawData,
		}
	}
	_, err := tx.CopyFrom(ctx,
		[]string{"gps_locations"},
		[]string{
			"time", "vehicle_id", "org_id", "latitude", "longitude",
			"altitude", "speed", "heading", "accuracy", "ignition",
			"odometer", "battery_level", "satellites", "raw_data",
		},
		pgx.CopyFromRows(rows),
	)
	return err
}

func (tp *TelemetryProcessor) storeOBDDataBatchTx(ctx context.Context, tx database.Transaction, items []*models.OBDTelemetry) error {
	if len(items) == 0 {
		return nil
	}
	rows := make([][]any, len(items))
	for i, obd := range items {
		rows[i] = []any{
			obd.Time, obd.VehicleID, obd.OrgID, obd.RPM, obd.CoolantTemp,
			obd.EngineLoad, obd.ThrottlePos, obd.FuelLevel, obd.DTCs, obd.VIN, obd.RawData,
		}
	}
	_, err := tx.CopyFrom(ctx,
		[]string{"obd_readings"},
		[]string{
			"time", "vehicle_id", "org_id", "rpm", "coolant_temp",
			"engine_load", "throttle_pos", "fuel_level", "dtcs", "vin", "raw_data",
		},
		pgx.CopyFromRows(rows),
	)
	return err
}

func (tp *TelemetryProcessor) storeFuelDataBatchTx(ctx context.Context, tx database.Transaction, items []*models.FuelTelemetry) error {
	if len(items) == 0 {
		return nil
	}
	rows := make([][]any, len(items))
	for i, fuel := range items {
		// PostGIS points handled differently in CopyFrom.
		// A common architecture for raw CopyFrom points is mapping the bytes or string.
		// If driver doesn't natively map ST_MakePoint, it may be easier to use Execute in batch or build WKB.
		// For simplicity, we drop to a prepared execution string for PostGIS if complex, or map Lat/Lng separately if schema allows.
		// Assuming the database schema has a standard geometry point type:
		var wkt string
		if fuel.Latitude != nil && fuel.Longitude != nil {
			wkt = fmt.Sprintf("SRID=4326;POINT(%f %f)", *fuel.Longitude, *fuel.Latitude)
		} else {
			wkt = ""
		}

		var loc any = nil
		if wkt != "" {
			loc = wkt
		}

		rows[i] = []any{
			fuel.Time, fuel.VehicleID, fuel.OrgID, fuel.LevelPercent,
			fuel.LevelLiters, fuel.RefuelEvent, fuel.TheftEvent, loc, fuel.RawData,
		}
	}
	_, err := tx.CopyFrom(ctx,
		[]string{"fuel_readings"},
		[]string{
			"time", "vehicle_id", "org_id", "level_percent", "level_liters",
			"refuel_event", "theft_event", "location", "raw_data",
		},
		pgx.CopyFromRows(rows),
	)
	return err
}

func (tp *TelemetryProcessor) updateVehicleLocationTx(ctx context.Context, tx database.Transaction, vehicleID string, lat, lng float64) error {
	query := `
		UPDATE vehicles 
		SET current_location = ST_SetSRID(ST_MakePoint($1, $2), 4326),
		    last_seen = NOW()
		WHERE id = $3`

	return tx.Execute(ctx, query, lng, lat, vehicleID)
}

func (tp *TelemetryProcessor) storeGPSData(ctx context.Context, gps *models.GPSTelemetry) error {
	return tp.db.WithTenant(ctx, gps.OrgID.String(), func(ctx context.Context, tx database.Transaction) error {
		return tp.storeGPSDataBatchTx(ctx, tx, []*models.GPSTelemetry{gps})
	})
}

func (tp *TelemetryProcessor) storeOBDData(ctx context.Context, obd *models.OBDTelemetry) error {
	return tp.db.WithTenant(ctx, obd.OrgID.String(), func(ctx context.Context, tx database.Transaction) error {
		return tp.storeOBDDataBatchTx(ctx, tx, []*models.OBDTelemetry{obd})
	})
}

func (tp *TelemetryProcessor) storeFuelData(ctx context.Context, fuel *models.FuelTelemetry) error {
	return tp.db.WithTenant(ctx, fuel.OrgID.String(), func(ctx context.Context, tx database.Transaction) error {
		return tp.storeFuelDataBatchTx(ctx, tx, []*models.FuelTelemetry{fuel})
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
		return tx.Execute(ctx, query, heartbeat.DeviceID, heartbeat.VehicleID, heartbeat.OrgID, heartbeat.Status, heartbeat.LastSeen, heartbeat.Battery, heartbeat.SignalStrength, heartbeat.ErrorCodes)
	})
}

func (tp *TelemetryProcessor) updateVehicleLocation(ctx context.Context, orgID, vehicleID string, lat, lng float64) error {
	return tp.db.WithTenant(ctx, orgID, func(ctx context.Context, tx database.Transaction) error {
		return tp.updateVehicleLocationTx(ctx, tx, vehicleID, lat, lng)
	})
}

func (tp *TelemetryProcessor) processAnalytics(orgID, vehicleIDStr string, startTime, endTime time.Time) {
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		tp.logger.Error("Failed to parse vehicle ID for analytics", zap.String("org_id", orgID), zap.String("vehicle_id", vehicleIDStr), zap.Error(err))
		return
	}

	err = tp.analytics.ProcessAdvancedAnalytics(context.Background(), orgID, vehicleID, startTime, endTime)
	if err != nil {
		tp.logger.Error("Failed to process advanced analytics", zap.String("org_id", orgID), zap.String("vehicle_id", vehicleIDStr), zap.Error(err))
		return
	}

	tp.logger.Debug("Advanced analytics processed",
		zap.String("org_id", orgID),
		zap.String("vehicle_id", vehicleIDStr),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime))
}

func (tp *TelemetryProcessor) updateVehicleCache(vehicleID, dataType string, data interface{}) {
	if tp.redis == nil {
		return
	}

	key := fmt.Sprintf("vehicle:%s:%s", vehicleID, dataType)
	cacheData, err := json.Marshal(data)
	if err != nil {
		tp.logger.Warn("Failed to marshal cache data", zap.Error(err))
		return
	}

	// Store for 24 hours as the "latest" status of this type
	err = tp.redis.Set(context.Background(), key, cacheData, 24*time.Hour)
	if err != nil {
		tp.logger.Warn("Failed to update vehicle cache", zap.String("key", key), zap.Error(err))
	}

	// Also update a general "latest" key
	latestKey := fmt.Sprintf("vehicle:%s:latest", vehicleID)
	_ = tp.redis.Set(context.Background(), latestKey, cacheData, 24*time.Hour)
}
