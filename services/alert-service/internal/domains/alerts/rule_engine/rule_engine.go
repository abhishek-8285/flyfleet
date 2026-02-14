package rule_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/alert-service/internal/domains/alerts/models"
	"github.com/flyfleet/alert-service/internal/domains/alerts/repository"
)

// RuleEngine processes telemetry events and generates alerts based on rules
type RuleEngine struct {
	repo   *repository.AlertRepository
	logger *zap.Logger
}

// TelemetryEvent represents incoming telemetry data
type TelemetryEvent struct {
	VehicleID uuid.UUID          `json:"vehicle_id"`
	DriverID  *uuid.UUID         `json:"driver_id"`
	Timestamp time.Time          `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// GPSEvent represents GPS telemetry data
type GPSEvent struct {
	VehicleID   uuid.UUID `json:"vehicle_id"`
	DriverID    *uuid.UUID `json:"driver_id"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Speed       float64   `json:"speed"`
	Heading     float64   `json:"heading"`
	Accuracy    float64   `json:"accuracy"`
	Ignition    bool      `json:"ignition"`
	BatteryLevel float64  `json:"battery_level"`
	Satellites  int       `json:"satellites"`
	Timestamp   time.Time `json:"timestamp"`
}

// OBDEvent represents OBD telemetry data
type OBDEvent struct {
	VehicleID     uuid.UUID `json:"vehicle_id"`
	DriverID      *uuid.UUID `json:"driver_id"`
	RPM           int       `json:"rpm"`
	Speed         float64   `json:"speed"`
	FuelLevel     float64   `json:"fuel_level"`
	EngineTemp    float64   `json:"engine_temp"`
	IntakeTemp    float64   `json:"intake_temp"`
	ThrottlePos   float64   `json:"throttle_position"`
	EngineLoad    float64   `json:"engine_load"`
	Timestamp     time.Time `json:"timestamp"`
}

// FuelEvent represents fuel telemetry data
type FuelEvent struct {
	VehicleID     uuid.UUID `json:"vehicle_id"`
	DriverID      *uuid.UUID `json:"driver_id"`
	FuelLevel     float64   `json:"fuel_level"`
	FuelConsumed  float64   `json:"fuel_consumed"`
	Timestamp     time.Time `json:"timestamp"`
}

// DriverEvent represents driver behavior data
type DriverEvent struct {
	VehicleID     uuid.UUID `json:"vehicle_id"`
	DriverID      uuid.UUID `json:"driver_id"`
	EventType     string    `json:"event_type"` // harsh_acceleration, harsh_braking, harsh_cornering, phone_usage, fatigue
	Severity      string    `json:"severity"`   // low, medium, high, critical
	Value         float64   `json:"value"`      // acceleration g-force, speed, etc.
	Timestamp     time.Time `json:"timestamp"`
}

// NewRuleEngine creates a new rule engine
func NewRuleEngine(repo *repository.AlertRepository, logger *zap.Logger) *RuleEngine {
	return &RuleEngine{
		repo:   repo,
		logger: logger,
	}
}

// ProcessGPSEvent processes GPS telemetry events and checks for rule violations
func (re *RuleEngine) ProcessGPSEvent(data []byte) error {
	var event GPSEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal GPS event: %w", err)
	}

	// Get active alert rules for this vehicle/org
	rules, err := re.repo.GetActiveRulesForVehicle(context.Background(), event.VehicleID)
	if err != nil {
		re.logger.Error("Failed to get active rules", zap.Error(err))
		return err
	}

	// Check each rule
	for _, rule := range rules {
		if err := re.evaluateGPSRule(rule, event); err != nil {
			re.logger.Error("Failed to evaluate GPS rule", 
				zap.String("rule_id", rule.ID), 
				zap.Error(err))
		}
	}

	return nil
}

// ProcessOBDEvent processes OBD telemetry events
func (re *RuleEngine) ProcessOBDEvent(data []byte) error {
	var event OBDEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal OBD event: %w", err)
	}

	// Get active alert rules for this vehicle/org
	rules, err := re.repo.GetActiveRulesForVehicle(context.Background(), event.VehicleID)
	if err != nil {
		re.logger.Error("Failed to get active rules", zap.Error(err))
		return err
	}

	// Check each rule
	for _, rule := range rules {
		if err := re.evaluateOBDRule(rule, event); err != nil {
			re.logger.Error("Failed to evaluate OBD rule", 
				zap.String("rule_id", rule.ID), 
				zap.Error(err))
		}
	}

	return nil
}

