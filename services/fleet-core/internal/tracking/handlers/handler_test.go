package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/flyfleet/fleet-core/internal/tracking/models"
	"github.com/flyfleet/fleet-core/internal/tracking/repository"
	"github.com/flyfleet/fleet-core/internal/tracking/service"
)

// ============================================================================
// Service Factory Helpers (using in-memory repos)
// ============================================================================

// withChiParam sets a chi URL parameter on the request context.
func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func makeLocationsHandler(t *testing.T, vehicleRepo repository.VehicleRepositoryer, liveCache repository.LiveCacher, telemRepo repository.TelemetryRepositoryer) *LocationsHandler {
	t.Helper()
	bs := service.NewBatchLocationService(vehicleRepo, liveCache, telemRepo, zap.NewNop(), 500)
	return NewLocationsHandler(bs, zap.NewNop())
}

func makeHistoryHandler(t *testing.T, telemRepo repository.TelemetryRepositoryer, liveCache repository.LiveCacher, tenantChk repository.TenantChecker) *HistoryHandler {
	t.Helper()
	qs := service.NewTelemetryQueryService(telemRepo, liveCache, tenantChk, zap.NewNop())
	return NewHistoryHandler(qs, zap.NewNop())
}

func makeIngestHandlerSimple(telemRepo repository.TelemetryRepositoryer, liveCache repository.LiveCacher) *IngestHandler {
	svc := &simpleIngestSvc{telemRepo: telemRepo, liveCache: liveCache}
	return &IngestHandler{ingestion: svc, logger: zap.NewNop()}
}

// simpleIngestSvc is a minimal IngestionService implementation for testing.
type simpleIngestSvc struct {
	telemRepo repository.TelemetryRepositoryer
	liveCache repository.LiveCacher
}

func (s *simpleIngestSvc) Ingest(ctx context.Context, event *models.TelemetryEvent) error {
	if err := s.telemRepo.Write(ctx, event); err != nil {
		return err
	}
	status := models.CalculateStatus(event.Speed, event.Ignition, event.RecordedAt)
	loc := &models.LiveLocation{
		VehicleID: event.VehicleID,
		Lat:       event.Lat,
		Lng:       event.Lng,
		Speed:     event.Speed,
		Heading:   event.Heading,
		Ignition:  event.Ignition,
		FuelLevel: event.FuelLevel,
		Odometer:  event.Odometer,
		LastSeen:  event.RecordedAt,
		Status:    status,
	}
	s.liveCache.SetLiveLocation(ctx, event.VehicleID, loc)
	return nil
}

// ============================================================================
// Locations Handler Tests
// ============================================================================

func TestLocationsHandler_MethodNotAllowed(t *testing.T) {
	vr := repository.NewMemVehicleRepo()
	lc := repository.NewMemLiveCache()
	tr := repository.NewMemTelemRepo()
	h := makeLocationsHandler(t, vr, lc, tr)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/locations", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestLocationsHandler_MissingOrg(t *testing.T) {
	vr := repository.NewMemVehicleRepo()
	lc := repository.NewMemLiveCache()
	tr := repository.NewMemTelemRepo()
	h := makeLocationsHandler(t, vr, lc, tr)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/locations", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLocationsHandler_EmptyOrg(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4())
	vr := repository.NewMemVehicleRepo()
	lc := repository.NewMemLiveCache()
	tr := repository.NewMemTelemRepo()
	h := makeLocationsHandler(t, vr, lc, tr)

	ctx := ctxkeys.WithTenantID(context.Background(), tenantID.String())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/locations", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	meta := resp["meta"].(map[string]interface{})
	if int(meta["total"].(float64)) != 0 {
		t.Errorf("expected total=0, got %v", meta["total"])
	}
}

