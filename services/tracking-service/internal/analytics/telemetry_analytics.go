package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)
type TelemetryAnalytics struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

type AnalyticsResult struct {
	ID             string    `json:"id"`
	VehicleID      string    `json:"vehicle_id"`
	AnalysisType   string    `json:"analysis_type"`
	Data           any       `json:"data"`
	Timestamp      time.Time `json:"timestamp"`
	Confidence     float64   `json:"confidence"`
	AnomalyScore   *float64  `json:"anomaly_score,omitempty"`
	Recommendation *string   `json:"recommendation,omitempty"`
}

type SpeedAnalysis struct {
	AverageSpeed      float64 `json:"average_speed"`
	MaxSpeed          float64 `json:"max_speed"`
	OverspeedEvents   int     `json:"overspeed_events"`
	SpeedLimit        float64 `json:"speed_limit"`
	OverspeedDuration int     `json:"overspeed_duration_seconds"`
}

type HarshEventAnalysis struct {
	HarshAccelerations  int     `json:"harsh_accelerations"`
	HarshBrakings       int     `json:"harsh_brakings"`
	HarshCornerings     int     `json:"harsh_cornerings"`
	AvgAccelerationRate float64 `json:"avg_acceleration_rate"`
	AvgDecelerationRate float64 `json:"avg_deceleration_rate"`
}

type FuelEfficiencyAnalysis struct {
	AvgConsumptionPerHour float64 `json:"avg_consumption_per_hour"`
	AvgConsumptionPerKM   float64 `json:"avg_consumption_per_km"`
	IdleTimePercentage    float64 `json:"idle_time_percentage"`
	EfficiencyScore       float64 `json:"efficiency_score"`
	EstimatedSavings      float64 `json:"estimated_savings"`
}

type DistanceAnalysis struct {
	TotalDistanceKM     float64   `json:"total_distance_km"`
	RouteEfficiency     float64   `json:"route_efficiency"`
	Stops               int       `json:"stops"`
	StopDurationSeconds int       `json:"stop_duration_seconds"`
	Locations           []Point   `json:"locations"`
}

type Point struct {
	Latitude  float64   `json:"lat"`
	Longitude float64   `json:"lng"`
	Timestamp time.Time `json:"timestamp"`
}

func NewTelemetryAnalytics(db *pgxpool.Pool, logger *zap.Logger) *TelemetryAnalytics {
	return &TelemetryAnalytics{
		db:     db,
		logger: logger,
	}
}

func (ta *TelemetryAnalytics) ProcessAdvancedAnalytics(vehicleID uuid.UUID, startTime, endTime time.Time) error {
	// Process speed analysis
	speedAnalysis, err := ta.analyzeSpeed(vehicleID, startTime, endTime)
	if err != nil {
		ta.logger.Error("Failed to analyze speed", zap.Error(err))
	} else if speedAnalysis != nil {
		err = ta.saveAnalyticsResult(vehicleID, "speed_analysis", speedAnalysis, 0.95)
		if err != nil {
			ta.logger.Error("Failed to save speed analysis", zap.Error(err))
		}
	}

	// Process harsh event analysis
	harshEventAnalysis, err := ta.analyzeHarshEvents(vehicleID, startTime, endTime)
	if err != nil {
		ta.logger.Error("Failed to analyze harsh events", zap.Error(err))
	} else if harshEventAnalysis != nil {
		err = ta.saveAnalyticsResult(vehicleID, "harsh_event_analysis", harshEventAnalysis, 0.90)
		if err != nil {
			ta.logger.Error("Failed to save harsh event analysis", zap.Error(err))
		}
	}

	// Process fuel efficiency analysis
	fuelEfficiencyAnalysis, err := ta.analyzeFuelEfficiency(vehicleID, startTime, endTime)
	if err != nil {
		ta.logger.Error("Failed to analyze fuel efficiency", zap.Error(err))
	} else if fuelEfficiencyAnalysis != nil {
		err = ta.saveAnalyticsResult(vehicleID, "fuel_efficiency_analysis", fuelEfficiencyAnalysis, 0.85)
		if err != nil {
			ta.logger.Error("Failed to save fuel efficiency analysis", zap.Error(err))
		}
	}

	// Process distance and route analysis
	distanceAnalysis, err := ta.analyzeDistanceAndRoute(vehicleID, startTime, endTime)
	if err != nil {
		ta.logger.Error("Failed to analyze distance and route", zap.Error(err))
	} else if distanceAnalysis != nil {
		err = ta.saveAnalyticsResult(vehicleID, "distance_analysis", distanceAnalysis, 0.92)
		if err != nil {
			ta.logger.Error("Failed to save distance analysis", zap.Error(err))
		}
	}

	// Process anomaly detection
	anomalyResults, err := ta.detectAnomalies(vehicleID, startTime, endTime)
	if err != nil {
		ta.logger.Error("Failed to detect anomalies", zap.Error(err))
	} else {
		for _, anomaly := range anomalyResults {
			err = ta.saveAnomalyResult(vehicleID, anomaly.Type, anomaly.Data, anomaly.Confidence, anomaly.Score, &anomaly.Recommendation)
			if err != nil {
				ta.logger.Error("Failed to save anomaly result", zap.Error(err))
			}
		}
	}

	return nil
}

