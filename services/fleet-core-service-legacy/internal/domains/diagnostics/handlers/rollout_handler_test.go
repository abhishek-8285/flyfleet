package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/models"
	diagservice "github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type rolloutServiceMock struct {
	diagservice.DeviceDiagnosticsService

	result     *models.FirmwareRollout
	getResult  *models.FirmwareRollout
	getErr     error
	events     []*models.FirmwareRolloutEvent
	eventsErr  error
	err        error
	callCount  int
	getCalls   int
	eventsCall int
	noteCalls  int
	lastRollID uuid.UUID
	lastOrgID  uuid.UUID
	lastNext   string
	lastReason *string
	lastActor  *uuid.UUID
	lastNote   string
	lastEventType string
	lastLimit int
	lastOffset int
}

func (m *rolloutServiceMock) GetRollout(_ context.Context, rolloutID, orgID uuid.UUID) (*models.FirmwareRollout, error) {
	m.getCalls++
	m.lastRollID = rolloutID
	m.lastOrgID = orgID
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.getResult != nil {
		return m.getResult, nil
	}
	return &models.FirmwareRollout{ID: rolloutID, OrgID: orgID, Status: "scheduled"}, nil
}

func (m *rolloutServiceMock) ListRolloutEvents(_ context.Context, rolloutID, orgID uuid.UUID, eventType string, limit, offset int) ([]*models.FirmwareRolloutEvent, error) {
	m.eventsCall++
	m.lastRollID = rolloutID
	m.lastOrgID = orgID
	m.lastEventType = eventType
	m.lastLimit = limit
	m.lastOffset = offset
	if m.eventsErr != nil {
		return nil, m.eventsErr
	}
	return m.events, nil
}

func (m *rolloutServiceMock) AddRolloutNote(_ context.Context, rolloutID, orgID uuid.UUID, actorUserID *uuid.UUID, message string, metadata map[string]interface{}) (*models.FirmwareRolloutEvent, error) {
	m.noteCalls++
	m.lastRollID = rolloutID
	m.lastOrgID = orgID
	m.lastActor = actorUserID
	m.lastNote = message
	if m.err != nil {
		return nil, m.err
	}
	return &models.FirmwareRolloutEvent{
		ID:          uuid.New(),
		RolloutID:   rolloutID,
		OrgID:       orgID,
		EventType:   "rollout.note",
		ActorUserID: actorUserID,
		Message:     message,
		Metadata:    metadata,
	}, nil
}

func (m *rolloutServiceMock) TransitionRollout(_ context.Context, rolloutID, orgID uuid.UUID, nextStatus string, actorUserID *uuid.UUID, reason *string) (*models.FirmwareRollout, error) {
	m.callCount++
	m.lastRollID = rolloutID
	m.lastOrgID = orgID
	m.lastNext = nextStatus
	m.lastActor = actorUserID
	m.lastReason = reason
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return &models.FirmwareRollout{ID: rolloutID, OrgID: orgID, Status: nextStatus}, nil
}

func TestTransitionRollout_InvalidRolloutID(t *testing.T) {
	mockSvc := &rolloutServiceMock{}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodPatch, "/rollouts/not-a-uuid/status", strings.NewReader(`{"status":"in_progress"}`))
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "not-a-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", uuid.NewString())

	h.TransitionRollout(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if mockSvc.callCount != 0 {
		t.Fatalf("expected service not called, got %d calls", mockSvc.callCount)
	}
}

