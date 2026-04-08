package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core/internal/platform/iam/models"
	"github.com/flyfleet/fleet-core/internal/platform/iam/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockIAMService is a mock implementation of IAMService
type MockIAMService struct {
	mock.Mock
}

func (m *MockIAMService) Register(ctx context.Context, email, password, fullName string, orgID uuid.UUID) (*models.User, error) {
	return nil, nil // Not needed for middleware tests
}

func (m *MockIAMService) Login(ctx context.Context, email, password string) (string, error) {
	return "", nil // Not needed for middleware tests
}

func (m *MockIAMService) ValidateToken(token string) (*service.IAMClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.IAMClaims), args.Error(1)
}

func (m *MockIAMService) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Bool(0), args.Error(1)
}

func TestAuthMiddleware(t *testing.T) {
	mockSvc := new(MockIAMService)
	handler := AuthMiddleware(mockSvc)

	// Case: Valid Token - verify UUIDs stored as strings
	userID := uuid.New()
	orgID := uuid.New()
	mockSvc.On("ValidateToken", "valid-token").Return(&service.IAMClaims{UserID: userID, OrgID: orgID}, nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SECURITY FIX: Verify UserID is stored as string (not uuid.UUID)
		userIDVal, ok := r.Context().Value(UserIDKey).(string)
		assert.True(t, ok, "UserID should be string")
		assert.Equal(t, userID.String(), userIDVal)

		// Verify OrgID is stored as string (not uuid.UUID)
		orgIDVal, ok := r.Context().Value(OrgIDKey).(string)
		assert.True(t, ok, "OrgID should be string")
		assert.Equal(t, orgID.String(), orgIDVal)

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	handler(next).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case: Missing Header
	req = httptest.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	handler(next).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRBACMiddleware(t *testing.T) {
	mockSvc := new(MockIAMService)
	handler := RBACMiddleware(mockSvc, "vehicle", "read")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	userID := uuid.New()

	// Case: Allowed - Context stores UserID as string, RBACMiddleware parses it to uuid.UUID
	mockSvc.On("CheckPermission", mock.Anything, userID, "vehicle", "read").Return(true, nil).Once()

	req := httptest.NewRequest("GET", "/", nil)
	// SECURITY FIX: Store UserID as string in context (as AuthMiddleware does)
	ctx := context.WithValue(req.Context(), UserIDKey, userID.String())
	rr := httptest.NewRecorder()

	handler(next).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case: Forbidden
	mockSvc.On("CheckPermission", mock.Anything, userID, "vehicle", "read").Return(false, nil).Once()
	rr = httptest.NewRecorder()
	handler(next).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusForbidden, rr.Code)
}