func (ta *TelemetryAnalytics) analyzeSpeed(vehicleID uuid.UUID, startTime, endTime time.Time) (*SpeedAnalysis, error) {
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

	row := ta.db.QueryRow(context.Background(), query, vehicleID, startTime, 80.0, endTime) // Assuming 80 km/h speed limit
	err := row.Scan(&avgSpeed, &maxSpeed, &overspeedEvents, &overspeedDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to scan speed analysis: %w", err)
	}

	speedAnalysis := &SpeedAnalysis{
		AverageSpeed:      0,
		MaxSpeed:          0,
		OverspeedEvents:   0,
		SpeedLimit:        80.0,
		OverspeedDuration: 0,
	}

	if avgSpeed.Valid {
		speedAnalysis.AverageSpeed = avgSpeed.Float64
	}
	if maxSpeed.Valid {
		speedAnalysis.MaxSpeed = maxSpeed.Float64
	}
	if overspeedEvents.Valid {
		speedAnalysis.OverspeedEvents = int(overspeedEvents.Int64)
	}
	if overspeedDuration.Valid {
		speedAnalysis.OverspeedDuration = int(overspeedDuration.Float64)
	}

	return speedAnalysis, nil
}

func (ta *TelemetryAnalytics) analyzeHarshEvents(vehicleID uuid.UUID, startTime, endTime time.Time) (*HarshEventAnalysis, error) {
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
				prev_speed,
				prev_time,
				CASE 
					WHEN prev_speed IS NOT NULL AND prev_time IS NOT NULL THEN
						EXTRACT(EPOCH FROM (time - prev_time))
					ELSE 0 
				END as time_diff,
				CASE 
					WHEN prev_speed IS NOT NULL AND prev_time IS NOT NULL THEN
						(speed - prev_speed) / EXTRACT(EPOCH FROM (time - prev_time)) * 3.6  -- Convert to km/h/s
					ELSE 0 
				END as acceleration_rate
			FROM speed_changes
			WHERE prev_speed IS NOT NULL AND prev_time IS NOT NULL AND time_diff > 0
		)
		SELECT 
			COUNT(CASE WHEN acceleration_rate > 5.0 THEN 1 END) as harsh_accelerations,
			COUNT(CASE WHEN acceleration_rate < -5.0 THEN 1 END) as harsh_brakings,
			AVG(ABS(acceleration_rate)) FILTER (WHERE ABS(acceleration_rate) > 3.0) as avg_harsh_rate
		FROM calculated_changes`

	var harshAccels, harshBrakes sql.NullInt64
	var avgHarshRate sql.NullFloat64

	row := ta.db.QueryRow(context.Background(), query, vehicleID, startTime, endTime)
	err := row.Scan(&harshAccels, &harshBrakes, &avgHarshRate)
	if err != nil {
		return nil, fmt.Errorf("failed to scan harsh event analysis: %w", err)
	}

	harshAnalysis := &HarshEventAnalysis{
		HarshAccelerations:  0,
		HarshBrakings:       0,
		HarshCornerings:     0, // Cornering requires heading data analysis
		AvgAccelerationRate: 0,
		AvgDecelerationRate: 0,
	}

	if harshAccels.Valid {
		harshAnalysis.HarshAccelerations = int(harshAccels.Int64)
	}
	if harshBrakes.Valid {
		harshAnalysis.HarshBrakings = int(harshBrakes.Int64)
	}
	if avgHarshRate.Valid {
		harshAnalysis.AvgAccelerationRate = math.Abs(avgHarshRate.Float64)
		harshAnalysis.AvgDecelerationRate = -math.Abs(avgHarshRate.Float64)
	}

	// Analyze cornering separately if needed
	corneringEvents, err := ta.countHarshCorneringEvents(vehicleID, startTime, endTime)
	if err != nil {
		ta.logger.Warn("Failed to count harsh cornering events", zap.Error(err))
	} else {
		harshAnalysis.HarshCornerings = corneringEvents
	}

	return harshAnalysis, nil
}

func (ta *TelemetryAnalytics) countHarshCorneringEvents(vehicleID uuid.UUID, startTime, endTime time.Time) (int, error) {
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
				AND speed > 10  -- Only when moving
			ORDER BY time
		),
		calculated_turns AS (
			SELECT 
				time,
				prev_time,
				CASE 
					WHEN prev_heading IS NOT NULL AND prev_time IS NOT NULL THEN
						ABS(heading - prev_heading)  -- Simplified angle difference
					ELSE 0 
				END as heading_change,
				CASE 
					WHEN prev_heading IS NOT NULL AND prev_time IS NOT NULL THEN
						EXTRACT(EPOCH FROM (time - prev_time))
					ELSE 0 
				END as time_diff
			FROM heading_changes
			WHERE prev_heading IS NOT NULL AND prev_time IS NOT NULL AND time_diff > 0
		)
		SELECT COUNT(*) as harsh_cornering_events
		FROM calculated_turns
		WHERE heading_change > 45  -- Significant turn (>45 degrees)
			AND time_diff < 5  -- Happened quickly (<5 seconds)
	`

	var harshCorneringEvents sql.NullInt64
	row := ta.db.QueryRow(context.Background(), query, vehicleID, startTime, endTime)
	err := row.Scan(&harshCorneringEvents)
	if err != nil {
		return 0, fmt.Errorf("failed to count harsh cornering events: %w", err)
	}

	if harshCorneringEvents.Valid {
		return int(harshCorneringEvents.Int64), nil
	}
	return 0, nil
}

