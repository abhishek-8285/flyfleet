package service

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/platform/iam/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockRepository is a mock implementation of IAMRepository
type MockRepository struct {
	mock.Mock
}

type MockOrganizationService struct {
	mock.Mock
}

func (m *MockOrganizationService) Onboard(ctx context.Context, name, ownerID string) (*models.Organization, error) {
	args := m.Called(ctx, name, ownerID)
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockOrganizationService) GetOrganization(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) GetRolesByOrganization(ctx context.Context, orgID uuid.UUID) ([]*models.Role, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockRepository) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRepository) GetPermissionsForUser(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Permission), args.Error(1)
}

func TestIAMService_JWT(t *testing.T) {
	repo := new(MockRepository)
	orgSvc := new(MockOrganizationService)
	secret := "test-secret"
	svc := NewIAMService(repo, orgSvc, secret)

	userID := uuid.New()
	orgID := uuid.New()

	// Test Login
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := &models.User{
		ID:             userID,
		OrganizationID: orgID,
		Email:          "test@example.com",
		PasswordHash:   string(hashedPassword),
	}
	repo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)

	token, err := svc.Login(context.Background(), "test@example.com", "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test ValidateToken
	claims, err := svc.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, orgID, claims.OrgID)
}

func TestIAMService_CheckPermission(t *testing.T) {
	repo := new(MockRepository)
	orgSvc := new(MockOrganizationService)
	svc := NewIAMService(repo, orgSvc, "secret")

	userID := uuid.New()
	permissions := []*models.Permission{
		{ID: uuid.New(), Resource: "vehicle", Action: "read"},
	}

	repo.On("GetPermissionsForUser", mock.Anything, userID).Return(permissions, nil)

	// Case: Permission exists
	allowed, err := svc.CheckPermission(context.Background(), userID, "vehicle", "read")
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Case: Permission does not exist
	allowed, err = svc.CheckPermission(context.Background(), userID, "vehicle", "write")
	assert.NoError(t, err)
	assert.False(t, allowed)
}
