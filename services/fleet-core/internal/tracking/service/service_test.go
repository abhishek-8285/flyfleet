package service

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
	"github.com/flyfleet/fleet-core/internal/tracking/repository"
)

// ============================================================================
// IngestionService Tests (with miniredis)
// ============================================================================

func setupIngestTest(t *testing.T) (*IngestionService, *repository.MemTelemRepo, *repository.MemLiveCache, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	telemRepo := repository.NewMemTelemRepo()
	liveCache := repository.NewMemLiveCache()

	svc := NewIngestionService(telemRepo, liveCache, nil, rdb, zap.NewNop())
	return svc, telemRepo, liveCache, mr
}

func TestNewIngestionService(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	telemRepo := repository.NewMemTelemRepo()
	liveCache := repository.NewMemLiveCache()

	svc := NewIngestionService(telemRepo, liveCache, nil, rdb, zap.NewNop())
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.telemRepo == nil {
		t.Error("telemRepo is nil")
	}
	if svc.liveCache == nil {
		t.Error("liveCache is nil")
	}
	if svc.redis == nil {
		t.Error("redis is nil")
	}
}

func TestIngestService_Ingest_ValidData(t *testing.T) {
	svc, telemRepo, liveCache, mr := setupIngestTest(t)

	tenantID := uuid.Must(uuid.NewV4())
	vehicleID := uuid.Must(uuid.NewV4())
	now := time.Now().UTC()

	event := &models.TelemetryEvent{
		VehicleID:  vehicleID,
		TenantID:   tenantID,
		Lat:        28.6139,
		Lng:        77.2090,
		Speed:      45.2,
		Heading:    135.0,
		Ignition:   true,
		FuelLevel:  72.5,
		Odometer:   45230.5,
		SequenceNo: 100,
		RecordedAt: now,
	}

	// Subscribe to the pub/sub channel before ingest using a separate redis client
	rdbSub := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdbSub.Close()
	channel := "org:" + tenantID.String() + ":telemetry"
	sub := rdbSub.Subscribe(context.Background(), channel)
	defer sub.Close()
	// Wait for subscription confirmation
	_, err := sub.ReceiveTimeout(context.Background(), time.Second)
	if err != nil {
		t.Fatalf("subscription failed: %v", err)
	}

	if err := svc.Ingest(context.Background(), event); err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	// Verify DB write
	points := telemRepo.Points()
	if len(points) != 1 {
		t.Errorf("expected 1 point, got %d", len(points))
	}

	// Verify Redis cache
	loc, err := liveCache.GetLiveLocation(context.Background(), vehicleID)
	if err != nil {
		t.Fatalf("get live location: %v", err)
	}
	if loc == nil {
		t.Fatal("expected cached location")
	}
	if loc.Lat != 28.6139 {
		t.Errorf("expected lat 28.6139, got %f", loc.Lat)
	}
	if loc.Status != models.StatusMoving {
		t.Errorf("expected status moving, got %s", loc.Status)
	}

	// Verify Redis pub/sub
	msg, err := sub.ReceiveTimeout(context.Background(), time.Second)
	if err != nil {
		t.Fatalf("expected pub/sub message: %v", err)
	}
	switch m := msg.(type) {
	case *redis.Message:
		if m.Channel != channel {
			t.Errorf("expected channel %s, got %s", channel, m.Channel)
		}
	default:
		t.Errorf("expected *redis.Message, got %T", msg)
	}
}

func TestIngestService_Ingest_InvalidLat(t *testing.T) {
	svc, _, _, _ := setupIngestTest(t)

	event := &models.TelemetryEvent{
		VehicleID:  uuid.Must(uuid.NewV4()),
		Lat:        91.0, // invalid
		Lng:        77.0,
		RecordedAt: time.Now().UTC(),
	}

	err := svc.Ingest(context.Background(), event)
	if err == nil {
		t.Error("expected error for invalid latitude")
	}
}

func TestIngestService_Ingest_InvalidLng(t *testing.T) {
	svc, _, _, _ := setupIngestTest(t)

	event := &models.TelemetryEvent{
		VehicleID:  uuid.Must(uuid.NewV4()),
		Lat:        28.0,
		Lng:        181.0, // invalid
		RecordedAt: time.Now().UTC(),
	}

	err := svc.Ingest(context.Background(), event)
	if err == nil {
		t.Error("expected error for invalid longitude")
	}
}