func (ta *TelemetryAnalytics) analyzeFuelEfficiency(vehicleID uuid.UUID, startTime, endTime time.Time) (*FuelEfficiencyAnalysis, error) {
	query := `
		WITH location_data AS (
			SELECT 
				time,
				speed,
				odometer,
				LAG(odometer) OVER (ORDER BY time) as prev_odometer,
				LAG(time) OVER (ORDER BY time) as prev_time
			FROM gps_locations 
			WHERE vehicle_id = $1 
				AND time BETWEEN $2 AND $3
				AND odometer IS NOT NULL
			ORDER BY time
		),
		trip_segments AS (
			SELECT 
				time,
				speed,
				odometer,
				prev_odometer,
				prev_time,
				CASE 
					WHEN prev_odometer IS NOT NULL THEN odometer - prev_odometer
					ELSE 0
				END as distance_km,
				CASE 
					WHEN prev_time IS NOT NULL THEN EXTRACT(EPOCH FROM (time - prev_time))/3600.0
					ELSE 0
				END as duration_hours
			FROM location_data
			WHERE prev_odometer IS NOT NULL AND prev_time IS NOT NULL
		),
		fuel_data AS (
			SELECT 
				time,
				level_liters,
				LAG(level_liters) OVER (ORDER BY time) as prev_fuel_level
			FROM fuel_readings
			WHERE vehicle_id = $1 
				AND time BETWEEN $2 AND $3
		),
		total_fuel_consumed AS (
			SELECT SUM(prev_fuel_level - level_liters) as consumed
			FROM fuel_data
			WHERE prev_fuel_level IS NOT NULL AND level_liters < prev_fuel_level
		)
		SELECT 
			SUM(distance_km) as total_distance,
			SUM(duration_hours) as total_duration,
			SUM(CASE WHEN speed < 5 THEN duration_hours ELSE 0 END) as idle_duration,
			(SELECT consumed FROM total_fuel_consumed LIMIT 1) as total_fuel_consumed
		FROM trip_segments`

	var totalDistance, totalDuration, idleDuration, totalFuelConsumed sql.NullFloat64

	row := ta.db.QueryRow(context.Background(), query, vehicleID, startTime, endTime)
	err := row.Scan(&totalDistance, &totalDuration, &idleDuration, &totalFuelConsumed)
	if err != nil {
		return nil, fmt.Errorf("failed to scan fuel efficiency analysis: %w", err)
	}

	fuelAnalysis := &FuelEfficiencyAnalysis{
		AvgConsumptionPerHour: 0,
		AvgConsumptionPerKM:   0,
		IdleTimePercentage:    0,
		EfficiencyScore:       0,
		EstimatedSavings:      0,
	}

	totalDist := 0.0
	if totalDistance.Valid && totalDistance.Float64 > 0 {
		totalDist = totalDistance.Float64
		fuelAnalysis.AvgConsumptionPerKM = totalFuelConsumed.Float64 / totalDist
	}

	totalDur := 0.0
	if totalDuration.Valid && totalDuration.Float64 > 0 {
		totalDur = totalDuration.Float64
		fuelAnalysis.AvgConsumptionPerHour = totalFuelConsumed.Float64 / totalDur
	}

	if totalDur > 0 {
		fuelAnalysis.IdleTimePercentage = (idleDuration.Float64 / totalDur) * 100
	}

	// Efficiency score calculation (higher is better)
	fuelAnalysis.EfficiencyScore = calculateEfficiencyScore(totalFuelConsumed.Float64, totalDist, idleDuration.Float64, totalDur)

	return fuelAnalysis, nil
}

