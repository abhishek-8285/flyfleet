package handlers

import (
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
)

func TestHasSignificantChange_NoChange(t *testing.T) {
	loc := models.LiveLocation{
		Lat: 28.6139, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving,
	}
	if hasSignificantChange(loc, loc) {
		t.Error("expected no significant change for identical locations")
	}
}

func TestHasSignificantChange_SmallLatChange(t *testing.T) {
	old := models.LiveLocation{Lat: 28.6139, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving}
	// 0.00005 < 0.0001 threshold (~5.5m)
	new := models.LiveLocation{Lat: 28.61395, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving}
	if hasSignificantChange(old, new) {
		t.Error("expected no significant change for small lat delta")
	}
}

func TestHasSignificantChange_LargeLatChange(t *testing.T) {
	old := models.LiveLocation{Lat: 28.6139, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving}
	// 0.0002 > 0.0001 threshold (~22m)
	new := models.LiveLocation{Lat: 28.6141, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving}
	if !hasSignificantChange(old, new) {
		t.Error("expected significant change for large lat delta")
	}
}

func TestHasSignificantChange_LargeLngChange(t *testing.T) {
	old := models.LiveLocation{Lat: 28.6139, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving}
	new := models.LiveLocation{Lat: 28.6139, Lng: 77.2092, Speed: 40.0, Status: models.StatusMoving}
	if !hasSignificantChange(old, new) {
		t.Error("expected significant change for large lng delta")
	}
}

func TestHasSignificantChange_SmallSpeedChange(t *testing.T) {
	old := models.LiveLocation{Lat: 28.6139, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving}
	new := models.LiveLocation{Lat: 28.6139, Lng: 77.2090, Speed: 41.0, Status: models.StatusMoving}
	if hasSignificantChange(old, new) {
		t.Error("expected no significant change for 1 km/h speed delta")
	}
}

func TestHasSignificantChange_LargeSpeedChange(t *testing.T) {
	old := models.LiveLocation{Lat: 28.6139, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving}
	new := models.LiveLocation{Lat: 28.6139, Lng: 77.2090, Speed: 43.0, Status: models.StatusMoving}
	if !hasSignificantChange(old, new) {
		t.Error("expected significant change for 3 km/h speed delta (>2.0)")
	}
}

func TestHasSignificantChange_StatusChange(t *testing.T) {
	old := models.LiveLocation{Lat: 28.6139, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving}
	new := models.LiveLocation{Lat: 28.6139, Lng: 77.2090, Speed: 40.0, Status: models.StatusIdle}
	if !hasSignificantChange(old, new) {
		t.Error("expected significant change for status change")
	}
}

func TestHasSignificantChange_NegativeDelta(t *testing.T) {
	old := models.LiveLocation{Lat: 28.6141, Lng: 77.2092, Speed: 43.0, Status: models.StatusMoving}
	new := models.LiveLocation{Lat: 28.6139, Lng: 77.2090, Speed: 40.0, Status: models.StatusMoving}
	if !hasSignificantChange(old, new) {
		t.Error("expected significant change for negative deltas")
	}
}

func TestCalculateStatus_Moving(t *testing.T) {
	status := models.CalculateStatus(10.0, true, time.Now())
	if status != models.StatusMoving {
		t.Errorf("expected moving, got %s", status)
	}
}

func TestCalculateStatus_Idle(t *testing.T) {
	status := models.CalculateStatus(1.0, true, time.Now())
	if status != models.StatusIdle {
		t.Errorf("expected idle, got %s", status)
	}
}

func TestCalculateStatus_Stopped(t *testing.T) {
	status := models.CalculateStatus(0.0, false, time.Now())
	if status != models.StatusStopped {
		t.Errorf("expected stopped, got %s", status)
	}
}

func TestCalculateStatus_Offline(t *testing.T) {
	status := models.CalculateStatus(40.0, true, time.Now().Add(-10*time.Minute))
	if status != models.StatusOffline {
		t.Errorf("expected offline, got %s", status)
	}
}