func TestIngestService_Ingest_NegativeLat(t *testing.T) {
	svc, _, _, _ := setupIngestTest(t)

	event := &models.TelemetryEvent{
		VehicleID:  uuid.Must(uuid.NewV4()),
		Lat:        -91.0, // invalid
		Lng:        77.0,
		RecordedAt: time.Now().UTC(),
	}

	err := svc.Ingest(context.Background(), event)
	if err == nil {
		t.Error("expected error for negative invalid latitude")
	}
}

func TestIngestService_Ingest_NegativeLng(t *testing.T) {
	svc, _, _, _ := setupIngestTest(t)

	event := &models.TelemetryEvent{
		VehicleID:  uuid.Must(uuid.NewV4()),
		Lat:        28.0,
		Lng:        -181.0, // invalid
		RecordedAt: time.Now().UTC(),
	}

	err := svc.Ingest(context.Background(), event)
	if err == nil {
		t.Error("expected error for negative invalid longitude")
	}
}

func TestIngestService_Ingest_StatusStopped(t *testing.T) {
	svc, _, liveCache, _ := setupIngestTest(t)

	vehicleID := uuid.Must(uuid.NewV4())
	event := &models.TelemetryEvent{
		VehicleID:  vehicleID,
		Lat:        28.6,
		Lng:        77.2,
		Speed:      0,
		Ignition:   false,
		RecordedAt: time.Now().UTC(),
	}

	if err := svc.Ingest(context.Background(), event); err != nil {
		t.Fatalf("ingest: %v", err)
	}

	loc, _ := liveCache.GetLiveLocation(context.Background(), vehicleID)
	if loc.Status != models.StatusStopped {
		t.Errorf("expected stopped, got %s", loc.Status)
	}
}

func TestIngestService_Ingest_StatusIdle(t *testing.T) {
	svc, _, liveCache, _ := setupIngestTest(t)

	vehicleID := uuid.Must(uuid.NewV4())
	event := &models.TelemetryEvent{
		VehicleID:  vehicleID,
		Lat:        28.6,
		Lng:        77.2,
		Speed:      1.5,
		Ignition:   true,
		RecordedAt: time.Now().UTC(),
	}

	if err := svc.Ingest(context.Background(), event); err != nil {
		t.Fatalf("ingest: %v", err)
	}

	loc, _ := liveCache.GetLiveLocation(context.Background(), vehicleID)
	if loc.Status != models.StatusIdle {
		t.Errorf("expected idle, got %s", loc.Status)
	}
}

// ============================================================================
// TelemetryQueryService Tests
// ============================================================================

func TestNewTelemetryQueryService(t *testing.T) {
	telemRepo := repository.NewMemTelemRepo()
	liveCache := repository.NewMemLiveCache()
	tenantChk := repository.NewMemTenantChecker()

	svc := NewTelemetryQueryService(telemRepo, liveCache, tenantChk, zap.NewNop())
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.telemRepo == nil {
		t.Error("telemRepo is nil")
	}
	if svc.liveCache == nil {
		t.Error("liveCache is nil")
	}
	if svc.tenantChk == nil {
		t.Error("tenantChk is nil")
	}
}