func calculateEfficiencyScore(fuelConsumed, distanceKM, idleHours, totalHours float64) float64 {
	if distanceKM <= 0 || fuelConsumed <= 0 {
		return 100 // Perfect score if no movement or no fuel used
	}

	// Base score based on fuel efficiency (km/L)
	efficiencyKMPerL := distanceKM / fuelConsumed
	baseScore := math.Min(efficiencyKMPerL*10, 100) // Scale to 0-100

	// Penalty for idle time
	idleRatio := 0.0
	if totalHours > 0 {
		idleRatio = idleHours / totalHours
	}
	idlePenalty := idleRatio * 30 // Up to 30 point penalty for idling

	score := baseScore - idlePenalty
	return math.Max(score, 0) // Ensure non-negative score
}

func (ta *TelemetryAnalytics) analyzeDistanceAndRoute(vehicleID uuid.UUID, startTime, endTime time.Time) (*DistanceAnalysis, error) {
	query := `
		WITH ordered_locations AS (
			SELECT 
				latitude,
				longitude,
				time,
				odometer,
				LAG(odometer) OVER (ORDER BY time) as prev_odometer,
				LAG(time) OVER (ORDER BY time) as prev_time
			FROM gps_locations 
			WHERE vehicle_id = $1 
				AND time BETWEEN $2 AND $3
				AND latitude IS NOT NULL
				AND longitude IS NOT NULL
			ORDER BY time
		),
		segments AS (
			SELECT 
				latitude,
				longitude,
				time,
				CASE 
					WHEN prev_odometer IS NOT NULL AND odometer IS NOT NULL 
					THEN odometer - prev_odometer
					ELSE 0
				END as segment_distance,
				CASE 
					WHEN prev_time IS NOT NULL 
					THEN EXTRACT(EPOCH FROM (time - prev_time))
					ELSE 0
				END as segment_duration
			FROM ordered_locations
			WHERE prev_odometer IS NOT NULL AND prev_time IS NOT NULL
		)
		SELECT 
			SUM(segment_distance) as total_distance,
			COUNT(CASE WHEN segment_distance < 0.1 AND segment_duration > 300 THEN 1 END) as stops,
			SUM(CASE WHEN segment_distance < 0.1 AND segment_duration > 300 THEN segment_duration ELSE 0 END) as stop_duration,
			ARRAY_AGG(latitude ORDER BY time) as lats,
			ARRAY_AGG(longitude ORDER BY time) as lngs,
			ARRAY_AGG(time ORDER BY time) as times
		FROM segments`

	var totalDistance, stopDuration sql.NullFloat64
	var lats, lngs []float64
	var times []time.Time

	row := ta.db.QueryRow(context.Background(), query, vehicleID, startTime, endTime)
	err := row.Scan(&totalDistance, &stopDuration, &stopDuration, &lats, &lngs, &times)
	if err != nil {
		return nil, fmt.Errorf("failed to scan distance analysis: %w", err)
	}

	// Prepare locations array
	var locations []Point
	if len(lats) == len(lngs) && len(lats) == len(times) {
		locations = make([]Point, len(lats))
		for i := range lats {
			locations[i] = Point{
				Latitude:  lats[i],
				Longitude: lngs[i],
				Timestamp: times[i],
			}
		}
	}

	distanceAnalysis := &DistanceAnalysis{
		TotalDistanceKM:     0,
		RouteEfficiency:     0,
		Stops:               0,
		StopDurationSeconds: 0,
		Locations:           locations,
	}

	if totalDistance.Valid {
		distanceAnalysis.TotalDistanceKM = totalDistance.Float64
	}
	if stopDuration.Valid {
		distanceAnalysis.StopDurationSeconds = int(stopDuration.Float64)
	}

	// Calculate route efficiency compared to straight-line distance
	if len(locations) > 1 {
		startLoc := locations[0]
		endLoc := locations[len(locations)-1]
		straightLineDistance := haversineDistance(startLoc.Latitude, startLoc.Longitude, endLoc.Latitude, endLoc.Longitude)
		
		if straightLineDistance > 0 && distanceAnalysis.TotalDistanceKM > straightLineDistance {
			distanceAnalysis.RouteEfficiency = (straightLineDistance / distanceAnalysis.TotalDistanceKM) * 100
		} else if distanceAnalysis.TotalDistanceKM > 0 {
			distanceAnalysis.RouteEfficiency = 100 // Same location or very short trip
		}
	}

	return distanceAnalysis, nil
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return R * c
}

