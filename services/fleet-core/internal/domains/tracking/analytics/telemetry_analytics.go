package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fleet-platform/shared/pkg/database"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TelemetryAnalytics struct {
	db     database.Database
	logger *zap.Logger
}

func NewTelemetryAnalytics(db database.Database, logger *zap.Logger) *TelemetryAnalytics {
	return &TelemetryAnalytics{
		db:     db,
		logger: logger,
	}
}

type SpeedAnalysis struct {
	AvgSpeed          float64 `json:"avg_speed"`
	MaxSpeed          float64 `json:"max_speed"`
	OverspeedEvents   int     `json:"overspeed_events"`
	OverspeedDuration float64 `json:"overspeed_duration_seconds"`
}

type HarshEventAnalysis struct {
	HarshAccelerations int     `json:"harsh_accelerations"`
	HarshBrakings      int     `json:"harsh_brakings"`
	HarshCornering     int     `json:"harsh_cornering"`
	AvgHarshRate       float64 `json:"avg_harsh_rate"`
}

type FuelEfficiencyAnalysis struct {
	AvgConsumptionPerHour float64 `json:"avg_consumption_per_hour"`
	AvgConsumptionPerKM   float64 `json:"avg_consumption_per_km"`
	IdleTimePercentage    float64 `json:"idle_time_percentage"`
	EfficiencyScore       float64 `json:"efficiency_score"`
	EstimatedSavings      float64 `json:"estimated_savings_usd"`
}

type Point struct {
	Latitude  float64   `json:"lat"`
	Longitude float64   `json:"lng"`
	Timestamp time.Time `json:"timestamp"`
}

type DistanceAnalysis struct {
	TotalDistanceKM     float64 `json:"total_distance_km"`
	RouteEfficiency     float64 `json:"route_efficiency_percent"`
	Stops               int     `json:"stops"`
	StopDurationSeconds int     `json:"stop_duration_seconds"`
	Locations           []Point `json:"locations,omitempty"`
}

type AnomalyResult struct {
	Type           string  `json:"type"`
	Data           any     `json:"data"`
	Confidence     float64 `json:"confidence"`
	Score          float64 `json:"score"`
	Recommendation string  `json:"recommendation"`
}

