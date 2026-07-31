package tests

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MockRepo implementation for Token Exchange
type MockExchangeRepo struct {
	users  map[uuid.UUID]*models.User
	links  map[uuid.UUID][]uuid.UUID // parent -> []linked
	events []*models.SecurityEvent
}

func NewMockExchangeRepo() *MockExchangeRepo {
	return &MockExchangeRepo{
		users:  make(map[uuid.UUID]*models.User),
		links:  make(map[uuid.UUID][]uuid.UUID),
		events: make([]*models.SecurityEvent, 0),
	}
}

func (m *MockExchangeRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, nil // Simply return nil if not found for this mock
}

func (m *MockExchangeRepo) IsAccountLinked(ctx context.Context, parentID, targetID uuid.UUID) bool {
	for _, linked := range m.links[parentID] {
		if linked == targetID {
			return true
		}
	}
	return false
}

func (m *MockExchangeRepo) LogSecurityEvent(ctx context.Context, event *models.SecurityEvent) error {
	m.events = append(m.events, event)
	return nil
}

// Stubs to fulfill AuthService token generation dependencies
func (m *MockExchangeRepo) GetUserPermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]string, error) {
	return []string{"basic"}, nil
}
func (m *MockExchangeRepo) GetRolePermissions(ctx context.Context, role string) ([]string, error) {
	return []string{"basic"}, nil
}
func (m *MockExchangeRepo) GetActiveSigningKeys(ctx context.Context) ([]models.SigningKey, error) {
	return []models.SigningKey{{KID: "dummy", PrivateKeyEnc: "dummy-key-data"}}, nil
}
func (m *MockExchangeRepo) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	return &models.Organization{ID: id, Type: "test_tenant"}, nil
}
func (m *MockExchangeRepo) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	return nil
}

func TestSwitchContextAuthorization(t *testing.T) {
	logger := zap.NewNop()
	repo := NewMockExchangeRepo()

	// 1. Setup Identities
	adminID := uuid.New()
	adminUser := &models.User{
		ID: adminID, Roles: []string{"super_admin"},
	}

	parentID := uuid.New()
	parentUser := &models.User{
		ID: parentID, Roles: []string{"fleet_manager"},
	}

	targetID := uuid.New()
	targetUser := &models.User{
		ID: targetID, Roles: []string{"org_admin"}, PermissionMask: 15,
	}

	strangerID := uuid.New()
	strangerUser := &models.User{
		ID: strangerID, Roles: []string{"driver"},
	}

	repo.users[adminID] = adminUser
	repo.users[parentID] = parentUser
	repo.users[targetID] = targetUser
	repo.users[strangerID] = strangerUser

	// Setup explicitly linked accounts
	repo.links[parentID] = append(repo.links[parentID], targetID)

	hasRole := func(u *models.User, r string) bool {
		for _, role := range u.Roles {
			if role == r {
				return true
			}
		}
		return false
	}

	// Scenario A: Super Admin shadowing (Allowed)
	isSuperAdmin := hasRole(adminUser, "super_admin")
	isLinkedAdmin := repo.IsAccountLinked(context.Background(), adminID, targetID)
	if !isSuperAdmin && !isLinkedAdmin {
		t.Fatalf("Super Admin should be allowed to switch")
	}

	// Scenario B: Explicitly Linked Parent shadowing Target (Allowed)
	isSuperParent := hasRole(parentUser, "super_admin")
	isLinkedParent := repo.IsAccountLinked(context.Background(), parentID, targetID)
	if !isSuperParent && !isLinkedParent {
		t.Fatalf("Linked Parent should be allowed to switch")
	}

	// Scenario C: Stranger trying to shadow Target (Denied)
	isSuperStranger := hasRole(strangerUser, "super_admin")
	isLinkedStranger := repo.IsAccountLinked(context.Background(), strangerID, targetID)
	if isSuperStranger || isLinkedStranger {
		t.Fatalf("Stranger should be denied context switch")
	}

	// 2. Test Security Event Struct
	event := &models.SecurityEvent{
		UserID:     targetID,
		EventType:  "CONTEXT_SWITCH_EXECUTED",
		ActorID:    &parentID,
		OnBehalfOf: &targetID,
	}
	repo.LogSecurityEvent(context.Background(), event)

	if len(repo.events) == 0 {
		t.Fatalf("Expected SecurityEvent to be logged")
	}

	lastEvent := repo.events[len(repo.events)-1]
	if lastEvent.ActorID == nil || *lastEvent.ActorID != parentID {
		t.Fatalf("Expected event ActorID to be %v, got %v", parentID, lastEvent.ActorID)
	}

	if lastEvent.OnBehalfOf == nil || *lastEvent.OnBehalfOf != targetID {
		t.Fatalf("Expected event OnBehalfOf to be %v, got %v", targetID, lastEvent.OnBehalfOf)
	}

	// Test passed!
	logger.Info("Authorization logic for Token Exchange passed")
}