type AnomalyDetectionResult struct {
	Type         string
	Data         any
	Confidence   float64
	Score        float64
	Recommendation string
}

func (ta *TelemetryAnalytics) detectAnomalies(vehicleID uuid.UUID, startTime, endTime time.Time) ([]AnomalyDetectionResult, error) {
	var anomalies []AnomalyDetectionResult

	// Detect unusual speed patterns
	speedAnomalies, err := ta.detectSpeedAnomalies(vehicleID, startTime, endTime)
	if err != nil {
		ta.logger.Warn("Failed to detect speed anomalies", zap.Error(err))
	} else {
		anomalies = append(anomalies, speedAnomalies...)
	}

	// Detect unusual fuel patterns
	fuelAnomalies, err := ta.detectFuelAnomalies(vehicleID, startTime, endTime)
	if err != nil {
		ta.logger.Warn("Failed to detect fuel anomalies", zap.Error(err))
	} else {
		anomalies = append(anomalies, fuelAnomalies...)
	}

	// Detect unusual location patterns
	locationAnomalies, err := ta.detectLocationAnomalies(vehicleID, startTime, endTime)
	if err != nil {
		ta.logger.Warn("Failed to detect location anomalies", zap.Error(err))
	} else {
		anomalies = append(anomalies, locationAnomalies...)
	}

	return anomalies, nil
}