// ProcessFuelEvent processes fuel telemetry events
func (re *RuleEngine) ProcessFuelEvent(data []byte) error {
	var event FuelEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal fuel event: %w", err)
	}

	// Get active alert rules for this vehicle/org
	rules, err := re.repo.GetActiveRulesForVehicle(context.Background(), event.VehicleID)
	if err != nil {
		re.logger.Error("Failed to get active rules", zap.Error(err))
		return err
	}

	// Check each rule
	for _, rule := range rules {
		if err := re.evaluateFuelRule(rule, event); err != nil {
			re.logger.Error("Failed to evaluate fuel rule", 
				zap.String("rule_id", rule.ID), 
				zap.Error(err))
		}
	}

	return nil
}

// ProcessDriverEvent processes driver behavior events
func (re *RuleEngine) ProcessDriverEvent(data []byte) error {
	var event DriverEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal driver event: %w", err)
	}

	// Get active alert rules for this driver/org
	rules, err := re.repo.GetActiveRulesForDriver(context.Background(), event.DriverID)
	if err != nil {
		re.logger.Error("Failed to get active rules", zap.Error(err))
		return err
	}

	// Check each rule
	for _, rule := range rules {
		if err := re.evaluateDriverRule(rule, event); err != nil {
			re.logger.Error("Failed to evaluate driver rule", 
				zap.String("rule_id", rule.ID), 
				zap.Error(err))
		}
	}

	return nil
}

// evaluateGPSRule evaluates GPS-based rules
func (re *RuleEngine) evaluateGPSRule(rule *models.AlertRule, event GPSEvent) error {
	switch rule.Type {
	case models.AlertTypeOverspeed:
		return re.checkOverspeedRule(rule, event)
	case models.AlertTypeGeofenceEntry:
		return re.checkGeofenceEntryRule(rule, event)
	case models.AlertTypeGeofenceExit:
		return re.checkGeofenceExitRule(rule, event)
	case models.AlertTypeVehicleOffline:
		return re.checkVehicleOfflineRule(rule, event)
	default:
		return nil
	}
}

// evaluateOBDRule evaluates OBD-based rules
func (re *RuleEngine) evaluateOBDRule(rule *models.AlertRule, event OBDEvent) error {
	switch rule.Type {
	case models.AlertTypeHarshAcceleration:
		return re.checkHarshAccelerationRule(rule, event)
	case models.AlertTypeHarshBraking:
		return re.checkHarshBrakingRule(rule, event)
	case models.AlertTypeHarshCornering:
		return re.checkHarshCorneringRule(rule, event)
	case models.AlertTypeIdleTime:
		return re.checkIdleTimeRule(rule, event)
	default:
		return nil
	}
}

// evaluateFuelRule evaluates fuel-based rules
func (re *RuleEngine) evaluateFuelRule(rule *models.AlertRule, event FuelEvent) error {
	switch rule.Type {
	case models.AlertTypeFuelTheft:
		return re.checkFuelTheftRule(rule, event)
	default:
		return nil
	}
}

// evaluateDriverRule evaluates driver behavior rules
func (re *RuleEngine) evaluateDriverRule(rule *models.AlertRule, event DriverEvent) error {
	switch rule.Type {
	case models.AlertTypeDriverFatigue:
		return re.checkDriverFatigueRule(rule, event)
	case models.AlertTypeADASEvent:
		return re.checkADASRule(rule, event)
	default:
		return nil
	}
}

// checkOverspeedRule checks if vehicle is overspeeding
func (re *RuleEngine) checkOverspeedRule(rule *models.AlertRule, event GPSEvent) error {
	speedLimit, ok := rule.Conditions["speed_limit"].(float64)
	if !ok {
		return fmt.Errorf("invalid speed_limit condition")
	}

	durationSeconds, ok := rule.Conditions["duration_seconds"].(float64)
	if !ok {
		durationSeconds = 30 // Default 30 seconds
	}

	if event.Speed > speedLimit {
		// Create alert for overspeeding
		alert := &models.Alert{
			ID:        uuid.New().String(),
			OrgID:     rule.OrgID,
			RuleID:    rule.ID,
			Type:      models.AlertTypeOverspeed,
			Severity:  rule.Severity,
			VehicleID: &event.VehicleID,
			DriverID:  event.DriverID,
			Location: &models.Location{
				Latitude:  event.Latitude,
				Longitude: event.Longitude,
			},
			Timestamp: event.Timestamp,
			Details: map[string]interface{}{
				"current_speed": event.Speed,
				"speed_limit":   speedLimit,
				"duration":      durationSeconds,
			},
			Status: models.AlertStatusNew,
		}

		if err := re.repo.CreateAlert(context.Background(), alert); err != nil {
			return fmt.Errorf("failed to create overspeed alert: %w", err)
		}

		re.logger.Info("Overspeed alert created", 
			zap.String("vehicle_id", event.VehicleID.String()),
			zap.Float64("speed", event.Speed),
			zap.Float64("limit", speedLimit))
	}

	return nil
}