func (ta *TelemetryAnalytics) ProcessAdvancedAnalytics(ctx context.Context, tenantID string, vehicleID uuid.UUID, startTime, endTime time.Time) error {
	return ta.db.WithTenant(ctx, tenantID, func(ctx context.Context, tx database.Transaction) error {
		// Process speed analysis
		speedAnalysis, err := ta.analyzeSpeed(ctx, tx, vehicleID, startTime, endTime)
		if err != nil {
			ta.logger.Error("Failed to analyze speed", zap.String("tenant_id", tenantID), zap.Error(err))
		} else if speedAnalysis != nil {
			err = ta.saveAnalyticsResult(ctx, tx, vehicleID, "speed_analysis", speedAnalysis, 0.95)
			if err != nil {
				ta.logger.Error("Failed to save speed analysis", zap.String("tenant_id", tenantID), zap.Error(err))
			}
		}

		// Process harsh event analysis
		harshEventAnalysis, err := ta.analyzeHarshEvents(ctx, tx, vehicleID, startTime, endTime)
		if err != nil {
			ta.logger.Error("Failed to analyze harsh events", zap.String("tenant_id", tenantID), zap.Error(err))
		} else if harshEventAnalysis != nil {
			err = ta.saveAnalyticsResult(ctx, tx, vehicleID, "harsh_event_analysis", harshEventAnalysis, 0.90)
			if err != nil {
				ta.logger.Error("Failed to save harsh event analysis", zap.String("tenant_id", tenantID), zap.Error(err))
			}
		}

		// Process fuel efficiency analysis
		fuelEfficiencyAnalysis, err := ta.analyzeFuelEfficiency(ctx, tx, vehicleID, startTime, endTime)
		if err != nil {
			ta.logger.Error("Failed to analyze fuel efficiency", zap.String("tenant_id", tenantID), zap.Error(err))
		} else if fuelEfficiencyAnalysis != nil {
			err = ta.saveAnalyticsResult(ctx, tx, vehicleID, "fuel_efficiency_analysis", fuelEfficiencyAnalysis, 0.85)
			if err != nil {
				ta.logger.Error("Failed to save fuel efficiency analysis", zap.String("tenant_id", tenantID), zap.Error(err))
			}
		}

		// Process distance and route analysis
		distanceAnalysis, err := ta.analyzeDistanceAndRoute(ctx, tx, vehicleID, startTime, endTime)
		if err != nil {
			ta.logger.Error("Failed to analyze distance and route", zap.String("tenant_id", tenantID), zap.Error(err))
		} else if distanceAnalysis != nil {
			err = ta.saveAnalyticsResult(ctx, tx, vehicleID, "distance_analysis", distanceAnalysis, 0.92)
			if err != nil {
				ta.logger.Error("Failed to save distance analysis", zap.String("tenant_id", tenantID), zap.Error(err))
			}
		}

		// Process anomaly detection
		anomalyResults, err := ta.detectAnomalies(ctx, tx, vehicleID, startTime, endTime)
		if err != nil {
			ta.logger.Error("Failed to detect anomalies", zap.String("tenant_id", tenantID), zap.Error(err))
		} else {
			for _, anomaly := range anomalyResults {
				err = ta.saveAnomalyResult(ctx, tx, vehicleID, anomaly.Type, anomaly.Data, anomaly.Confidence, anomaly.Score, &anomaly.Recommendation)
				if err != nil {
					ta.logger.Error("Failed to save anomaly result", zap.String("tenant_id", tenantID), zap.Error(err))
				}
			}
		}

		return nil
	})
}

func (ta *TelemetryAnalytics) analyzeSpeed(ctx context.Context, tx database.Transaction, vehicleID uuid.UUID, startTime, endTime time.Time) (*SpeedAnalysis, error) {
	query := `
		SELECT 
			AVG(speed) as avg_speed,
			MAX(speed) as max_speed,
			COUNT(CASE WHEN speed > $3 THEN 1 END) as overspeed_events,
			SUM(CASE WHEN speed > $3 AND EXTRACT(EPOCH FROM (LEAD(time) OVER (ORDER BY time) - time)) < 300 THEN 
				EXTRACT(EPOCH FROM (LEAD(time) OVER (ORDER BY time) - time)) 
				ELSE 0 END) as overspeed_duration
		FROM gps_locations 
		WHERE vehicle_id = $1 
			AND time BETWEEN $2 AND $4
			AND speed IS NOT NULL
		ORDER BY time`

	var avgSpeed, maxSpeed, overspeedDuration sql.NullFloat64
	var overspeedEvents sql.NullInt64

	row := tx.QueryRow(ctx, query, vehicleID, startTime, 80.0, endTime)
	err := row.Scan(&avgSpeed, &maxSpeed, &overspeedEvents, &overspeedDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to scan speed analysis: %w", err)
	}

	analysis := &SpeedAnalysis{}
	if avgSpeed.Valid {
		analysis.AvgSpeed = avgSpeed.Float64
	}
	if maxSpeed.Valid {
		analysis.MaxSpeed = maxSpeed.Float64
	}
	if overspeedEvents.Valid {
		analysis.OverspeedEvents = int(overspeedEvents.Int64)
	}
	if overspeedDuration.Valid {
		analysis.OverspeedDuration = overspeedDuration.Float64
	}

	return analysis, nil
}