func (ta *TelemetryAnalytics) detectSpeedAnomalies(vehicleID uuid.UUID, startTime, endTime time.Time) ([]AnomalyDetectionResult, error) {
	query := `
		WITH speed_stats AS (
			SELECT 
				PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY speed) as median_speed,
				STDDEV(speed) as speed_stddev,
				AVG(speed) as avg_speed
			FROM gps_locations 
			WHERE vehicle_id = $1 
				AND time BETWEEN $2 AND $3
				AND speed IS NOT NULL
				AND speed > 0
		),
		outliers AS (
			SELECT 
				time,
				speed,
				ABS(speed - (SELECT avg_speed FROM speed_stats)) / (SELECT NULLIF(speed_stddev, 0) FROM speed_stats) as z_score
			FROM gps_locations, speed_stats
			WHERE vehicle_id = $1 
				AND time BETWEEN $2 AND $3
				AND speed IS NOT NULL
				AND speed > 0
				AND (SELECT speed_stddev FROM speed_stats) > 0
				AND ABS(speed - (SELECT avg_speed FROM speed_stats)) / (SELECT NULLIF(speed_stddev, 0) FROM speed_stats) > 2.0
		)
		SELECT 
			AVG(z_score) as avg_z_score,
			COUNT(*) as outlier_count
		FROM outliers`

	var avgZScore, outlierCount sql.NullFloat64

	row := ta.db.QueryRow(context.Background(), query, vehicleID, startTime, endTime)
	err := row.Scan(&avgZScore, &outlierCount)
	if err != nil {
		return nil, fmt.Errorf("failed to detect speed anomalies: %w", err)
	}

	var anomalies []AnomalyDetectionResult
	if outlierCount.Valid && outlierCount.Float64 > 0 {
		anomaly := AnomalyDetectionResult{
			Type:       "unusual_speed_pattern",
			Data:       map[string]interface{}{"outlier_count": int(outlierCount.Float64)},
			Confidence: 0.85,
			Score:      avgZScore.Float64,
			Recommendation: "Review driving behavior and check for possible unauthorized use",
		}
		anomalies = append(anomalies, anomaly)
	}

	return anomalies, nil
}

func (ta *TelemetryAnalytics) detectFuelAnomalies(vehicleID uuid.UUID, startTime, endTime time.Time) ([]AnomalyDetectionResult, error) {
	query := `
		WITH fuel_stats AS (
			SELECT 
				AVG(level_percent) as avg_fuel_level,
				STDDEV(level_percent) as fuel_stddev
			FROM fuel_readings 
			WHERE vehicle_id = $1 
				AND time BETWEEN $2 AND $3
				AND level_percent IS NOT NULL
		),
		fuel_changes AS (
			SELECT 
				time,
				level_percent,
				LAG(level_percent) OVER (ORDER BY time) as prev_level,
				LAG(time) OVER (ORDER BY time) as prev_time
			FROM fuel_readings
			WHERE vehicle_id = $1 
				AND time BETWEEN $2 AND $3
				AND level_percent IS NOT NULL
			ORDER BY time
		),
		unusual_drops AS (
			SELECT 
				time,
				level_percent,
				prev_level,
				(prev_level - level_percent) as drop_amount
			FROM fuel_changes
			WHERE prev_level IS NOT NULL 
				AND (prev_level - level_percent) > 5  -- Significant drop (>5%)
				AND level_percent < (SELECT avg_fuel_level FROM fuel_stats WHERE fuel_stddev > 0)
		)
		SELECT 
			COUNT(*) as unusual_drops,
			AVG(drop_amount) as avg_drop
		FROM unusual_drops`

	var unusualDrops, avgDrop sql.NullFloat64

	row := ta.db.QueryRow(context.Background(), query, vehicleID, startTime, endTime)
	err := row.Scan(&unusualDrops, &avgDrop)
	if err != nil {
		return nil, fmt.Errorf("failed to detect fuel anomalies: %w", err)
	}

	var anomalies []AnomalyDetectionResult
	if unusualDrops.Valid && unusualDrops.Float64 > 0 {
		anomaly := AnomalyDetectionResult{
			Type:       "fuel_theft_suspected",
			Data:       map[string]interface{}{"unusual_drops": int(unusualDrops.Float64), "avg_drop_percent": avgDrop.Float64},
			Confidence: 0.90,
			Score:      unusualDrops.Float64 * 10, // Higher score for more drops
			Recommendation: "Investigate potential fuel theft and check fuel system integrity",
		}
		anomalies = append(anomalies, anomaly)
	}

	return anomalies, nil
}