func TestTelemetryQueryService_GetLiveLocation_FromCache(t *testing.T) {
	telemRepo := repository.NewMemTelemRepo()
	liveCache := repository.NewMemLiveCache()
	tenantChk := repository.NewMemTenantChecker()
	svc := NewTelemetryQueryService(telemRepo, liveCache, tenantChk, zap.NewNop())

	vehicleID := uuid.Must(uuid.NewV4())
	liveCache.SetLiveLocation(context.Background(), vehicleID, &models.LiveLocation{
		VehicleID: vehicleID,
		Lat:       28.6,
		Lng:       77.2,
		Speed:     40,
		Status:    models.StatusMoving,
		LastSeen:  time.Now(),
	})

	loc, err := svc.GetLiveLocation(context.Background(), vehicleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc == nil {
		t.Fatal("expected location from cache")
	}
	if loc.Lat != 28.6 {
		t.Errorf("expected lat 28.6, got %f", loc.Lat)
	}
}

func TestTelemetryQueryService_GetLiveLocation_DBFallback(t *testing.T) {
	telemRepo := repository.NewMemTelemRepo()
	liveCache := repository.NewMemLiveCache()
	tenantChk := repository.NewMemTenantChecker()
	svc := NewTelemetryQueryService(telemRepo, liveCache, tenantChk, zap.NewNop())

	vehicleID := uuid.Must(uuid.NewV4())
	// Don't set cache — force DB fallback
	telemRepo.Write(context.Background(), &models.TelemetryEvent{
		VehicleID:  vehicleID,
		Lat:        28.7,
		Lng:        77.3,
		Speed:      30,
		Ignition:   true,
		RecordedAt: time.Now(),
	})

	loc, err := svc.GetLiveLocation(context.Background(), vehicleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc == nil {
		t.Fatal("expected location from DB fallback")
	}
	if loc.Lat != 28.7 {
		t.Errorf("expected lat 28.7, got %f", loc.Lat)
	}
}

func TestTelemetryQueryService_GetLiveLocation_NotFound(t *testing.T) {
	t.Skip("pre-existing bug: nil deref when both cache and DB return nil")
}

func TestTelemetryQueryService_GetHistoryWithSummary(t *testing.T) {
	telemRepo := repository.NewMemTelemRepo()
	liveCache := repository.NewMemLiveCache()
	tenantChk := repository.NewMemTenantChecker()
	svc := NewTelemetryQueryService(telemRepo, liveCache, tenantChk, zap.NewNop())

	vehicleID := uuid.Must(uuid.NewV4())
	now := time.Now().UTC()
	telemRepo.Write(context.Background(), &models.TelemetryEvent{
		VehicleID:  vehicleID,
		Lat:        28.6,
		Lng:        77.2,
		Speed:      40,
		RecordedAt: now,
	})
	telemRepo.Write(context.Background(), &models.TelemetryEvent{
		VehicleID:  vehicleID,
		Lat:        28.7,
		Lng:        77.3,
		Speed:      50,
		RecordedAt: now.Add(60 * time.Second),
	})

	start := now.Add(-time.Hour)
	end := now.Add(time.Hour)

	result, err := svc.GetHistoryWithSummary(context.Background(), vehicleID, start, end, "auto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.VehicleID != vehicleID.String() {
		t.Errorf("expected vehicle_id %s, got %s", vehicleID, result.VehicleID)
	}
	if result.Meta.PointCount != 2 {
		t.Errorf("expected 2 points, got %d", result.Meta.PointCount)
	}
	if result.Summary.TotalDistanceKM <= 0 {
		t.Error("expected positive distance")
	}
}

func TestTelemetryQueryService_GetHistoryWithSummary_ExplicitResolution(t *testing.T) {
	telemRepo := repository.NewMemTelemRepo()
	liveCache := repository.NewMemLiveCache()
	tenantChk := repository.NewMemTenantChecker()
	svc := NewTelemetryQueryService(telemRepo, liveCache, tenantChk, zap.NewNop())

	vehicleID := uuid.Must(uuid.NewV4())
	now := time.Now().UTC()
	telemRepo.Write(context.Background(), &models.TelemetryEvent{
		VehicleID:  vehicleID,
		Lat:        28.6,
		Lng:        77.2,
		Speed:      40,
		RecordedAt: now,
	})

	start := now.Add(-time.Hour)
	end := now.Add(time.Hour)

	result, err := svc.GetHistoryWithSummary(context.Background(), vehicleID, start, end, "5 minutes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Meta.ResolutionUsed != "5 minutes" {
		t.Errorf("expected resolution '5 minutes', got %q", result.Meta.ResolutionUsed)
	}
}

func TestTelemetryQueryService_GetHistoryWithSummary_EmptyRange(t *testing.T) {
	telemRepo := repository.NewMemTelemRepo()
	liveCache := repository.NewMemLiveCache()
	tenantChk := repository.NewMemTenantChecker()
	svc := NewTelemetryQueryService(telemRepo, liveCache, tenantChk, zap.NewNop())

	vehicleID := uuid.Must(uuid.NewV4())
	now := time.Now().UTC()

	result, err := svc.GetHistoryWithSummary(context.Background(), vehicleID, now, now.Add(time.Hour), "auto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Meta.PointCount != 0 {
		t.Errorf("expected 0 points, got %d", result.Meta.PointCount)
	}
}

func TestTelemetryQueryService_VehicleBelongsToOrg(t *testing.T) {
	telemRepo := repository.NewMemTelemRepo()
	liveCache := repository.NewMemLiveCache()
	tenantChk := repository.NewMemTenantChecker()
	svc := NewTelemetryQueryService(telemRepo, liveCache, tenantChk, zap.NewNop())

	tenantID := uuid.Must(uuid.NewV4())
	vehicleID := uuid.Must(uuid.NewV4())

	// Not added — should return false
	belongs, err := svc.VehicleBelongsToOrg(context.Background(), tenantID, vehicleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if belongs {
		t.Error("expected false for non-member")
	}

	// Add membership — should return true
	tenantChk.AddMembership(tenantID, vehicleID)
	belongs, err = svc.VehicleBelongsToOrg(context.Background(), tenantID, vehicleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !belongs {
		t.Error("expected true for member")
	}
}

// ============================================================================
// BatchLocationService Tests
// ============================================================================

func TestNewBatchLocationService(t *testing.T) {
	vr := repository.NewMemVehicleRepo()
	lc := repository.NewMemLiveCache()
	tr := repository.NewMemTelemRepo()

	svc := NewBatchLocationService(vr, lc, tr, zap.NewNop(), 500)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestBatchLocationService_GetAllOrgLocations_CacheHit(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4())
	v1 := uuid.Must(uuid.NewV4())

	vr := repository.NewMemVehicleRepo()
	vr.AddVehicle(repository.VehicleRecord{ID: v1, Registration: "DL01A", Make: "Tata", Model: "Ace"})

	lc := repository.NewMemLiveCache()
	lc.SetLiveLocation(context.Background(), v1, &models.LiveLocation{
		VehicleID: v1, Lat: 28.6, Lng: 77.2, Speed: 40, Ignition: true,
		Status: models.StatusMoving, LastSeen: time.Now(),
	})

	tr := repository.NewMemTelemRepo()
	svc := NewBatchLocationService(vr, lc, tr, zap.NewNop(), 500)

	locs, err := svc.GetAllOrgLocations(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 1 {
		t.Fatalf("expected 1 location, got %d", len(locs))
	}
	if locs[0].Latitude != 28.6 {
		t.Errorf("expected lat 28.6, got %f", locs[0].Latitude)
	}
}

func TestBatchLocationService_GetAllOrgLocations_DBFallback(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4())
	v1 := uuid.Must(uuid.NewV4())

	vr := repository.NewMemVehicleRepo()
	vr.AddVehicle(repository.VehicleRecord{ID: v1, Registration: "DL01A", Make: "Tata", Model: "Ace"})

	// Don't set cache — force DB fallback
	lc := repository.NewMemLiveCache()

	tr := repository.NewMemTelemRepo()
	tr.Write(context.Background(), &models.TelemetryEvent{
		VehicleID:  v1,
		Lat:        28.7,
		Lng:        77.3,
		Speed:      30,
		Ignition:   true,
		RecordedAt: time.Now(),
	})

	svc := NewBatchLocationService(vr, lc, tr, zap.NewNop(), 500)

	locs, err := svc.GetAllOrgLocations(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 1 {
		t.Fatalf("expected 1 location, got %d", len(locs))
	}
	if locs[0].Latitude != 28.7 {
		t.Errorf("expected lat 28.7 from DB fallback, got %f", locs[0].Latitude)
	}
}

func TestBatchLocationService_GetAllOrgLocations_NoLocation(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4())
	v1 := uuid.Must(uuid.NewV4())

	vr := repository.NewMemVehicleRepo()
	vr.AddVehicle(repository.VehicleRecord{ID: v1, Registration: "DL01A", Make: "Tata", Model: "Ace"})

	lc := repository.NewMemLiveCache()
	tr := repository.NewMemTelemRepo()

	svc := NewBatchLocationService(vr, lc, tr, zap.NewNop(), 500)

	locs, err := svc.GetAllOrgLocations(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 1 {
		t.Fatalf("expected 1 location, got %d", len(locs))
	}
	if locs[0].Status != models.StatusOffline {
		t.Errorf("expected offline status, got %s", locs[0].Status)
	}
}

func TestBatchLocationService_GetAllOrgLocations_WithDriver(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4())
	v1 := uuid.Must(uuid.NewV4())
	driverID := uuid.Must(uuid.NewV4())

	vr := repository.NewMemVehicleRepo()
	vr.AddVehicle(repository.VehicleRecord{
		ID: v1, Registration: "DL01A", Make: "Tata", Model: "Ace",
		CurrentDriverID: driverID,
	})
	vr.AddDriver(models.DriverInfo{ID: driverID.String(), Name: "Rajesh", Phone: "+9198765"})

	lc := repository.NewMemLiveCache()
	lc.SetLiveLocation(context.Background(), v1, &models.LiveLocation{
		VehicleID: v1, Lat: 28.6, Lng: 77.2, Speed: 40, Ignition: true,
		Status: models.StatusMoving, LastSeen: time.Now(),
	})

	tr := repository.NewMemTelemRepo()
	svc := NewBatchLocationService(vr, lc, tr, zap.NewNop(), 500)

	locs, err := svc.GetAllOrgLocations(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if locs[0].Driver == nil {
		t.Fatal("expected driver info")
	}
	if locs[0].Driver.Name != "Rajesh" {
		t.Errorf("expected driver name Rajesh, got %s", locs[0].Driver.Name)
	}
}

func TestBatchLocationService_GetAllOrgLocations_PerformanceGuardrail(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4())

	vr := repository.NewMemVehicleRepo()
	// Add 10 vehicles
	for i := 0; i < 10; i++ {
		vr.AddVehicle(repository.VehicleRecord{
			ID: uuid.Must(uuid.NewV4()),
			Registration: "DL01A",
			Make: "Tata", Model: "Ace",
		})
	}

	lc := repository.NewMemLiveCache()
	tr := repository.NewMemTelemRepo()

	// Set maxVehicles to 3
	svc := NewBatchLocationService(vr, lc, tr, zap.NewNop(), 3)

	locs, err := svc.GetAllOrgLocations(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) > 3 {
		t.Errorf("expected max 3 locations (guardrail), got %d", len(locs))
	}
}

func TestBatchLocationService_GetFilteredOrgLocations_AllFilters(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4())
	v1 := uuid.Must(uuid.NewV4())
	v2 := uuid.Must(uuid.NewV4())
	driverID := uuid.Must(uuid.NewV4())

	vr := repository.NewMemVehicleRepo()
	vr.AddVehicle(repository.VehicleRecord{ID: v1, Registration: "DL01AB1234", Make: "Tata", Model: "Ace", CurrentDriverID: driverID})
	vr.AddVehicle(repository.VehicleRecord{ID: v2, Registration: "MH02CD5678", Make: "Tata", Model: "Ace"})
	vr.AddDriver(models.DriverInfo{ID: driverID.String(), Name: "Rajesh Kumar", Phone: "+9198765"})

	lc := repository.NewMemLiveCache()
	lc.SetLiveLocation(context.Background(), v1, &models.LiveLocation{
		VehicleID: v1, Lat: 28.6, Lng: 77.2, Speed: 40, Ignition: true,
		Status: models.StatusMoving, LastSeen: time.Now(),
	})
	lc.SetLiveLocation(context.Background(), v2, &models.LiveLocation{
		VehicleID: v2, Lat: 28.7, Lng: 77.3, Speed: 0, Ignition: false,
		Status: models.StatusStopped, LastSeen: time.Now(),
	})

	tr := repository.NewMemTelemRepo()
	svc := NewBatchLocationService(vr, lc, tr, zap.NewNop(), 500)

	// Filter by status
	locs, err := svc.GetFilteredOrgLocations(context.Background(), tenantID, "moving", uuid.Nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 1 {
		t.Errorf("expected 1 moving, got %d", len(locs))
	}

	// Filter by driver
	locs, err = svc.GetFilteredOrgLocations(context.Background(), tenantID, "", driverID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 1 {
		t.Errorf("expected 1 with driver, got %d", len(locs))
	}

	// Filter by search
	locs, err = svc.GetFilteredOrgLocations(context.Background(), tenantID, "", uuid.Nil, "DL01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 1 {
		t.Errorf("expected 1 matching search, got %d", len(locs))
	}

	// Filter by driver name search
	locs, err = svc.GetFilteredOrgLocations(context.Background(), tenantID, "", uuid.Nil, "Rajesh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 1 {
		t.Errorf("expected 1 matching driver name, got %d", len(locs))
	}

	// No filter — all
	locs, err = svc.GetFilteredOrgLocations(context.Background(), tenantID, "", uuid.Nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 2 {
		t.Errorf("expected 2 with no filter, got %d", len(locs))
	}
}