func TestLocationsHandler_WithVehicles(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4())
	v1 := uuid.Must(uuid.NewV4())
	v2 := uuid.Must(uuid.NewV4())

	vr := repository.NewMemVehicleRepo()
	vr.AddVehicle(repository.VehicleRecord{ID: v1, Registration: "DL01A", Make: "Tata", Model: "Ace"})
	vr.AddVehicle(repository.VehicleRecord{ID: v2, Registration: "DL02B", Make: "Tata", Model: "Ace"})

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
	h := makeLocationsHandler(t, vr, lc, tr)

	ctx := ctxkeys.WithTenantID(context.Background(), tenantID.String())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/locations", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	data := resp["data"].([]interface{})
	if len(data) != 2 {
		t.Fatalf("expected 2 vehicles, got %d", len(data))
	}

	meta := resp["meta"].(map[string]interface{})
	if int(meta["total"].(float64)) != 2 {
		t.Errorf("expected total=2, got %v", meta["total"])
	}
	if int(meta["moving"].(float64)) != 1 {
		t.Errorf("expected moving=1, got %v", meta["moving"])
	}
	if int(meta["stopped"].(float64)) != 1 {
		t.Errorf("expected stopped=1, got %v", meta["stopped"])
	}
	if int(meta["online"].(float64)) != 2 {
		t.Errorf("expected online=2, got %v", meta["online"])
	}
}

func TestLocationsHandler_StatusFilter(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4())
	v1 := uuid.Must(uuid.NewV4())
	v2 := uuid.Must(uuid.NewV4())

	vr := repository.NewMemVehicleRepo()
	vr.AddVehicle(repository.VehicleRecord{ID: v1, Registration: "DL01A", Make: "Tata", Model: "Ace"})
	vr.AddVehicle(repository.VehicleRecord{ID: v2, Registration: "DL02B", Make: "Tata", Model: "Ace"})

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
	h := makeLocationsHandler(t, vr, lc, tr)

	ctx := ctxkeys.WithTenantID(context.Background(), tenantID.String())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/locations?status=moving", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	meta := resp["meta"].(map[string]interface{})
	if int(meta["total"].(float64)) != 1 {
		t.Errorf("expected total=1 with status filter, got %v", meta["total"])
	}
}

func TestLocationsHandler_SearchFilter(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4())
	v1 := uuid.Must(uuid.NewV4())
	v2 := uuid.Must(uuid.NewV4())

	vr := repository.NewMemVehicleRepo()
	vr.AddVehicle(repository.VehicleRecord{ID: v1, Registration: "DL01AB1234", Make: "Tata", Model: "Ace"})
	vr.AddVehicle(repository.VehicleRecord{ID: v2, Registration: "MH02CD5678", Make: "Tata", Model: "Ace"})

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
	h := makeLocationsHandler(t, vr, lc, tr)

	ctx := ctxkeys.WithTenantID(context.Background(), tenantID.String())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/locations?search=DL01", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	meta := resp["meta"].(map[string]interface{})
	if int(meta["total"].(float64)) != 1 {
		t.Errorf("expected total=1 with search, got %v", meta["total"])
	}
}

// ============================================================================
// History Handler Tests
// ============================================================================