func (ta *TelemetryAnalytics) detectLocationAnomalies(vehicleID uuid.UUID, startTime, endTime time.Time) ([]AnomalyDetectionResult, error) {
	// This would involve detecting if the vehicle went to unusual locations
	// For now, we'll implement a basic version that detects excessive idling in one spot
	
	query := `
		WITH location_clusters AS (
			SELECT 
				latitude,
				longitude,
				time,
				ROW_NUMBER() OVER (ORDER BY time) as rn
			FROM gps_locations 
			WHERE vehicle_id = $1 
				AND time BETWEEN $2 AND $3
				AND latitude IS NOT NULL
				AND longitude IS NOT NULL
				AND speed < 5  -- Stationary or very slow
		),
		idle_periods AS (
			SELECT 
				latitude,
				longitude,
				COUNT(*) as idle_points,
				MAX(time) - MIN(time) as idle_duration
			FROM location_clusters
			GROUP BY latitude, longitude
			HAVING COUNT(*) > 10  -- More than 10 readings in same spot
		)
		SELECT 
			COUNT(*) as excessive_idle_locations,
			AVG(EXTRACT(EPOCH FROM idle_duration)) as avg_idle_duration
		FROM idle_periods
		WHERE EXTRACT(EPOCH FROM idle_duration) > 3600  -- More than 1 hour idle
	`

	var excessiveIdleLocations, avgIdleDuration sql.NullFloat64

	row := ta.db.QueryRow(context.Background(), query, vehicleID, startTime, endTime)
	err := row.Scan(&excessiveIdleLocations, &avgIdleDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to detect location anomalies: %w", err)
	}

	var anomalies []AnomalyDetectionResult
	if excessiveIdleLocations.Valid && excessiveIdleLocations.Float64 > 0 {
		anomaly := AnomalyDetectionResult{
			Type:       "excessive_idle_time",
			Data:       map[string]interface{}{"excessive_idle_locations": int(excessiveIdleLocations.Float64), "avg_idle_duration_seconds": avgIdleDuration.Float64},
			Confidence: 0.75,
			Score:      excessiveIdleLocations.Float64 * 5, // Score based on number of locations
			Recommendation: "Review vehicle usage patterns and driver behavior for unnecessary idling",
		}
		anomalies = append(anomalies, anomaly)
	}

	return anomalies, nil
}

func (ta *TelemetryAnalytics) saveAnalyticsResult(vehicleID uuid.UUID, analysisType string, data any, confidence float64) error {
	query := `
		INSERT INTO telemetry_analytics (
			id, vehicle_id, analysis_type, data, timestamp, confidence
		) VALUES ($1, $2, $3, $4, $5, $6)`

	id := uuid.New().String()
	now := time.Now()

	// Convert data to JSON
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal analytics data: %w", err)
	}

	_, err = ta.db.Exec(context.Background(), query, id, vehicleID, analysisType, dataBytes, now, confidence)
	if err != nil {
		return fmt.Errorf("failed to save analytics result: %w", err)
	}

	return nil
}

func (ta *TelemetryAnalytics) saveAnomalyResult(vehicleID uuid.UUID, analysisType string, data any, confidence, anomalyScore float64, recommendation *string) error {
	query := `
		INSERT INTO telemetry_anomalies (
			id, vehicle_id, anomaly_type, data, timestamp, confidence, anomaly_score, recommendation
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	id := uuid.New().String()
	now := time.Now()

	// Convert data to JSON
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal anomaly data: %w", err)
	}

	_, err = ta.db.Exec(context.Background(), query, id, vehicleID, analysisType, dataBytes, now, confidence, anomalyScore, recommendation)
	if err != nil {
		return fmt.Errorf("failed to save anomaly result: %w", err)
	}

	return nil
}
