package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type rolloutRepoMock struct {
	repository.Repository

	rollouts       map[uuid.UUID]*models.FirmwareRollout
	updateCalls    int
	lastUpdateID   uuid.UUID
	lastUpdateNext string
	events         map[uuid.UUID][]*models.FirmwareRolloutEvent
	lastEventType  string
	getErr         error
	updateErr      error
	eventsErr      error
}

func (m *rolloutRepoMock) GetRollout(_ context.Context, rolloutID uuid.UUID) (*models.FirmwareRollout, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	r, ok := m.rollouts[rolloutID]
	if !ok {
		return nil, fmt.Errorf("rollout not found")
	}
	cp := *r
	return &cp, nil
}

func (m *rolloutRepoMock) UpdateRolloutStatus(_ context.Context, id uuid.UUID, status string) error {
	m.updateCalls++
	m.lastUpdateID = id
	m.lastUpdateNext = status
	if m.updateErr != nil {
		return m.updateErr
	}
	r, ok := m.rollouts[id]
	if !ok {
		return fmt.Errorf("rollout not found")
	}
	r.Status = status
	return nil
}

func (m *rolloutRepoMock) ListRolloutEvents(_ context.Context, rolloutID, _ uuid.UUID, eventType string, _, _ int) ([]*models.FirmwareRolloutEvent, error) {
	m.lastEventType = eventType
	if m.eventsErr != nil {
		return nil, m.eventsErr
	}
	if m.events == nil {
		return []*models.FirmwareRolloutEvent{}, nil
	}
	return m.events[rolloutID], nil
}

func (m *rolloutRepoMock) CreateRolloutEvent(_ context.Context, event *models.FirmwareRolloutEvent) error {
	if m.events == nil {
		m.events = map[uuid.UUID][]*models.FirmwareRolloutEvent{}
	}
	m.events[event.RolloutID] = append(m.events[event.RolloutID], event)
	return nil
}

func TestTransitionRollout_ValidTransition(t *testing.T) {
	orgID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {
				ID:     rolloutID,
				OrgID:  orgID,
				Status: "scheduled",
			},
		},
	}

	svc := &service{repo: repoMock, logger: zap.NewNop()}
	updated, err := svc.TransitionRollout(context.Background(), rolloutID, orgID, "in_progress", nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != "in_progress" {
		t.Fatalf("expected updated status in_progress, got %s", updated.Status)
	}
	if repoMock.updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", repoMock.updateCalls)
	}
	if repoMock.lastUpdateID != rolloutID || repoMock.lastUpdateNext != "in_progress" {
		t.Fatalf("unexpected update args id=%s status=%s", repoMock.lastUpdateID, repoMock.lastUpdateNext)
	}
}

func TestTransitionRollout_InvalidStatusRejected(t *testing.T) {
	orgID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {
				ID:     rolloutID,
				OrgID:  orgID,
				Status: "scheduled",
			},
		},
	}

	svc := &service{repo: repoMock, logger: zap.NewNop()}
	_, err := svc.TransitionRollout(context.Background(), rolloutID, orgID, "not_a_real_state", nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid rollout status") {
		t.Fatalf("unexpected error: %v", err)
	}
	if repoMock.updateCalls != 0 {
		t.Fatalf("expected no update call, got %d", repoMock.updateCalls)
	}
}

func TestTransitionRollout_InvalidStateTransitionRejected(t *testing.T) {
	orgID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {
				ID:     rolloutID,
				OrgID:  orgID,
				Status: "completed",
			},
		},
	}

	svc := &service{repo: repoMock, logger: zap.NewNop()}
	_, err := svc.TransitionRollout(context.Background(), rolloutID, orgID, "in_progress", nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
	if !strings.Contains(err.Error(), "invalid rollout transition") {
		t.Fatalf("unexpected error: %v", err)
	}
	if repoMock.updateCalls != 0 {
		t.Fatalf("expected no update call, got %d", repoMock.updateCalls)
	}
}

