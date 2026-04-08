package service

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
)

func TestCalculateSummary_EmptyPoints(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &TelemetryQueryService{
		logger: logger,
	}

	summary := svc.calculateSummary([]models.TelemetryEvent{})
	if summary.TotalDistanceKM != 0 {
		t.Errorf("expected 0 distance for empty points, got %f", summary.TotalDistanceKM)
	}
	if summary.MovingTimeHours != 0 {
		t.Errorf("expected 0 moving time for empty points, got %f", summary.MovingTimeHours)
	}
}

func TestCalculateSummary_SinglePoint(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &TelemetryQueryService{
		logger: logger,
	}

	now := time.Now().UTC()
	points := []models.TelemetryEvent{
		{Lat: 28.6139, Lng: 77.2090, Speed: 45.0, RecordedAt: now},
	}

	summary := svc.calculateSummary(points)
	if summary.MaxSpeedKmh != 45.0 {
		t.Errorf("expected max speed 45.0, got %f", summary.MaxSpeedKmh)
	}
	if summary.TotalDistanceKM != 0 {
		t.Errorf("expected 0 distance for single point, got %f", summary.TotalDistanceKM)
	}
}

func TestCalculateSummary_MovingVehicle(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &TelemetryQueryService{
		logger: logger,
	}

	base := time.Now().UTC()
	points := []models.TelemetryEvent{
		{Lat: 28.6139, Lng: 77.2090, Speed: 40.0, RecordedAt: base},
		{Lat: 28.6200, Lng: 77.2150, Speed: 50.0, RecordedAt: base.Add(60 * time.Second)},
		{Lat: 28.6300, Lng: 77.2250, Speed: 0.0, RecordedAt: base.Add(120 * time.Second)},
	}

	summary := svc.calculateSummary(points)
	if summary.TotalDistanceKM <= 0 {
		t.Error("expected positive distance for moving vehicle")
	}
	if summary.MaxSpeedKmh != 50.0 {
		t.Errorf("expected max speed 50.0, got %f", summary.MaxSpeedKmh)
	}
	if summary.MovingTimeHours <= 0 {
		t.Error("expected positive moving time")
	}
}

func TestResolveResolution(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &TelemetryQueryService{
		logger: logger,
	}

	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"30min", 30 * time.Minute, "10 seconds"},
		{"59min", 59 * time.Minute, "10 seconds"},
		{"3h", 3 * time.Hour, "30 seconds"},
		{"5h", 5 * time.Hour, "30 seconds"},
		{"12h", 12 * time.Hour, "60 seconds"},
		{"23h", 23 * time.Hour, "60 seconds"},
		{"3d", 72 * time.Hour, "5 minutes"},
		{"6d", 144 * time.Hour, "5 minutes"},
		{"15d", 360 * time.Hour, "15 minutes"},
		{"30d", 720 * time.Hour, "15 minutes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now().UTC().Add(-tt.duration)
			end := time.Now().UTC()
			result := svc.resolveResolution(start, end, "auto")
			if result != tt.expected {
				t.Errorf("duration %v: expected %q, got %q", tt.duration, tt.expected, result)
			}
		})
	}
}

func TestResolveResolution_ExplicitOverride(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &TelemetryQueryService{
		logger: logger,
	}

	start := time.Now().UTC().Add(-24 * time.Hour)
	end := time.Now().UTC()

	// Explicit resolution should override auto
	result := svc.resolveResolution(start, end, "5 minutes")
	if result != "5 minutes" {
		t.Errorf("expected explicit 5 minutes, got %q", result)
	}
}

func TestGetFilteredOrgLocations_StatusFilter(t *testing.T) {
	locations := []models.VehicleLocationResponse{
		{VehicleID: "v1", Registration: "DL01AB1234", Status: models.StatusMoving},
		{VehicleID: "v2", Registration: "DL02CD5678", Status: models.StatusIdle},
		{VehicleID: "v3", Registration: "DL03EF9012", Status: models.StatusStopped},
		{VehicleID: "v4", Registration: "DL04GH3456", Status: models.StatusOffline},
	}

	// Test status filter: moving
	filtered := filterLocations(locations, "moving", uuid.Nil, "")
	if len(filtered) != 1 || filtered[0].Status != models.StatusMoving {
		t.Errorf("expected 1 moving, got %d", len(filtered))
	}

	// Test status filter: offline
	filtered = filterLocations(locations, "offline", uuid.Nil, "")
	if len(filtered) != 1 || filtered[0].Status != models.StatusOffline {
		t.Errorf("expected 1 offline, got %d", len(filtered))
	}

	// Test no filter
	filtered = filterLocations(locations, "", uuid.Nil, "")
	if len(filtered) != 4 {
		t.Errorf("expected 4 locations with no filter, got %d", len(filtered))
	}
}