// checkGeofenceEntryRule checks for geofence entry violations
func (re *RuleEngine) checkGeofenceEntryRule(rule *models.AlertRule, event GPSEvent) error {
	// Implementation would check if vehicle entered a restricted geofence
	// This requires geofence data and point-in-polygon calculations
	return nil
}

// checkGeofenceExitRule checks for geofence exit violations
func (re *RuleEngine) checkGeofenceExitRule(rule *models.AlertRule, event GPSEvent) error {
	// Implementation would check if vehicle exited an authorized geofence
	return nil
}

// checkHarshAccelerationRule checks for harsh acceleration events
func (re *RuleEngine) checkHarshAccelerationRule(rule *models.AlertRule, event OBDEvent) error {
	threshold, ok := rule.Conditions["acceleration_threshold"].(float64)
	if !ok {
		threshold = 0.3 // Default 0.3g
	}

	// Calculate acceleration from RPM and speed changes (simplified)
	// In reality, this would use accelerometer data
	if event.ThrottlePos > 80 && event.Speed > 20 {
		alert := &models.Alert{
			ID:        uuid.New().String(),
			OrgID:     rule.OrgID,
			RuleID:    rule.ID,
			Type:      models.AlertTypeHarshAcceleration,
			Severity:  rule.Severity,
			VehicleID: &event.VehicleID,
			DriverID:  event.DriverID,
			Location: &models.Location{
				Latitude:  0, // Would get from GPS event
				Longitude: 0,
			},
			Timestamp: event.Timestamp,
			Details: map[string]interface{}{
				"throttle_position": event.ThrottlePos,
				"speed":            event.Speed,
				"threshold":        threshold,
			},
			Status: models.AlertStatusNew,
		}

		if err := re.repo.CreateAlert(context.Background(), alert); err != nil {
			return fmt.Errorf("failed to create harsh acceleration alert: %w", err)
		}

		re.logger.Info("Harsh acceleration alert created", 
			zap.String("vehicle_id", event.VehicleID.String()))
	}

	return nil
}

// checkHarshBrakingRule checks for harsh braking events
func (re *RuleEngine) checkHarshBrakingRule(rule *models.AlertRule, event OBDEvent) error {
	// Implementation would detect rapid deceleration
	return nil
}

// checkHarshCorneringRule checks for harsh cornering events
func (re *RuleEngine) checkHarshCorneringRule(rule *models.AlertRule, event OBDEvent) error {
	// Implementation would detect aggressive turns
	return nil
}

// checkIdleTimeRule checks for excessive idling
func (re *RuleEngine) checkIdleTimeRule(rule *models.AlertRule, event OBDEvent) error {
	// Implementation would track idle time duration
	return nil
}

// checkFuelTheftRule checks for fuel theft/drainage
func (re *RuleEngine) checkFuelTheftRule(rule *models.AlertRule, event FuelEvent) error {
	dropThreshold, ok := rule.Conditions["drop_threshold_liters"].(float64)
	if !ok {
		dropThreshold = 5 // Default 5 liters
	}
	_ = dropThreshold

	durationMinutes, ok := rule.Conditions["duration_minutes"].(float64)
	if !ok {
		durationMinutes = 10 // Default 10 minutes
	}
	_ = durationMinutes

	// Implementation would compare current fuel level with previous readings
	// and detect significant drops within specified time period
	return nil
}

// checkDriverFatigueRule checks for driver fatigue indicators
func (re *RuleEngine) checkDriverFatigueRule(rule *models.AlertRule, event DriverEvent) error {
	// Implementation would analyze driving patterns, time of day, etc.
	return nil
}

// checkADASRule checks for ADAS (Advanced Driver Assistance Systems) events
func (re *RuleEngine) checkADASRule(rule *models.AlertRule, event DriverEvent) error {
	// Implementation would process ADAS sensor data
	return nil
}

// checkVehicleOfflineRule checks if vehicle has gone offline
func (re *RuleEngine) checkVehicleOfflineRule(rule *models.AlertRule, event GPSEvent) error {
	offlineThreshold, ok := rule.Conditions["offline_threshold_minutes"].(float64)
	if !ok {
		offlineThreshold = 30 // Default 30 minutes
	}
	_ = offlineThreshold

	// Implementation would check last heartbeat timestamp
	return nil
}

// EvaluateCustomRule evaluates custom business rules
func (re *RuleEngine) EvaluateCustomRule(rule *models.AlertRule, eventData map[string]interface{}) error {
	// Generic rule evaluation for custom conditions
	// This would support complex rule expressions
	return nil
}
