//go:build integration
// +build integration

package service

import (
	"math"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/tracking/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupAnomalyEngine() *AnomalyEngine {
	return NewAnomalyEngine(DefaultAnomalyConfig())
}

func baseEvent() *models.TelemetryEvent {
	return &models.TelemetryEvent{
		EventID:    uuid.NewString(),
		DeviceID:   "DEV-001",
		VehicleID:  uuid.New(),
		Lat:        12.9716,
		Lng:        77.5946,
		Speed:      60,
		Heading:    90,
		Ignition:   true,
		EventType:  models.EventTypeGPS,
		RecordedAt: time.Now(),
	}
}

// ── 1. Normal event ────────────────────────────────────────────────────────────

func TestAnomalyEngine_NoAnomaly(t *testing.T) {
	eng := setupAnomalyEngine()
	cur := baseEvent()

	// Prev event: same position, similar speed, 5 s ago
	prev := baseEvent()
	prev.RecordedAt = cur.RecordedAt.Add(-5 * time.Second)
	prev.Speed = 58

	anomalies := eng.Evaluate(cur, prev)
	assert.Empty(t, anomalies, "normal event should produce no anomalies")
}

// ── 2. Panic button ────────────────────────────────────────────────────────────

func TestAnomalyEngine_PanicButton(t *testing.T) {
	eng := setupAnomalyEngine()
	cur := baseEvent()
	cur.EventType = models.EventTypePanic

	anomalies := eng.Evaluate(cur, nil)

	assert.Len(t, anomalies, 1)
	assert.Equal(t, models.AnomalyPanic, anomalies[0].Type)
	assert.Equal(t, "CRITICAL", anomalies[0].Severity)
}

// ── 3. Overspeed ──────────────────────────────────────────────────────────────

func TestAnomalyEngine_Overspeed(t *testing.T) {
	eng := setupAnomalyEngine()
	cur := baseEvent()
	cur.Speed = 150 // > 120 threshold

	anomalies := eng.Evaluate(cur, nil)

	assert.Len(t, anomalies, 1)
	assert.Equal(t, models.AnomalyOverspeed, anomalies[0].Type)
	assert.Equal(t, "HIGH", anomalies[0].Severity)
}

func TestAnomalyEngine_SpeedAtThreshold_NoAnomaly(t *testing.T) {
	eng := setupAnomalyEngine()
	cur := baseEvent()
	cur.Speed = 120 // exactly at threshold — not over

	anomalies := eng.Evaluate(cur, nil)
	assert.Empty(t, anomalies)
}

// ── 4. Harsh acceleration ──────────────────────────────────────────────────────

func TestAnomalyEngine_HarshAcceleration(t *testing.T) {
	eng := setupAnomalyEngine()
	now := time.Now()

	cur := baseEvent()
	cur.Speed = 100
	cur.RecordedAt = now

	prev := baseEvent()
	prev.Speed = 60
	prev.RecordedAt = now.Add(-1 * time.Second) // speed jumped 40 km/h in 1 s

	anomalies := eng.Evaluate(cur, prev)

	harsh := findAnomaly(anomalies, models.AnomalyHarshEvent)
	assert.NotNil(t, harsh)
	assert.Equal(t, "MEDIUM", harsh.Severity)
}

func TestAnomalyEngine_HarshBraking(t *testing.T) {
	eng := setupAnomalyEngine()
	now := time.Now()

	cur := baseEvent()
	cur.Speed = 20
	cur.RecordedAt = now

	prev := baseEvent()
	prev.Speed = 80
	prev.RecordedAt = now.Add(-1 * time.Second) // dropped 60 km/h in 1 s

	anomalies := eng.Evaluate(cur, prev)
	harsh := findAnomaly(anomalies, models.AnomalyHarshEvent)
	assert.NotNil(t, harsh)
}

// ── 5. Teleport ───────────────────────────────────────────────────────────────

func TestAnomalyEngine_Teleport(t *testing.T) {
	eng := setupAnomalyEngine()
	now := time.Now()

	cur := baseEvent()
	cur.Lat = 19.0760 // Mumbai
	cur.Lng = 72.8777
	cur.RecordedAt = now

	prev := baseEvent()
	prev.Lat = 12.9716 // Bengaluru (> 1000 km away)
	prev.Lng = 77.5946
	prev.RecordedAt = now.Add(-3 * time.Second) // within teleport window

	anomalies := eng.Evaluate(cur, prev)
	teleport := findAnomaly(anomalies, models.AnomalyTeleport)
	assert.NotNil(t, teleport)
}

func TestAnomalyEngine_NoTeleport_OutsideWindow(t *testing.T) {
	eng := setupAnomalyEngine()
	now := time.Now()

	cur := baseEvent()
	cur.Lat = 19.0760
	cur.Lng = 72.8777
	cur.RecordedAt = now

	prev := baseEvent()
	prev.Lat = 12.9716
	prev.Lng = 77.5946
	prev.RecordedAt = now.Add(-60 * time.Second) // outside 5-s window

	anomalies := eng.Evaluate(cur, prev)
	teleport := findAnomaly(anomalies, models.AnomalyTeleport)
	assert.Nil(t, teleport, "large distance outside time window should not be flagged as teleport")
}

// ── 6. Fuel drop ──────────────────────────────────────────────────────────────

func TestAnomalyEngine_FuelDrop(t *testing.T) {
	eng := setupAnomalyEngine()
	now := time.Now()

	prev := baseEvent()
	prevFuel := 80.0
	prev.FuelLevel = &prevFuel
	prev.RecordedAt = now.Add(-30 * time.Second)

	cur := baseEvent()
	curFuel := 60.0 // dropped 20% — above 10% threshold
	cur.FuelLevel = &curFuel
	cur.RecordedAt = now

	anomalies := eng.Evaluate(cur, prev)
	drop := findAnomaly(anomalies, models.AnomalyFuelDrop)
	assert.NotNil(t, drop)
}

func TestAnomalyEngine_SmallFuelDrop_NoAnomaly(t *testing.T) {
	eng := setupAnomalyEngine()
	now := time.Now()

	prev := baseEvent()
	prevFuel := 80.0
	prev.FuelLevel = &prevFuel
	prev.RecordedAt = now.Add(-30 * time.Second)

	cur := baseEvent()
	curFuel := 77.0 // only 3% drop
	cur.FuelLevel = &curFuel
	cur.RecordedAt = now

	anomalies := eng.Evaluate(cur, prev)
	drop := findAnomaly(anomalies, models.AnomalyFuelDrop)
	assert.Nil(t, drop)
}

// ── 7. Odometer rollback ──────────────────────────────────────────────────────

func TestAnomalyEngine_OdometerRollback(t *testing.T) {
	eng := setupAnomalyEngine()
	now := time.Now()

	prev := baseEvent()
	prevOdo := 12500.0
	prev.Odometer = &prevOdo
	prev.RecordedAt = now.Add(-10 * time.Second)

	cur := baseEvent()
	curOdo := 12000.0 // rolled back
	cur.Odometer = &curOdo
	cur.RecordedAt = now

	anomalies := eng.Evaluate(cur, prev)
	rollback := findAnomaly(anomalies, models.AnomalyOdometerRollback)
	assert.NotNil(t, rollback)
}

// ── 8. Nil prev ───────────────────────────────────────────────────────────────

func TestAnomalyEngine_NilPrev_NoDeltas(t *testing.T) {
	eng := setupAnomalyEngine()
	cur := baseEvent()
	cur.Speed = 119 // just under overspeed threshold

	anomalies := eng.Evaluate(cur, nil)
	assert.Empty(t, anomalies)
}

// ── 9. Haversine helper ───────────────────────────────────────────────────────

func TestHaversine_BengaluruMumbai(t *testing.T) {
	// Bengaluru → Mumbai great-circle ≈ 845 km
	distM := haversineM(12.9716, 77.5946, 19.0760, 72.8777)
	distKm := distM / 1000
	assert.InDelta(t, 845, distKm, 20, "Bengaluru–Mumbai distance")
}

func TestHaversine_SamePoint(t *testing.T) {
	distM := haversineM(12.9716, 77.5946, 12.9716, 77.5946)
	assert.InDelta(t, 0, distM, 0.001)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func findAnomaly(anomalies []models.AnomalyEvent, typ string) *models.AnomalyEvent {
	for i := range anomalies {
		if anomalies[i].Type == typ {
			return &anomalies[i]
		}
	}
	return nil
}

// Ensure haversineM is exported for test usage from the same package.
var _ = math.Sqrt