func TestTransitionRollout_OrgBoundaryEnforced(t *testing.T) {
	rolloutOrgID := uuid.New()
	requestOrgID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {
				ID:     rolloutID,
				OrgID:  rolloutOrgID,
				Status: "scheduled",
			},
		},
	}

	svc := &service{repo: repoMock, logger: zap.NewNop()}
	_, err := svc.TransitionRollout(context.Background(), rolloutID, requestOrgID, "in_progress", nil, nil)
	if err == nil {
		t.Fatal("expected org mismatch error")
	}
	if !strings.Contains(err.Error(), "does not belong to organization") {
		t.Fatalf("unexpected error: %v", err)
	}
	if repoMock.updateCalls != 0 {
		t.Fatalf("expected no update call, got %d", repoMock.updateCalls)
	}
}

func TestTransitionRollout_UpdateFailurePropagates(t *testing.T) {
	orgID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {
				ID:     rolloutID,
				OrgID:  orgID,
				Status: "scheduled",
			},
		},
		updateErr: fmt.Errorf("db update failed"),
	}

	svc := &service{repo: repoMock, logger: zap.NewNop()}
	_, err := svc.TransitionRollout(context.Background(), rolloutID, orgID, "in_progress", nil, nil)
	if err == nil {
		t.Fatal("expected update failure")
	}
	if !strings.Contains(err.Error(), "db update failed") {
		t.Fatalf("unexpected error: %v", err)
	}
	if repoMock.updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", repoMock.updateCalls)
	}
}

func TestGetRollout_Success(t *testing.T) {
	orgID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {
				ID:     rolloutID,
				OrgID:  orgID,
				Status: "scheduled",
			},
		},
	}

	svc := &service{repo: repoMock, logger: zap.NewNop()}
	rollout, err := svc.GetRollout(context.Background(), rolloutID, orgID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rollout.ID != rolloutID || rollout.OrgID != orgID {
		t.Fatalf("unexpected rollout returned: %+v", rollout)
	}
}

func TestGetRollout_OrgBoundaryEnforced(t *testing.T) {
	rolloutOrgID := uuid.New()
	requestOrgID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {
				ID:     rolloutID,
				OrgID:  rolloutOrgID,
				Status: "scheduled",
			},
		},
	}

	svc := &service{repo: repoMock, logger: zap.NewNop()}
	_, err := svc.GetRollout(context.Background(), rolloutID, requestOrgID)
	if err == nil {
		t.Fatal("expected org mismatch error")
	}
	if !strings.Contains(err.Error(), "does not belong to organization") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListRolloutEvents_Success(t *testing.T) {
	orgID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {ID: rolloutID, OrgID: orgID, Status: "in_progress"},
		},
		events: map[uuid.UUID][]*models.FirmwareRolloutEvent{
			rolloutID: {
				{
					ID:        uuid.New(),
					RolloutID: rolloutID,
					OrgID:     orgID,
					EventType: "rollout.transition",
					Message:   "Rollout transitioned",
				},
			},
		},
	}

	svc := &service{repo: repoMock, logger: zap.NewNop()}
	events, err := svc.ListRolloutEvents(context.Background(), rolloutID, orgID, "", 0, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestListRolloutEvents_OrgBoundaryEnforced(t *testing.T) {
	rolloutOrgID := uuid.New()
	requestOrgID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {ID: rolloutID, OrgID: rolloutOrgID, Status: "scheduled"},
		},
	}

	svc := &service{repo: repoMock, logger: zap.NewNop()}
	_, err := svc.ListRolloutEvents(context.Background(), rolloutID, requestOrgID, "", 10, 0)
	if err == nil {
		t.Fatal("expected org mismatch error")
	}
	if !strings.Contains(err.Error(), "does not belong to organization") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListRolloutEvents_NormalizesFilterAlias(t *testing.T) {
	orgID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {ID: rolloutID, OrgID: orgID, Status: "in_progress"},
		},
	}
	svc := &service{repo: repoMock, logger: zap.NewNop()}
	_, err := svc.ListRolloutEvents(context.Background(), rolloutID, orgID, "note", 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repoMock.lastEventType != "rollout.note" {
		t.Fatalf("expected normalized event_type rollout.note, got %s", repoMock.lastEventType)
	}
}

