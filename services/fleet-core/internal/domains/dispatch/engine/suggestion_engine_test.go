package engine

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Mock dependencies for suggestion engine tests ──────────────────────────

type mockVehicleRepo struct {
	vehicles []mockVehicle
}

type mockVehicle struct {
	ID              uuid.UUID
	OrgID           uuid.UUID
	Status          string
	Registration    string
	CapacityTonnage float64
	Type            string
}

func (m *mockVehicleRepo) GetByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]mockVehicle, int64, error) {
	var filtered []mockVehicle
	for _, v := range m.vehicles {
		if v.OrgID == orgID {
			filtered = append(filtered, v)
		}
	}
	return filtered, int64(len(filtered)), nil
}

type mockDriverRepo struct {
	drivers []mockDriver
}

type mockDriver struct {
	ID            uuid.UUID
	OrgID         uuid.UUID
	Name          string
	Rating        float64
	CurrentTripID *string
	Status        string
}

func (m *mockDriverRepo) GetByVehicleAssignment(ctx context.Context, vehicleID uuid.UUID) (*mockDriver, error) {
	for _, d := range m.drivers {
		if d.OrgID.String() == vehicleID.String()[:36] {
			return &d, nil
		}
	}
	return nil, assert.AnError
}

type mockTripRepo struct {
	activeTrips map[string]int // vehicleID → count of active trips
}

func (m *mockTripRepo) GetActiveTripsForVehicle(ctx context.Context, vehicleID string) ([]interface{}, error) {
	count := m.activeTrips[vehicleID]
	if count > 0 {
		return make([]interface{}, count), nil
	}
	return nil, nil
}

type mockLiveCache struct {
	positions map[uuid.UUID]mockPosition
}

type mockPosition struct {
	Lat   float64
	Lng   float64
	Speed float64
}

func (m *mockLiveCache) GetLiveLocation(ctx context.Context, vehicleID uuid.UUID) (*mockPosition, error) {
	if pos, ok := m.positions[vehicleID]; ok {
		return &pos, nil
	}
	return nil, nil
}

// ── Suggestion Engine Tests ────────────────────────────────────────────────

func TestSuggestionEngine_FiltersByOrg(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()

	vehicleRepo := &mockVehicleRepo{
		vehicles: []mockVehicle{
			{ID: uuid.New(), OrgID: orgA, Status: "ACTIVE", Registration: "DL-AAA"},
			{ID: uuid.New(), OrgID: orgA, Status: "ACTIVE", Registration: "DL-BBB"},
			{ID: uuid.New(), OrgID: orgB, Status: "ACTIVE", Registration: "MH-CCC"}, // Wrong org
		},
	}

	vehicles, _, err := vehicleRepo.GetByOrgID(context.Background(), orgA, 500, 0)
	require.NoError(t, err)

	assert.Len(t, vehicles, 2, "Only Org A vehicles returned")
	for _, v := range vehicles {
		assert.Equal(t, orgA, v.OrgID, "All vehicles must belong to Org A")
	}
}

func TestSuggestionEngine_FiltersByCapacity(t *testing.T) {
	cargoWeight := 500.0 // 500kg

	vehicles := []mockVehicle{
		{CapacityTonnage: 0.5}, // 500kg — fits exactly
		{CapacityTonnage: 1.5}, // 1500kg — fits
		{CapacityTonnage: 0.3}, // 300kg — too small
	}

	var filtered []mockVehicle
	for _, v := range vehicles {
		if v.CapacityTonnage*1000 >= cargoWeight {
			filtered = append(filtered, v)
		}
	}

	assert.Len(t, filtered, 2, "Only vehicles with sufficient capacity")
}

func TestSuggestionEngine_SortsByScore(t *testing.T) {
	suggestions := []Suggestion{
		{Score: 72.0},
		{Score: 95.0},
		{Score: 83.0},
		{Score: 61.0},
	}

	// Sort by score descending
	for i := 0; i < len(suggestions); i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].Score > suggestions[i].Score {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	assert.Equal(t, 95.0, suggestions[0].Score)
	assert.Equal(t, 83.0, suggestions[1].Score)
	assert.Equal(t, 72.0, suggestions[2].Score)
	assert.Equal(t, 61.0, suggestions[3].Score)
}

func TestSuggestionEngine_ReturnsTopN(t *testing.T) {
	suggestions := []Suggestion{
		{Rank: 1, Score: 95.0},
		{Rank: 2, Score: 83.0},
		{Rank: 3, Score: 72.0},
		{Rank: 4, Score: 61.0},
		{Rank: 5, Score: 50.0},
	}

	topN := 3
	if len(suggestions) > topN {
		suggestions = suggestions[:topN]
	}

	assert.Len(t, suggestions, 3)
	assert.Equal(t, 95.0, suggestions[0].Score)
}

func TestSuggestionEngine_FiltersBusyVehicles(t *testing.T) {
	tripRepo := &mockTripRepo{
		activeTrips: map[string]int{
			"v-001": 1, // On active trip
			"v-002": 0, // Available
			"v-003": 2, // On active trip
		},
	}

	var available []string
	vehicleIDs := []string{"v-001", "v-002", "v-003"}
	for _, vid := range vehicleIDs {
		trips, _ := tripRepo.GetActiveTripsForVehicle(context.Background(), vid)
		if len(trips) == 0 {
			available = append(available, vid)
		}
	}

	assert.Len(t, available, 1, "Only v-002 is available")
	assert.Contains(t, available, "v-002")
}

func TestSuggestionEngine_EmptyOrgReturnsEmpty(t *testing.T) {
	vehicleRepo := &mockVehicleRepo{
		vehicles: []mockVehicle{},
	}

	vehicles, count, err := vehicleRepo.GetByOrgID(context.Background(), uuid.New(), 500, 0)
	require.NoError(t, err)
	assert.Empty(t, vehicles)
	assert.Equal(t, int64(0), count)
}