func TestHistoryHandler_MethodNotAllowed(t *testing.T) {
	tr := repository.NewMemTelemRepo()
	lc := repository.NewMemLiveCache()
	tc := repository.NewMemTenantChecker()
	h := makeHistoryHandler(t, tr, lc, tc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles/test/telemetry/history", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHistoryHandler_MissingVehicleID(t *testing.T) {
	tr := repository.NewMemTelemRepo()
	lc := repository.NewMemLiveCache()
	tc := repository.NewMemTenantChecker()
	h := makeHistoryHandler(t, tr, lc, tc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles//telemetry/history", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHistoryHandler_InvalidVehicleID(t *testing.T) {
	tr := repository.NewMemTelemRepo()
	lc := repository.NewMemLiveCache()
	tc := repository.NewMemTenantChecker()
	h := makeHistoryHandler(t, tr, lc, tc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/not-a-uuid/telemetry/history", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHistoryHandler_MissingOrg(t *testing.T) {
	tr := repository.NewMemTelemRepo()
	lc := repository.NewMemLiveCache()
	tc := repository.NewMemTenantChecker()
	h := makeHistoryHandler(t, tr, lc, tc)

	vehicleID := uuid.Must(uuid.NewV4())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/"+vehicleID.String()+"/telemetry/history", nil)
	req = withChiParam(req, "id", vehicleID.String())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHistoryHandler_TenantIsolationViolation(t *testing.T) {
	tr := repository.NewMemTelemRepo()
	lc := repository.NewMemLiveCache()
	tc := repository.NewMemTenantChecker()
	h := makeHistoryHandler(t, tr, lc, tc)

	vehicleID := uuid.Must(uuid.NewV4())
	tenantID := uuid.Must(uuid.NewV4())
	// Don't add membership — should fail
	start := time.Now().Add(-time.Hour).Format(time.RFC3339)
	end := time.Now().Format(time.RFC3339)
	ctx := ctxkeys.WithTenantID(context.Background(), tenantID.String())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/"+vehicleID.String()+"/telemetry/history?start="+start+"&end="+end, nil).WithContext(ctx)
	req = withChiParam(req, "id", vehicleID.String())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 tenant isolation, got %d", w.Code)
	}
}

func TestHistoryHandler_Success(t *testing.T) {
	tr := repository.NewMemTelemRepo()
	lc := repository.NewMemLiveCache()
	tc := repository.NewMemTenantChecker()
	h := makeHistoryHandler(t, tr, lc, tc)

	vehicleID := uuid.Must(uuid.NewV4())
	tenantID := uuid.Must(uuid.NewV4())
	tc.AddMembership(tenantID, vehicleID)

	now := time.Now().UTC()
	tr.Write(context.Background(), &models.TelemetryEvent{
		VehicleID: vehicleID, Lat: 28.6, Lng: 77.2, Speed: 40,
		RecordedAt: now,
	})
	tr.Write(context.Background(), &models.TelemetryEvent{
		VehicleID: vehicleID, Lat: 28.7, Lng: 77.3, Speed: 50,
		RecordedAt: now.Add(60 * time.Second),
	})

	start := now.Add(-time.Hour).Format(time.RFC3339)
	end := now.Add(time.Hour).Format(time.RFC3339)
	ctx := ctxkeys.WithTenantID(context.Background(), tenantID.String())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/vehicles/"+vehicleID.String()+"/telemetry/history?start="+start+"&end="+end, nil).WithContext(ctx)
	req = withChiParam(req, "id", vehicleID.String())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["vehicle_id"] != vehicleID.String() {
		t.Errorf("expected vehicle_id=%s, got %v", vehicleID, resp["vehicle_id"])
	}
	if resp["meta"].(map[string]interface{})["point_count"].(float64) != 2 {
		t.Errorf("expected 2 points, got %v", resp["meta"])
	}
}

// ============================================================================
// Ingest Handler Tests
// ============================================================================

func TestIngestHandler_MethodNotAllowed(t *testing.T) {
	tr := repository.NewMemTelemRepo()
	lc := repository.NewMemLiveCache()
	h := makeIngestHandlerSimple(tr, lc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/telemetry", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestIngestHandler_InvalidJSON(t *testing.T) {
	tr := repository.NewMemTelemRepo()
	lc := repository.NewMemLiveCache()
	h := makeIngestHandlerSimple(tr, lc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry", bytes.NewReader([]byte(`{bad`)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIngestHandler_InvalidVehicleID(t *testing.T) {
	tr := repository.NewMemTelemRepo()
	lc := repository.NewMemLiveCache()
	h := makeIngestHandlerSimple(tr, lc)

	payload := `{"vehicle_id":"bad","latitude":28.6,"longitude":77.2}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry", bytes.NewReader([]byte(payload)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestIngestHandler_AcceptsValidPayload(t *testing.T) {
	tr := repository.NewMemTelemRepo()
	lc := repository.NewMemLiveCache()
	h := makeIngestHandlerSimple(tr, lc)

	payload := `{
		"vehicle_id": "550e8400-e29b-41d4-a716-446655440000",
		"device_id": "550e8400-e29b-41d4-a716-446655440001",
		"latitude": 28.6139,
		"longitude": 77.2090,
		"speed_kmh": 45.2,
		"heading": 135.0,
		"ignition": true,
		"fuel_level": 72.5,
		"odometer": 45230.5,
		"sequence_no": 100,
		"recorded_at": "2026-04-08T10:30:00Z"
	}`
	ctx := ctxkeys.WithTenantID(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry", bytes.NewReader([]byte(payload))).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	if len(tr.Points()) != 1 {
		t.Errorf("expected 1 point stored, got %d", len(tr.Points()))
	}
	vID := uuid.Must(uuid.FromString("550e8400-e29b-41d4-a716-446655440000"))
	loc, _ := lc.GetLiveLocation(context.Background(), vID)
	if loc == nil {
		t.Error("expected live location cached")
	}
}

// ============================================================================
// Geofence Handler Tests
// ============================================================================

func TestGeofenceHandler_GetSuccess(t *testing.T) {
	gID := uuid.Must(uuid.NewV4())
	tenantID := uuid.Must(uuid.NewV4())
	repo := repository.NewMemGeofenceRepo()
	repo.AddGeofence(models.Geofence{ID: gID, TenantID: tenantID, Name: "Warehouse A", Type: "circle"})
	h := NewGeofenceHandler(repo, zap.NewNop())

	ctx := ctxkeys.WithTenantID(context.Background(), tenantID.String())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/geofences", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	data := resp["data"].([]interface{})
	if len(data) != 1 {
		t.Errorf("expected 1 geofence, got %d", len(data))
	}
}

func TestGeofenceHandler_GetEmpty(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4())
	repo := repository.NewMemGeofenceRepo()
	h := NewGeofenceHandler(repo, zap.NewNop())

	ctx := ctxkeys.WithTenantID(context.Background(), tenantID.String())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/geofences", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGeofenceHandler_PostMissingName(t *testing.T) {
	repo := repository.NewMemGeofenceRepo()
	h := NewGeofenceHandler(repo, zap.NewNop())

	tenantID := uuid.Must(uuid.NewV4())
	payload := `{"type":"circle"}`
	ctx := ctxkeys.WithTenantID(context.Background(), tenantID.String())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/geofences", bytes.NewReader([]byte(payload))).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGeofenceHandler_PostCreatesGeofence(t *testing.T) {
	repo := repository.NewMemGeofenceRepo()
	h := NewGeofenceHandler(repo, zap.NewNop())

	tenantID := uuid.Must(uuid.NewV4())
	payload := `{"name":"Test Zone","type":"circle","center_lat":28.6139,"center_lng":77.2090,"radius_m":1000}`
	ctx := ctxkeys.WithTenantID(context.Background(), tenantID.String())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/geofences", bytes.NewReader([]byte(payload))).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "Test Zone" {
		t.Errorf("expected name=Test Zone, got %v", data["name"])
	}
	if len(repo.Geofences()) != 1 {
		t.Errorf("expected 1 geofence stored, got %d", len(repo.Geofences()))
	}
}

func TestGeofenceHandler_MethodNotAllowed(t *testing.T) {
	repo := repository.NewMemGeofenceRepo()
	h := NewGeofenceHandler(repo, zap.NewNop())

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/geofences", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

// ============================================================================
// WebSocket Message Tests
// ============================================================================

func TestWSMessage_PingConstants(t *testing.T) {
	if models.WSMsgPing != "ping" {
		t.Errorf("WSMsgPing=%q, want 'ping'", models.WSMsgPing)
	}
	if models.WSMsgPong != "pong" {
		t.Errorf("WSMsgPong=%q, want 'pong'", models.WSMsgPong)
	}
	if models.WSMsgGeofenceAlert != "geofence_alert" {
		t.Errorf("WSMsgGeofenceAlert=%q, want 'geofence_alert'", models.WSMsgGeofenceAlert)
	}
}

func TestWSMessage_AllTypesMarshal(t *testing.T) {
	tests := []struct {
		name string
		msg  models.WSMessage
	}{
		{"ping", models.WSMessage{Type: models.WSMsgPing}},
		{"pong", models.WSMessage{Type: models.WSMsgPong, Data: map[string]string{"timestamp": "2026-04-08T10:30:00Z"}}},
		{"position_update", models.WSMessage{Type: models.WSMsgPositionUpdate, Data: models.PositionUpdate{VehicleID: "v1", Latitude: 28.6, Longitude: 77.2}}},
		{"geofence_alert", models.WSMessage{Type: models.WSMsgGeofenceAlert, Data: models.GeofenceWSMessage{VehicleID: "v1", GeofenceName: "Test", EventType: "enter"}}},
		{"connection_ack", models.WSMessage{Type: models.WSMsgConnectionAck, Data: map[string]interface{}{"status": "connected"}}},
		{"subscribe_vehicles", models.WSMessage{Type: models.WSMsgSubscribe, Data: models.SubscribeRequest{VehicleIDs: []string{"v1"}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var decoded models.WSMessage
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if decoded.Type != tt.msg.Type {
				t.Errorf("expected type=%s, got %s", tt.msg.Type, decoded.Type)
			}
		})
	}
}
