package service

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/platform/iam/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockRepository is a mock implementation of IAMRepository
type MockRepository struct {
	mock.Mock
}

type MockTenantService struct {
	mock.Mock
}

func (m *MockTenantService) Onboard(ctx context.Context, name, taxID string) (*models.Tenant, error) {
	args := m.Called(ctx, name, taxID)
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantService) GetTenant(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Tenant), args.Error(1)
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

func (m *MockRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockRepository) GetRolesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.Role, error) {
	args := m.Called(ctx, tenantID)
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
	tenantSvc := new(MockTenantService)
	secret := "test-secret"
	svc := NewIAMService(repo, tenantSvc, secret)

	userID := uuid.New()
	tenantID := uuid.New()

	// Test Login
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := &models.User{
		ID:             userID,
		TenantID:       tenantID,
		Email:          "test@example.com",
		PasswordHash:   string(hashedPassword),
		Status:         "active",
	}
	repo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	// ValidateToken also checks user status, so stub GetUserByID
	repo.On("GetUserByID", mock.Anything, userID).Return(user, nil)

	token, err := svc.Login(context.Background(), "test@example.com", "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test ValidateToken
	claims, err := svc.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, tenantID, claims.TenantID)
}

func TestIAMService_CheckPermission(t *testing.T) {
	repo := new(MockRepository)
	tenantSvc := new(MockTenantService)
	svc := NewIAMService(repo, tenantSvc, "secret")

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

func TestIAMService_ValidateToken_AlgorithmCheck(t *testing.T) {
	repo := new(MockRepository)
	tenantSvc := new(MockTenantService)
	iamSvc := NewIAMService(repo, tenantSvc, "test-secret")

	// Test 1: Reject missing algorithm
	t.Run("Reject missing algorithm", func(t *testing.T) {
		noAlgToken := `eyJ1c2VyX2lkIjoiMDAwMDAwMDAwMDAwMDAwIn0.eyJleHAiOjE3MDAwMDAwMDB9.`
		_, err := iamSvc.ValidateToken(noAlgToken)
		assert.Error(t, err)
		// Error message can be "missing algorithm" or generic "invalid token"
	})

	// Test 2: Reject non-HS256 algorithm
	t.Run("Reject non-HS256 algorithm", func(t *testing.T) {
		// Token with alg: RS256 (simulated)
		malformedToken := `eyJhbGciOiJSUzI1NiJ9.eyJ1c2VyX2lkIjoiMDAwMDAwMDAwMDAwMDAwIiwiZXhwIjoxNjc4OTk5OTk5fQ.`
		_, err := iamSvc.ValidateToken(malformedToken)
		assert.Error(t, err)
	})
}

func TestIAMService_ValidateToken_ExpirationCheck(t *testing.T) {
	validUserID := uuid.New()
	validTenantID := uuid.New()
	activeUser := &models.User{
		ID:             validUserID,
		TenantID:       validTenantID,
		Status:         "active",
	}

	// Create an expired token
	expiredClaims := &IAMClaims{
		UserID: validUserID,
		TenantID:  validTenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenStr, _ := expiredToken.SignedString([]byte("test-secret"))

	t.Run("Reject expired token", func(t *testing.T) {
		repo := new(MockRepository)
		tenantSvc := new(MockTenantService)
		iamSvc := NewIAMService(repo, tenantSvc, "test-secret")

		// User exists but token is expired
		repo.On("GetUserByID", mock.Anything, validUserID).Return(activeUser, nil)

		_, err := iamSvc.ValidateToken(expiredTokenStr)
		assert.Error(t, err)
		// jwt.ParseWithClaims returns generic "invalid token" for expiration
		// Our code also checks expiration explicitly, but the library catches it first
	})
}

func TestIAMService_ValidateToken_UserStatusCheck(t *testing.T) {
	validUserID := uuid.New()
	validTenantID := uuid.New()

	// Create a valid, non-expired token
	validClaims := &IAMClaims{
		UserID: validUserID,
		TenantID:  validTenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims)
	validTokenStr, _ := validToken.SignedString([]byte("test-secret"))

	t.Run("Reject inactive user", func(t *testing.T) {
		repo := new(MockRepository)
		tenantSvc := new(MockTenantService)
		iamSvc := NewIAMService(repo, tenantSvc, "test-secret")

		inactiveUser := &models.User{
			ID:             validUserID,
			TenantID:       validTenantID,
			Status:         "inactive",
		}
		repo.On("GetUserByID", mock.Anything, validUserID).Return(inactiveUser, nil)

		_, err := iamSvc.ValidateToken(validTokenStr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not active")
	})

	t.Run("Reject non-existent user", func(t *testing.T) {
		repo := new(MockRepository)
		tenantSvc := new(MockTenantService)
		iamSvc := NewIAMService(repo, tenantSvc, "test-secret")

		repo.On("GetUserByID", mock.Anything, validUserID).Return((*models.User)(nil), assert.AnError)

		_, err := iamSvc.ValidateToken(validTokenStr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("Accept active user", func(t *testing.T) {
		repo := new(MockRepository)
		tenantSvc := new(MockTenantService)
		iamSvc := NewIAMService(repo, tenantSvc, "test-secret")

		activeUser := &models.User{
			ID:             validUserID,
			TenantID:       validTenantID,
			Status:         "active",
		}
		repo.On("GetUserByID", mock.Anything, validUserID).Return(activeUser, nil)

		claims, err := iamSvc.ValidateToken(validTokenStr)
		assert.NoError(t, err)
		assert.Equal(t, validUserID, claims.UserID)
		assert.Equal(t, validTenantID, claims.TenantID)
	})
}
