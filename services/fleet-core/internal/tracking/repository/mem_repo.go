package repository

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
)

// ============================================================================
// In-Memory Implementations for Unit Testing
// ============================================================================

// MemVehicleRepo is a thread-safe in-memory vehicle store.
type MemVehicleRepo struct {
	mu      sync.RWMutex
	records []VehicleRecord
	drivers map[uuid.UUID]*models.DriverInfo
}

func NewMemVehicleRepo() *MemVehicleRepo {
	return &MemVehicleRepo{drivers: make(map[uuid.UUID]*models.DriverInfo)}
}

func (r *MemVehicleRepo) AddVehicle(v VehicleRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, v)
}

func (r *MemVehicleRepo) AddDriver(d models.DriverInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, _ := uuid.FromString(d.ID)
	r.drivers[id] = &d
}

func (r *MemVehicleRepo) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = nil
	r.drivers = make(map[uuid.UUID]*models.DriverInfo)
}

func (r *MemVehicleRepo) GetOrgVehicles(ctx context.Context, orgID uuid.UUID) ([]VehicleRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []VehicleRecord
	for _, v := range r.records {
		out = append(out, v)
	}
	return out, nil
}

func (r *MemVehicleRepo) GetDriverInfo(ctx context.Context, driverID uuid.UUID) (*models.DriverInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if d, ok := r.drivers[driverID]; ok {
		return d, nil
	}
	return nil, nil
}

func (r *MemVehicleRepo) GetDriversBatch(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*models.DriverInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[uuid.UUID]*models.DriverInfo)
	for _, id := range ids {
		if d, ok := r.drivers[id]; ok {
			result[id] = d
		}
	}
	return result, nil
}

// MemLiveCache is a thread-safe in-memory Redis cache mock.
type MemLiveCache struct {
	mu      sync.RWMutex
	locations map[uuid.UUID]*models.LiveLocation
	geoState  map[string]bool // "vehicleID:geofenceID" -> inside
	published []PublishedMsg
}

type PublishedMsg struct {
	OrgID   uuid.UUID
	Payload []byte
}

func NewMemLiveCache() *MemLiveCache {
	return &MemLiveCache{
		locations: make(map[uuid.UUID]*models.LiveLocation),
		geoState:  make(map[string]bool),
	}
}

func (r *MemLiveCache) SetLiveLocation(ctx context.Context, vehicleID uuid.UUID, loc *models.LiveLocation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.locations[vehicleID] = loc
	return nil
}

func (r *MemLiveCache) GetLiveLocation(ctx context.Context, vehicleID uuid.UUID) (*models.LiveLocation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if loc, ok := r.locations[vehicleID]; ok {
		return loc, nil
	}
	return nil, nil
}

func (r *MemLiveCache) GetBatchLiveLocations(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*models.LiveLocation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[uuid.UUID]*models.LiveLocation)
	for _, id := range ids {
		if loc, ok := r.locations[id]; ok {
			result[id] = loc
		}
	}
	return result, nil
}

func (r *MemLiveCache) PublishPositionUpdate(ctx context.Context, orgID uuid.UUID, msg []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.published = append(r.published, PublishedMsg{OrgID: orgID, Payload: msg})
	return nil
}

func (r *MemLiveCache) GetGeofenceState(ctx context.Context, vehicleID, geofenceID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.geoState[vehicleID+":"+geofenceID], nil
}

func (r *MemLiveCache) SetGeofenceState(ctx context.Context, vehicleID, geofenceID string, inside bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.geoState[vehicleID+":"+geofenceID] = inside
	return nil
}

func (r *MemLiveCache) GetPublished() []PublishedMsg {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]PublishedMsg, len(r.published))
	copy(out, r.published)
	return out
}

func (r *MemLiveCache) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.locations = make(map[uuid.UUID]*models.LiveLocation)
	r.geoState = make(map[string]bool)
	r.published = nil
}

// MemTelemRepo is a thread-safe in-memory telemetry store.
type MemTelemRepo struct {
	mu      sync.RWMutex
	points  []models.TelemetryEvent
	latest  map[uuid.UUID]models.TelemetryEvent
}

func NewMemTelemRepo() *MemTelemRepo {
	return &MemTelemRepo{latest: make(map[uuid.UUID]models.TelemetryEvent)}
}

func (r *MemTelemRepo) Write(ctx context.Context, event *models.TelemetryEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.points = append(r.points, *event)
	r.latest[event.VehicleID] = *event
	return nil
}

func (r *MemTelemRepo) GetLatestByVehicle(ctx context.Context, vehicleID uuid.UUID) (*models.TelemetryEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if t, ok := r.latest[vehicleID]; ok {
		return &t, nil
	}
	return nil, nil
}

func (r *MemTelemRepo) GetBatchLatest(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]models.TelemetryEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[uuid.UUID]models.TelemetryEvent)
	for _, id := range ids {
		if t, ok := r.latest[id]; ok {
			result[id] = t
		}
	}
	return result, nil
}