func TestTransitionRollout_EmptyStatusRejected(t *testing.T) {
	mockSvc := &rolloutServiceMock{}
	h := NewHandlers(mockSvc, zap.NewNop())
	rolloutID := uuid.New()

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/rollouts/%s/status", rolloutID), strings.NewReader(`{"status":"   "}`))
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", uuid.NewString())

	h.TransitionRollout(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if mockSvc.callCount != 0 {
		t.Fatalf("expected service not called, got %d calls", mockSvc.callCount)
	}
}

func TestTransitionRollout_ServiceErrorMapsTo400(t *testing.T) {
	mockSvc := &rolloutServiceMock{err: fmt.Errorf("invalid rollout transition")}
	h := NewHandlers(mockSvc, zap.NewNop())
	rolloutID := uuid.New()
	orgID := uuid.New()

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/rollouts/%s/status", rolloutID), strings.NewReader(`{"status":"in_progress"}`))
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", orgID.String())

	h.TransitionRollout(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if mockSvc.callCount != 1 {
		t.Fatalf("expected service call once, got %d", mockSvc.callCount)
	}
	if mockSvc.lastRollID != rolloutID || mockSvc.lastOrgID != orgID || mockSvc.lastNext != "in_progress" {
		t.Fatalf("unexpected service args: rollout=%s org=%s status=%s", mockSvc.lastRollID, mockSvc.lastOrgID, mockSvc.lastNext)
	}
}

func TestTransitionRollout_Success(t *testing.T) {
	rolloutID := uuid.New()
	orgID := uuid.New()
	mockSvc := &rolloutServiceMock{
		result: &models.FirmwareRollout{
			ID:     rolloutID,
			OrgID:  orgID,
			Status: "in_progress",
		},
	}
	h := NewHandlers(mockSvc, zap.NewNop())

	body, _ := json.Marshal(map[string]string{"status": "in_progress", "reason": "operator approved"})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/rollouts/%s/status", rolloutID), bytes.NewReader(body))
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", orgID.String())

	h.TransitionRollout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if mockSvc.callCount != 1 {
		t.Fatalf("expected service call once, got %d", mockSvc.callCount)
	}
	if mockSvc.lastActor != nil {
		t.Fatalf("expected nil actor without X-User-ID, got %v", *mockSvc.lastActor)
	}
	if mockSvc.lastReason == nil || *mockSvc.lastReason != "operator approved" {
		t.Fatalf("expected reason forwarded")
	}
	var response models.FirmwareRollout
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.ID != rolloutID || response.Status != "in_progress" {
		t.Fatalf("unexpected response payload: %+v", response)
	}
}

func TestGetRollout_InvalidRolloutID(t *testing.T) {
	mockSvc := &rolloutServiceMock{}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/rollouts/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "not-a-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("X-Organization-ID", uuid.NewString())

	h.GetRollout(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if mockSvc.getCalls != 0 {
		t.Fatalf("expected service not called, got %d calls", mockSvc.getCalls)
	}
}

func TestGetRollout_Success(t *testing.T) {
	rolloutID := uuid.New()
	orgID := uuid.New()
	mockSvc := &rolloutServiceMock{
		getResult: &models.FirmwareRollout{ID: rolloutID, OrgID: orgID, Status: "paused"},
	}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rollouts/%s", rolloutID), nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("X-Organization-ID", orgID.String())

	h.GetRollout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if mockSvc.getCalls != 1 {
		t.Fatalf("expected one service call, got %d", mockSvc.getCalls)
	}
	var response models.FirmwareRollout
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.ID != rolloutID || response.Status != "paused" {
		t.Fatalf("unexpected response payload: %+v", response)
	}
}

func TestGetRollout_ServiceErrorMapsTo404(t *testing.T) {
	rolloutID := uuid.New()
	orgID := uuid.New()
	mockSvc := &rolloutServiceMock{
		getErr: fmt.Errorf("rollout not found"),
	}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rollouts/%s", rolloutID), nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("X-Organization-ID", orgID.String())

	h.GetRollout(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
	if mockSvc.getCalls != 1 {
		t.Fatalf("expected one service call, got %d", mockSvc.getCalls)
	}
}

func TestListRolloutEvents_Success(t *testing.T) {
	rolloutID := uuid.New()
	orgID := uuid.New()
	mockSvc := &rolloutServiceMock{
		events: []*models.FirmwareRolloutEvent{
			{
				ID:        uuid.New(),
				RolloutID: rolloutID,
				OrgID:     orgID,
				EventType: "rollout.transition",
				Message:   "Rollout transitioned from scheduled to in_progress",
			},
		},
	}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rollouts/%s/events?limit=50", rolloutID), nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("X-Organization-ID", orgID.String())

	h.ListRolloutEvents(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if mockSvc.eventsCall != 1 {
		t.Fatalf("expected one service call, got %d", mockSvc.eventsCall)
	}
	if mockSvc.lastEventType != "" {
		t.Fatalf("expected empty event_type, got %s", mockSvc.lastEventType)
	}
}

func TestListRolloutEvents_InvalidRolloutID(t *testing.T) {
	mockSvc := &rolloutServiceMock{}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/rollouts/not-a-uuid/events", nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "not-a-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.ListRolloutEvents(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if mockSvc.eventsCall != 0 {
		t.Fatalf("expected no service call, got %d", mockSvc.eventsCall)
	}
}

func TestListRolloutEvents_ForwardsEventTypeFilter(t *testing.T) {
	rolloutID := uuid.New()
	orgID := uuid.New()
	mockSvc := &rolloutServiceMock{}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rollouts/%s/events?event_type=rollout.note", rolloutID), nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("X-Organization-ID", orgID.String())

	h.ListRolloutEvents(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if mockSvc.lastEventType != "rollout.note" {
		t.Fatalf("expected event_type forwarded, got %s", mockSvc.lastEventType)
	}
	if mockSvc.lastLimit != 50 || mockSvc.lastOffset != 0 {
		t.Fatalf("expected default pagination limit=50 offset=0, got limit=%d offset=%d", mockSvc.lastLimit, mockSvc.lastOffset)
	}
}

func TestListRolloutEvents_InvalidEventTypeMapsTo400(t *testing.T) {
	rolloutID := uuid.New()
	orgID := uuid.New()
	mockSvc := &rolloutServiceMock{
		eventsErr: fmt.Errorf("invalid event_type: bad_type"),
	}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rollouts/%s/events?event_type=bad_type", rolloutID), nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("X-Organization-ID", orgID.String())

	h.ListRolloutEvents(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestListRolloutEvents_InvalidLimitRejected(t *testing.T) {
	rolloutID := uuid.New()
	orgID := uuid.New()
	mockSvc := &rolloutServiceMock{}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rollouts/%s/events?limit=0", rolloutID), nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("X-Organization-ID", orgID.String())

	h.ListRolloutEvents(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if mockSvc.eventsCall != 0 {
		t.Fatalf("expected no service call, got %d", mockSvc.eventsCall)
	}
}

func TestListRolloutEvents_InvalidOffsetRejected(t *testing.T) {
	rolloutID := uuid.New()
	orgID := uuid.New()
	mockSvc := &rolloutServiceMock{}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rollouts/%s/events?offset=-1", rolloutID), nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("X-Organization-ID", orgID.String())

	h.ListRolloutEvents(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if mockSvc.eventsCall != 0 {
		t.Fatalf("expected no service call, got %d", mockSvc.eventsCall)
	}
}

func TestListRolloutEvents_NonNumericLimitRejected(t *testing.T) {
	rolloutID := uuid.New()
	orgID := uuid.New()
	mockSvc := &rolloutServiceMock{}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rollouts/%s/events?limit=abc", rolloutID), nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("X-Organization-ID", orgID.String())

	h.ListRolloutEvents(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if mockSvc.eventsCall != 0 {
		t.Fatalf("expected no service call, got %d", mockSvc.eventsCall)
	}
}

func TestListRolloutEvents_NonNumericOffsetRejected(t *testing.T) {
	rolloutID := uuid.New()
	orgID := uuid.New()
	mockSvc := &rolloutServiceMock{}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rollouts/%s/events?offset=abc", rolloutID), nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("X-Organization-ID", orgID.String())

	h.ListRolloutEvents(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if mockSvc.eventsCall != 0 {
		t.Fatalf("expected no service call, got %d", mockSvc.eventsCall)
	}
}

func TestListRolloutEvents_TooLargeLimitRejected(t *testing.T) {
	rolloutID := uuid.New()
	orgID := uuid.New()
	mockSvc := &rolloutServiceMock{}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rollouts/%s/events?limit=501", rolloutID), nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("X-Organization-ID", orgID.String())

	h.ListRolloutEvents(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if mockSvc.eventsCall != 0 {
		t.Fatalf("expected no service call, got %d", mockSvc.eventsCall)
	}
}

func TestListRolloutEvents_MaxLimitAccepted(t *testing.T) {
	rolloutID := uuid.New()
	orgID := uuid.New()
	mockSvc := &rolloutServiceMock{
		events: []*models.FirmwareRolloutEvent{},
	}
	h := NewHandlers(mockSvc, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rollouts/%s/events?limit=500&offset=3", rolloutID), nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("X-Organization-ID", orgID.String())

	h.ListRolloutEvents(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if mockSvc.eventsCall != 1 {
		t.Fatalf("expected one service call, got %d", mockSvc.eventsCall)
	}
	if mockSvc.lastLimit != 500 || mockSvc.lastOffset != 3 {
		t.Fatalf("expected forwarded limit=500 offset=3, got limit=%d offset=%d", mockSvc.lastLimit, mockSvc.lastOffset)
	}
}

func TestTransitionRollout_ForwardsActorUserID(t *testing.T) {
	mockSvc := &rolloutServiceMock{}
	h := NewHandlers(mockSvc, zap.NewNop())
	rolloutID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/rollouts/%s/status", rolloutID), strings.NewReader(`{"status":"in_progress","reason":"approved"}`))
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", orgID.String())
	req.Header.Set("X-User-ID", userID.String())

	h.TransitionRollout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if mockSvc.lastActor == nil || *mockSvc.lastActor != userID {
		t.Fatalf("expected actor user id %s, got %+v", userID, mockSvc.lastActor)
	}
}

func TestAddRolloutNote_Success(t *testing.T) {
	mockSvc := &rolloutServiceMock{}
	h := NewHandlers(mockSvc, zap.NewNop())
	rolloutID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/rollouts/%s/events", rolloutID), strings.NewReader(`{"message":"Operator note","metadata":{"source":"ui"}}`))
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", orgID.String())
	req.Header.Set("X-User-ID", userID.String())

	h.AddRolloutNote(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	if mockSvc.noteCalls != 1 {
		t.Fatalf("expected one note call, got %d", mockSvc.noteCalls)
	}
	if mockSvc.lastActor == nil || *mockSvc.lastActor != userID {
		t.Fatalf("expected actor user id %s, got %+v", userID, mockSvc.lastActor)
	}
	if mockSvc.lastNote != "Operator note" {
		t.Fatalf("expected note message forwarded")
	}
}

func TestAddRolloutNote_Validation(t *testing.T) {
	mockSvc := &rolloutServiceMock{}
	h := NewHandlers(mockSvc, zap.NewNop())
	rolloutID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/rollouts/%s/events", rolloutID), strings.NewReader(`{"message":"   "}`))
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", rolloutID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.Header.Set("Content-Type", "application/json")

	h.AddRolloutNote(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if mockSvc.noteCalls != 0 {
		t.Fatalf("expected no service call, got %d", mockSvc.noteCalls)
	}
}