func TestListRolloutEvents_AcceptsCanonicalFilter(t *testing.T) {
	orgID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {ID: rolloutID, OrgID: orgID, Status: "in_progress"},
		},
	}
	svc := &service{repo: repoMock, logger: zap.NewNop()}
	_, err := svc.ListRolloutEvents(context.Background(), rolloutID, orgID, "rollout.transition", 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repoMock.lastEventType != "rollout.transition" {
		t.Fatalf("expected canonical event_type rollout.transition, got %s", repoMock.lastEventType)
	}
}

func TestListRolloutEvents_InvalidFilterRejected(t *testing.T) {
	orgID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {ID: rolloutID, OrgID: orgID, Status: "in_progress"},
		},
	}
	svc := &service{repo: repoMock, logger: zap.NewNop()}
	_, err := svc.ListRolloutEvents(context.Background(), rolloutID, orgID, "bad_type", 10, 0)
	if err == nil {
		t.Fatal("expected invalid event_type error")
	}
	if !strings.Contains(err.Error(), "invalid event_type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransitionRollout_CapturesActorAndReason(t *testing.T) {
	orgID := uuid.New()
	actorID := uuid.New()
	rolloutID := uuid.New()
	reason := "Canary metrics stable"
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {
				ID:     rolloutID,
				OrgID:  orgID,
				Status: "scheduled",
			},
		},
	}

	svc := &service{repo: repoMock, logger: zap.NewNop()}
	_, err := svc.TransitionRollout(context.Background(), rolloutID, orgID, "in_progress", &actorID, &reason)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	events := repoMock.events[rolloutID]
	if len(events) == 0 {
		t.Fatal("expected rollout events to be captured")
	}
	last := events[len(events)-1]
	if last.ActorUserID == nil || *last.ActorUserID != actorID {
		t.Fatalf("expected actor %s, got %+v", actorID, last.ActorUserID)
	}
	if got, _ := last.Metadata["reason"].(string); got != reason {
		t.Fatalf("expected reason %q, got %q", reason, got)
	}
}

func TestAddRolloutNote_Success(t *testing.T) {
	orgID := uuid.New()
	actorID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {ID: rolloutID, OrgID: orgID, Status: "in_progress"},
		},
	}
	svc := &service{repo: repoMock, logger: zap.NewNop()}

	event, err := svc.AddRolloutNote(
		context.Background(),
		rolloutID,
		orgID,
		&actorID,
		"Observed stable telemetry",
		map[string]interface{}{"source": "ui"},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if event.EventType != "rollout.note" {
		t.Fatalf("expected rollout.note, got %s", event.EventType)
	}
	if event.ActorUserID == nil || *event.ActorUserID != actorID {
		t.Fatalf("expected actor %s, got %+v", actorID, event.ActorUserID)
	}
	if event.Message != "Observed stable telemetry" {
		t.Fatalf("unexpected message: %s", event.Message)
	}
}

func TestAddRolloutNote_Validation(t *testing.T) {
	orgID := uuid.New()
	rolloutID := uuid.New()
	repoMock := &rolloutRepoMock{
		rollouts: map[uuid.UUID]*models.FirmwareRollout{
			rolloutID: {ID: rolloutID, OrgID: orgID, Status: "in_progress"},
		},
	}
	svc := &service{repo: repoMock, logger: zap.NewNop()}

	_, err := svc.AddRolloutNote(context.Background(), rolloutID, orgID, nil, "   ", nil)
	if err == nil {
		t.Fatal("expected validation error for empty message")
	}
	if !strings.Contains(err.Error(), "message is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