func (ta *TelemetryAnalytics) analyzeHarshEvents(ctx context.Context, tx database.Transaction, vehicleID uuid.UUID, startTime, endTime time.Time) (*HarshEventAnalysis, error) {
	query := `
		WITH speed_changes AS (
			SELECT 
				speed,
				time,
				LAG(speed) OVER (ORDER BY time) as prev_speed,
				LAG(time) OVER (ORDER BY time) as prev_time
			FROM gps_locations 
			WHERE vehicle_id = $1 
				AND time BETWEEN $2 AND $3
				AND speed IS NOT NULL
			ORDER BY time
		),
		calculated_changes AS (
			SELECT 
				speed,
				time,
				(speed - prev_speed) / EXTRACT(EPOCH FROM (time - prev_time)) * 3.6 as acceleration_rate
			FROM speed_changes
			WHERE prev_speed IS NOT NULL AND prev_time IS NOT NULL AND EXTRACT(EPOCH FROM (time - prev_time)) > 0
		)
		SELECT 
			COUNT(CASE WHEN acceleration_rate > 5.0 THEN 1 END) as harsh_accelerations,
			COUNT(CASE WHEN acceleration_rate < -5.0 THEN 1 END) as harsh_brakings,
			AVG(ABS(acceleration_rate)) FILTER (WHERE ABS(acceleration_rate) > 3.0) as avg_harsh_rate
		FROM calculated_changes`

	var harshAccels, harshBrakes sql.NullInt64
	var avgHarshRate sql.NullFloat64

	row := tx.QueryRow(ctx, query, vehicleID, startTime, endTime)
	err := row.Scan(&harshAccels, &harshBrakes, &avgHarshRate)
	if err != nil {
		return nil, fmt.Errorf("failed to scan harsh events: %w", err)
	}

	analysis := &HarshEventAnalysis{}
	if harshAccels.Valid {
		analysis.HarshAccelerations = int(harshAccels.Int64)
	}
	if harshBrakes.Valid {
		analysis.HarshBrakings = int(harshBrakes.Int64)
	}
	if avgHarshRate.Valid {
		analysis.AvgHarshRate = avgHarshRate.Float64
	}

	// Cornering
	cornering, _ := ta.countHarshCorneringEvents(ctx, tx, vehicleID, startTime, endTime)
	analysis.HarshCornering = cornering

	return analysis, nil
}

func (ta *TelemetryAnalytics) countHarshCorneringEvents(ctx context.Context, tx database.Transaction, vehicleID uuid.UUID, startTime, endTime time.Time) (int, error) {
	query := `
		WITH heading_changes AS (
			SELECT 
				heading,
				time,
				LAG(heading) OVER (ORDER BY time) as prev_heading,
				LAG(time) OVER (ORDER BY time) as prev_time
			FROM gps_locations 
			WHERE vehicle_id = $1 
				AND time BETWEEN $2 AND $3
				AND heading IS NOT NULL
				AND speed > 10
			ORDER BY time
		)
		SELECT COUNT(*) 
		FROM heading_changes 
		WHERE prev_heading IS NOT NULL 
			AND ABS(heading - prev_heading) > 45 
			AND EXTRACT(EPOCH FROM (time - prev_time)) < 5`

	var count int
	row := tx.QueryRow(ctx, query, vehicleID, startTime, endTime)
	err := row.Scan(&count)
	return count, err
}

func (ta *TelemetryAnalytics) analyzeFuelEfficiency(ctx context.Context, tx database.Transaction, vehicleID uuid.UUID, startTime, endTime time.Time) (*FuelEfficiencyAnalysis, error) {
	query := `
		SELECT 
			(MAX(fuel_level) - MIN(fuel_level)) as consumed,
			AVG(fuel_level) as avg_level
		FROM obd_readings 
		WHERE vehicle_id = $1 
			AND time BETWEEN $2 AND $3
			AND fuel_level IS NOT NULL`

	var consumed, avgLevel sql.NullFloat64
	row := tx.QueryRow(ctx, query, vehicleID, startTime, endTime)
	err := row.Scan(&consumed, &avgLevel)
	if err != nil {
		return nil, err
	}

	analysis := &FuelEfficiencyAnalysis{}
	if consumed.Valid {
		analysis.AvgConsumptionPerHour = consumed.Float64
	} // Placeholder logic

	return analysis, nil
}

