package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/dispatch/engine"
	"github.com/flyfleet/fleet-core/internal/domains/dispatch/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// ── Mock dispatch service for handler tests ─────────────────────────────────

type mockDispatchService struct {
	assignCalled     bool
	lastAssign       service.AssignRequest
	assignResult     *service.AssignResult
	assignErr        error
	redispatchCalled bool
}

func (m *mockDispatchService) Assign(ctx context.Context, orgID uuid.UUID, req service.AssignRequest) (*service.AssignResult, error) {
	m.assignCalled = true
	m.lastAssign = req
	return m.assignResult, m.assignErr
}

func (m *mockDispatchService) Redispatch(ctx context.Context, orgID uuid.UUID, req service.RedispatchRequest) (*service.RedispatchResult, error) {
	m.redispatchCalled = true
	return nil, nil
}

type mockSuggestionEngine struct {
	suggestions []engine.Suggestion
	err         error
}

func (m *mockSuggestionEngine) Suggest(ctx context.Context, req engine.SuggestionRequest) ([]engine.Suggestion, error) {
	return m.suggestions, m.err
}

func testLogger() *zap.Logger {
	l, _ := zap.NewDevelopment()
	return l
}

// ── Handler Tests ────────────────────────────────────────────────────────────

func TestDispatchHandler_Assign_HappyPath(t *testing.T) {
	mockSvc := &mockDispatchService{
		assignResult: &service.AssignResult{
			TripID:         uuid.New().String(),
			Status:         "SCHEDULED",
			ShipmentStatus: "ASSIGNED",
		},
	}

	handler := &DispatchHandler{dispatchSvc: mockSvc, logger: testLogger()}

	body := service.AssignRequest{
		ShipmentID: uuid.New(),
		VehicleID:  uuid.New(),
		DriverID:   uuid.New(),
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch/assign", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "tenant_id", uuid.New().String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.Assign(rr, req)

	assert.True(t, mockSvc.assignCalled, "DispatchService.Assign must be called")
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestDispatchHandler_Assign_InvalidBody(t *testing.T) {
	mockSvc := &mockDispatchService{}
	handler := &DispatchHandler{dispatchSvc: mockSvc, logger: testLogger()}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch/assign",
		bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "tenant_id", uuid.New().String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.Assign(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.False(t, mockSvc.assignCalled, "Assign must NOT be called with invalid body")
}

func TestDispatchHandler_Assign_MissingOrg(t *testing.T) {
	mockSvc := &mockDispatchService{}
	handler := &DispatchHandler{dispatchSvc: mockSvc, logger: testLogger()}

	body := service.AssignRequest{ShipmentID: uuid.New()}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch/assign", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// No tenant_id in context

	rr := httptest.NewRecorder()
	handler.Assign(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, mockSvc.assignCalled)
}

func TestDispatchHandler_Suggest_ReturnsTopN(t *testing.T) {
	shipmentID := uuid.New()
	mockEngine := &mockSuggestionEngine{
		suggestions: []engine.Suggestion{
			{Rank: 1, Score: 95.0},
			{Rank: 2, Score: 82.0},
			{Rank: 3, Score: 71.0},
		},
	}

	handler := &DispatchHandler{suggestEng: mockEngine, logger: testLogger()}

	body := map[string]interface{}{
		"shipment_id": shipmentID.String(),
		"top_n":       3,
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch/suggest", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "tenant_id", uuid.New().String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetSuggestions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestDispatchHandler_Suggest_InvalidShipmentID(t *testing.T) {
	handler := &DispatchHandler{logger: testLogger()}

	body := map[string]interface{}{"shipment_id": "not-a-uuid"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch/suggest", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "tenant_id", uuid.New().String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetSuggestions(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDispatchHandler_Dashboard_DoesNotPanic(t *testing.T) {
	handler := &DispatchHandler{logger: testLogger()}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dispatch/dashboard", nil)
	ctx := context.WithValue(req.Context(), "tenant_id", uuid.New().String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetDashboard(rr, req)

	// Handler should not panic even with nil deps (graceful degradation)
	assert.NotPanics(t, func() {
		handler.GetDashboard(rr, req)
	})
}

func TestDispatchHandler_Redispatch_HappyPath(t *testing.T) {
	mockSvc := &mockDispatchService{}
	handler := &DispatchHandler{dispatchSvc: mockSvc, logger: testLogger()}

	body := service.RedispatchRequest{
		TripID:       uuid.New().String(),
		NewVehicleID: uuid.New(),
		NewDriverID:  uuid.New(),
		Reason:       "Vehicle breakdown",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dispatch/redispatch", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "tenant_id", uuid.New().String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.Redispatch(rr, req)

	assert.True(t, mockSvc.redispatchCalled, "Redispatch must be called")
	assert.NotEqual(t, http.StatusInternalServerError, rr.Code)
}