func (r *MemTelemRepo) GetHistoryBucket(ctx context.Context, vehicleID uuid.UUID, start, end time.Time, bucketInterval string) ([]models.TelemetryEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var filtered []models.TelemetryEvent
	for _, p := range r.points {
		if p.VehicleID == vehicleID && !p.RecordedAt.Before(start) && !p.RecordedAt.After(end) {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

func (r *MemTelemRepo) Points() []models.TelemetryEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]models.TelemetryEvent, len(r.points))
	copy(out, r.points)
	return out
}

func (r *MemTelemRepo) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.points = nil
	r.latest = make(map[uuid.UUID]models.TelemetryEvent)
}

// MemGeofenceRepo is a thread-safe in-memory geofence store.
type MemGeofenceRepo struct {
	mu     sync.RWMutex
	geofences []models.Geofence
	events  []models.GeofenceEvent
}

func NewMemGeofenceRepo() *MemGeofenceRepo {
	return &MemGeofenceRepo{}
}

func (r *MemGeofenceRepo) AddGeofence(g models.Geofence) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.geofences = append(r.geofences, g)
}

func (r *MemGeofenceRepo) GetOrgGeofences(ctx context.Context, orgID uuid.UUID) ([]models.Geofence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []models.Geofence
	for _, g := range r.geofences {
		if g.OrgID == orgID {
			out = append(out, g)
		}
	}
	return out, nil
}

func (r *MemGeofenceRepo) CreateGeofence(ctx context.Context, g *models.Geofence) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.geofences = append(r.geofences, *g)
	return nil
}

func (r *MemGeofenceRepo) CheckPointInGeofences(ctx context.Context, orgID uuid.UUID, lat, lng float64) ([]models.Geofence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Simplified: for unit testing, return geofences that match org
	var out []models.Geofence
	for _, g := range r.geofences {
		if g.OrgID != orgID {
			continue
		}
		// Circle: check distance
		if g.Type == "circle" && g.RadiusM != nil && g.CenterLat != nil && g.CenterLng != nil {
			dist := haversineKmTest(*g.CenterLat, *g.CenterLng, lat, lng) * 1000
			if dist <= *g.RadiusM {
				out = append(out, g)
			}
		}
	}
	return out, nil
}

func (r *MemGeofenceRepo) LogGeofenceEvent(ctx context.Context, event *models.GeofenceEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, *event)
	return nil
}

func (r *MemGeofenceRepo) Events() []models.GeofenceEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]models.GeofenceEvent, len(r.events))
	copy(out, r.events)
	return out
}

func (r *MemGeofenceRepo) Geofences() []models.Geofence {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]models.Geofence, len(r.geofences))
	copy(out, r.geofences)
	return out
}

func (r *MemGeofenceRepo) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.geofences = nil
	r.events = nil
}

// MemTenantChecker is an in-memory tenant checker for testing.
type MemTenantChecker struct {
	mu        sync.RWMutex
	membership map[string]bool // "orgID:vehicleID" -> belongs
}

func NewMemTenantChecker() *MemTenantChecker {
	return &MemTenantChecker{membership: make(map[string]bool)}
}

func (t *MemTenantChecker) AddMembership(orgID, vehicleID uuid.UUID) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.membership[orgID.String()+":"+vehicleID.String()] = true
}

func (t *MemTenantChecker) VehicleBelongsToOrg(ctx context.Context, orgID, vehicleID uuid.UUID) (bool, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.membership[orgID.String()+":"+vehicleID.String()], nil
}

func haversineKmTest(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	a := 0.5 - 0.5*cosT(lat2*0.01745329252-lat1*0.01745329252) +
		cosT(lat1*0.01745329252)*cosT(lat2*0.01745329252)*
			(0.5 - 0.5*cosT(lng2*0.01745329252-lng1*0.01745329252))
	return R * 2 * atan2T(sqrtT(a), sqrtT(1-a))
}

func sqrtT(x float64) float64 {
	r := x
	for i := 0; i < 10; i++ {
		r = (r + x/r) / 2
	}
	return r
}

func atan2T(y, x float64) float64 {
	if x == 0 {
		if y > 0 {
			return 1.5707963267948966
		}
		return -1.5707963267948966
	}
	atan := y / x
	absA := atan
	if absA < 0 {
		absA = -absA
	}
	result := absA / (1 + 0.28*absA*absA)
	if atan < 0 {
		result = -result
	}
	if x < 0 {
		if atan >= 0 {
			result += 3.141592653589793
		} else {
			result -= 3.141592653589793
		}
	}
	return result
}

func cosT(x float64) float64 {
	x2 := x * x
	return 1 - x2/2 + x2*x2/24 - x2*x2*x2/720
}

// ToJSON serializes any value to JSON string (for handler response checks).
func ToJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