func (ta *TelemetryAnalytics) analyzeDistanceAndRoute(ctx context.Context, tx database.Transaction, vehicleID uuid.UUID, startTime, endTime time.Time) (*DistanceAnalysis, error) {
	query := `
		SELECT 
			ST_Length(ST_MakeLine(location::geometry)::geography) / 1000.0 as distance_km
		FROM (
			SELECT location 
			FROM gps_locations 
			WHERE vehicle_id = $1 
				AND time BETWEEN $2 AND $3
			ORDER BY time
		) AS points`

	var distance sql.NullFloat64
	row := tx.QueryRow(ctx, query, vehicleID, startTime, endTime)
	err := row.Scan(&distance)
	if err != nil {
		return nil, err
	}

	analysis := &DistanceAnalysis{}
	if distance.Valid {
		analysis.TotalDistanceKM = distance.Float64
	}
	return analysis, nil
}

func (ta *TelemetryAnalytics) detectAnomalies(ctx context.Context, tx database.Transaction, vehicleID uuid.UUID, startTime, endTime time.Time) ([]AnomalyResult, error) {
	var results []AnomalyResult

	// Speed
	speedAnomalies, _ := ta.detectSpeedAnomalies(ctx, tx, vehicleID, startTime, endTime)
	results = append(results, speedAnomalies...)

	// Fuel
	fuelAnomalies, _ := ta.detectFuelAnomalies(ctx, tx, vehicleID, startTime, endTime)
	results = append(results, fuelAnomalies...)

	return results, nil
}

func (ta *TelemetryAnalytics) detectSpeedAnomalies(ctx context.Context, tx database.Transaction, vehicleID uuid.UUID, startTime, endTime time.Time) ([]AnomalyResult, error) {
	query := `SELECT speed, time FROM gps_locations WHERE vehicle_id = $1 AND time BETWEEN $2 AND $3 AND speed > 150`

	rows, err := tx.Query(ctx, query, vehicleID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var anomalies []AnomalyResult
	if len(rows) > 0 {
		anomalies = append(anomalies, AnomalyResult{
			Type:           "speeding",
			Confidence:     1.0,
			Recommendation: "Check driver compliance",
		})
	}
	return anomalies, nil
}

func (ta *TelemetryAnalytics) detectFuelAnomalies(ctx context.Context, tx database.Transaction, vehicleID uuid.UUID, startTime, endTime time.Time) ([]AnomalyResult, error) {
	// Basic fuel drop detection
	return nil, nil
}

func (ta *TelemetryAnalytics) detectLocationAnomalies(ctx context.Context, tx database.Transaction, vehicleID uuid.UUID, startTime, endTime time.Time) ([]AnomalyResult, error) {
	return nil, nil
}

func (ta *TelemetryAnalytics) saveAnalyticsResult(ctx context.Context, tx database.Transaction, vehicleID uuid.UUID, analysisType string, data any, confidence float64) error {
	query := `INSERT INTO telemetry_analytics (id, vehicle_id, analysis_type, data, timestamp, confidence) VALUES ($1, $2, $3, $4, $5, $6)`
	dataBytes, _ := json.Marshal(data)
	return tx.Execute(ctx, query, uuid.New(), vehicleID, analysisType, dataBytes, time.Now(), confidence)
}

func (ta *TelemetryAnalytics) saveAnomalyResult(ctx context.Context, tx database.Transaction, vehicleID uuid.UUID, analysisType string, data any, confidence, anomalyScore float64, recommendation *string) error {
	query := `INSERT INTO telemetry_anomalies (id, vehicle_id, anomaly_type, data, timestamp, confidence, anomaly_score, recommendation) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	dataBytes, _ := json.Marshal(data)
	return tx.Execute(ctx, query, uuid.New(), vehicleID, analysisType, dataBytes, time.Now(), confidence, anomalyScore, recommendation)
}
