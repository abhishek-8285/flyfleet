package service

import (
	"fmt"
	"math"

	"github.com/flyfleet/fleet-core-service/internal/domains/tracking/models"
)

// AnomalyConfig holds thresholds for anomaly detection.
type AnomalyConfig struct {
	OverspeedThresholdKmh float64 // default 120
	HarshAccelG           float64 // default 0.3G (roughly 10.6 km/h per sec)
	HarshBrakingG         float64 // default 0.4G
	TeleportDistanceM     float64 // default 500 m
	TeleportTimeWindowSec float64 // default 5 s
	FuelDropThresholdPct  float64 // default 10%
}

// DefaultAnomalyConfig returns production-default thresholds.
func DefaultAnomalyConfig() AnomalyConfig {
	return AnomalyConfig{
		OverspeedThresholdKmh: 120,
		HarshAccelG:           0.3,
		HarshBrakingG:         0.4,
		TeleportDistanceM:     500,
		TeleportTimeWindowSec: 5,
		FuelDropThresholdPct:  10,
	}
}

// AnomalyEngine is a stateless evaluator that compares the current event against
// the previous event to detect anomalies. It performs zero I/O.
type AnomalyEngine struct {
	cfg AnomalyConfig
}

// NewAnomalyEngine creates a new AnomalyEngine with the given config.
func NewAnomalyEngine(cfg AnomalyConfig) *AnomalyEngine {
	return &AnomalyEngine{cfg: cfg}
}

// Evaluate compares the current event with the previous one to detect anomalies.
func (e *AnomalyEngine) Evaluate(cur, prev *models.TelemetryEvent) []models.AnomalyEvent {
	var anomalies []models.AnomalyEvent

	// 1 ── Panic Trigger (Stateless)
	if cur.EventType == models.EventTypePanic {
		anomalies = append(anomalies, e.newAnomaly(cur, models.AnomalyPanic, "CRITICAL", "SOS button triggered"))
	}

	// 2 ── Overspeed detection
	if cur.Speed > e.cfg.OverspeedThresholdKmh {
		msg := fmt.Sprintf("Vehicle overspeeding: %g km/h (limit %g)", cur.Speed, e.cfg.OverspeedThresholdKmh)
		anomalies = append(anomalies, e.newAnomaly(cur, models.AnomalyOverspeed, "HIGH", msg))
	}

	if prev == nil {
		return anomalies
	}

	dtSec := cur.RecordedAt.Sub(prev.RecordedAt).Seconds()
	if dtSec <= 0 {
		return anomalies // out-of-order or identical timestamps
	}

	// 3 ── Harsh Acceleration/Braking (True Δv / Δt)
	// (curSpeed - prevSpeed) in km/h -> m/s is divide by 3.6
	dvMps := (cur.Speed - prev.Speed) / 3.6
	accelG := dvMps / (dtSec * 9.81)

	if accelG > e.cfg.HarshAccelG {
		msg := fmt.Sprintf("Harsh acceleration: %0.2fG", accelG)
		anomalies = append(anomalies, e.newAnomaly(cur, models.AnomalyHarshEvent, "MEDIUM", msg))
	} else if accelG < -e.cfg.HarshBrakingG {
		msg := fmt.Sprintf("Harsh braking: %0.2fG", -accelG)
		anomalies = append(anomalies, e.newAnomaly(cur, models.AnomalyHarshEvent, "MEDIUM", msg))
	}

	// 4 ── Teleport detection
	if dtSec <= e.cfg.TeleportTimeWindowSec {
		distM := haversineM(prev.Lat, prev.Lng, cur.Lat, cur.Lng)
		if distM > e.cfg.TeleportDistanceM {
			msg := fmt.Sprintf("Impossible location jump: %0.1fm in %0.1fs", distM, dtSec)
			anomalies = append(anomalies, e.newAnomaly(cur, models.AnomalyTeleport, "HIGH", msg))
		}
	}

	// 5 ── Fuel drop
	if cur.FuelLevel != nil && prev.FuelLevel != nil {
		drop := *prev.FuelLevel - *cur.FuelLevel
		if drop > e.cfg.FuelDropThresholdPct {
			msg := fmt.Sprintf("Significant fuel drop: %0.1f%%", drop)
			anomalies = append(anomalies, e.newAnomaly(cur, models.AnomalyFuelDrop, "HIGH", msg))
		}
	}

	// 6 ── Odometer rollback
	if cur.Odometer != nil && prev.Odometer != nil && *cur.Odometer < *prev.Odometer {
		msg := fmt.Sprintf("Odometer rollback: %g -> %g", *prev.Odometer, *cur.Odometer)
		anomalies = append(anomalies, e.newAnomaly(cur, models.AnomalyOdometerRollback, "HIGH", msg))
	}

	// 7 ── Ignition mismatch
	if !cur.Ignition && cur.Speed > 5 {
		anomalies = append(anomalies, e.newAnomaly(cur, models.AnomalyTamper, "MEDIUM", "Movement with ignition OFF"))
	}

	return anomalies
}

func (e *AnomalyEngine) newAnomaly(cur *models.TelemetryEvent, typ, severity, msg string) models.AnomalyEvent {
	return models.AnomalyEvent{
		Type:      typ,
		VehicleID: cur.VehicleID,
		DeviceID:  cur.DeviceID,
		Severity:  severity,
		Message:   msg,
		Payload: map[string]any{
			"lat":   cur.Lat,
			"lng":   cur.Lng,
			"speed": cur.Speed,
		},
		Timestamp: cur.RecordedAt,
	}
}

// haversineM returns the great-circle distance in metres between two lat/lng points.
func haversineM(lat1, lng1, lat2, lng2 float64) float64 {
	const earthR = 6_371_000.0 // metres
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthR * c
}