func TestGetFilteredOrgLocations_SearchFilter(t *testing.T) {
	locations := []models.VehicleLocationResponse{
		{VehicleID: "v1", Registration: "DL01AB1234", Status: models.StatusMoving},
		{VehicleID: "v2", Registration: "MH02CD5678", Status: models.StatusIdle},
		{VehicleID: "v3", Registration: "KA03EF9012", Status: models.StatusStopped},
	}

	// Search by registration
	filtered := filterLocations(locations, "", uuid.Nil, "DL01")
	if len(filtered) != 1 || filtered[0].Registration != "DL01AB1234" {
		t.Errorf("expected 1 match for DL01, got %d", len(filtered))
	}

	// Search case-insensitive
	filtered = filterLocations(locations, "", uuid.Nil, "dl01")
	if len(filtered) != 1 {
		t.Errorf("expected case-insensitive match, got %d", len(filtered))
	}

	// Search no match
	filtered = filterLocations(locations, "", uuid.Nil, "ZZZZ")
	if len(filtered) != 0 {
		t.Errorf("expected no match for ZZZZ, got %d", len(filtered))
	}
}

func TestGetFilteredOrgLocations_DriverFilter(t *testing.T) {
	driverID := uuid.Must(uuid.NewV4())
	otherDriverID := uuid.Must(uuid.NewV4())
	locations := []models.VehicleLocationResponse{
		{
			VehicleID:    "v1",
			Registration: "DL01AB1234",
			Status:       models.StatusMoving,
			Driver:       &models.DriverInfo{ID: driverID.String(), Name: "Rajesh Kumar"},
		},
		{
			VehicleID:    "v2",
			Registration: "DL02CD5678",
			Status:       models.StatusIdle,
			Driver:       &models.DriverInfo{ID: otherDriverID.String(), Name: "Amit Sharma"},
		},
	}

	// Filter by driver ID
	filtered := filterLocations(locations, "", driverID, "")
	if len(filtered) != 1 || filtered[0].Driver.ID != driverID.String() {
		t.Errorf("expected 1 match for driver, got %d", len(filtered))
	}

	// Filter by driver name search
	filtered = filterLocations(locations, "", uuid.Nil, "Rajesh")
	if len(filtered) != 1 || filtered[0].Driver.Name != "Rajesh Kumar" {
		t.Errorf("expected 1 match for driver name Rajesh, got %d", len(filtered))
	}

	// Combined: status + driver
	filtered = filterLocations(locations, models.StatusMoving, driverID, "")
	if len(filtered) != 1 || filtered[0].VehicleID != "v1" {
		t.Errorf("expected 1 match for moving+driver, got %d", len(filtered))
	}
}

// filterLocations mirrors the filter logic in GetFilteredOrgLocations.
func filterLocations(all []models.VehicleLocationResponse, statusFilter string, driverID uuid.UUID, search string) []models.VehicleLocationResponse {
	filtered := make([]models.VehicleLocationResponse, 0, len(all))
	for _, loc := range all {
		if statusFilter != "" && loc.Status != statusFilter {
			continue
		}
		if driverID != uuid.Nil {
			if loc.Driver == nil || loc.Driver.ID != driverID.String() {
				continue
			}
		}
		if search != "" {
			searchLower := toLower(search)
			matchReg := contains(toLower(loc.Registration), searchLower)
			matchDriver := loc.Driver != nil && contains(toLower(loc.Driver.Name), searchLower)
			if !matchReg && !matchDriver {
				continue
			}
		}
		filtered = append(filtered, loc)
	}
	return filtered
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
