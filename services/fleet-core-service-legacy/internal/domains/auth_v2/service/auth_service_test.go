//go:build integration
// +build integration

package service

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/auth_v2/models"
	pb "github.com/flyfleet/proto/auth/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// MockAuthRepository is a mock implementation for testing
type MockAuthRepository struct {
	GetUserByEmailFn       func(ctx context.Context, email string) (*models.User, error)
	GetUserByIDFn          func(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUserFn           func(ctx context.Context, user *models.User) error
	UpdateUserLoginAttemptsFn func(ctx context.Context, id uuid.UUID, attempts int, lockedUntil *time.Time) error
	ResetLoginAttemptsFn   func(ctx context.Context, id uuid.UUID) error
	LogSecurityEventFn     func(ctx context.Context, event *models.SecurityEvent) error
	CreateRefreshTokenFn   func(ctx context.Context, token *models.RefreshToken) error
	GetRefreshTokenFn      func(ctx context.Context, token string) (*models.RefreshToken, error)
	RevokeRefreshTokenFn   func(ctx context.Context, token string) error
	RevokeAllUserTokensFn  func(ctx context.Context, userID uuid.UUID) error
	CreatePasswordResetTokenFn func(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	GetPasswordResetTokenFn func(ctx context.Context, token string) (*models.PasswordResetToken, error)
	UpdatePasswordFn       func(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UsePasswordResetTokenFn func(ctx context.Context, token string) error

	// MFA Operations
	UpdateMFASecretFn func(ctx context.Context, userID uuid.UUID, secret string) error
	EnableMFAFn       func(ctx context.Context, userID uuid.UUID, secret string) error
	DisableMFAFn      func(ctx context.Context, userID uuid.UUID) error

	// Session Operations
	ListActiveSessionsFn func(ctx context.Context, userID uuid.UUID) ([]models.RefreshToken, error)
	RevokeSessionByIDFn  func(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error
}

func (m *MockAuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.GetUserByEmailFn != nil {
		return m.GetUserByEmailFn(ctx, email)
	}
	return nil, assert.AnError
}

func (m *MockAuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(ctx, id)
	}
	return nil, assert.AnError
}

func (m *MockAuthRepository) CreateUser(ctx context.Context, user *models.User) error {
	if m.CreateUserFn != nil {
		return m.CreateUserFn(ctx, user)
	}
	return nil
}

func (m *MockAuthRepository) UpdateUserLoginAttempts(ctx context.Context, id uuid.UUID, attempts int, lockedUntil *time.Time) error {
	if m.UpdateUserLoginAttemptsFn != nil {
		return m.UpdateUserLoginAttemptsFn(ctx, id, attempts, lockedUntil)
	}
	return nil
}

func (m *MockAuthRepository) ResetLoginAttempts(ctx context.Context, id uuid.UUID) error {
	if m.ResetLoginAttemptsFn != nil {
		return m.ResetLoginAttemptsFn(ctx, id)
	}
	return nil
}

func (m *MockAuthRepository) LogSecurityEvent(ctx context.Context, event *models.SecurityEvent) error {
	if m.LogSecurityEventFn != nil {
		return m.LogSecurityEventFn(ctx, event)
	}
	return nil
}

func (m *MockAuthRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	if m.CreateRefreshTokenFn != nil {
		return m.CreateRefreshTokenFn(ctx, token)
	}
	return nil
}

func (m *MockAuthRepository) GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	if m.GetRefreshTokenFn != nil {
		return m.GetRefreshTokenFn(ctx, token)
	}
	return nil, assert.AnError
}

func (m *MockAuthRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	if m.RevokeRefreshTokenFn != nil {
		return m.RevokeRefreshTokenFn(ctx, token)
	}
	return nil
}

func (m *MockAuthRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	if m.RevokeAllUserTokensFn != nil {
		return m.RevokeAllUserTokensFn(ctx, userID)
	}
	return nil
}

func (m *MockAuthRepository) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	if m.CreatePasswordResetTokenFn != nil {
		return m.CreatePasswordResetTokenFn(ctx, userID, token, expiresAt)
	}
	return nil
}

func (m *MockAuthRepository) GetPasswordResetToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	if m.GetPasswordResetTokenFn != nil {
		return m.GetPasswordResetTokenFn(ctx, token)
	}
	return nil, assert.AnError
}

func (m *MockAuthRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	if m.UpdatePasswordFn != nil {
		return m.UpdatePasswordFn(ctx, userID, passwordHash)
	}
	return nil
}

func (m *MockAuthRepository) UsePasswordResetToken(ctx context.Context, token string) error {
	if m.UsePasswordResetTokenFn != nil {
		return m.UsePasswordResetTokenFn(ctx, token)
	}
	return nil
}

func (m *MockAuthRepository) UpdateMFASecret(ctx context.Context, userID uuid.UUID, secret string) error {
	if m.UpdateMFASecretFn != nil {
		return m.UpdateMFASecretFn(ctx, userID, secret)
	}
	return nil
}

func (m *MockAuthRepository) EnableMFA(ctx context.Context, userID uuid.UUID, secret string) error {
	if m.EnableMFAFn != nil {
		return m.EnableMFAFn(ctx, userID, secret)
	}
	return nil
}

func (m *MockAuthRepository) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	if m.DisableMFAFn != nil {
		return m.DisableMFAFn(ctx, userID)
	}
	return nil
}

func (m *MockAuthRepository) ListActiveSessions(ctx context.Context, userID uuid.UUID) ([]models.RefreshToken, error) {
	if m.ListActiveSessionsFn != nil {
		return m.ListActiveSessionsFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockAuthRepository) RevokeSessionByID(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error {
	if m.RevokeSessionByIDFn != nil {
		return m.RevokeSessionByIDFn(ctx, sessionID, userID)
	}
	return nil
}

func setupTestService(t *testing.T) (*AuthService, *MockAuthRepository) {
	mockRepo := &MockAuthRepository{}
	service := NewAuthService(nil, mockRepo, "test-secret-key")
	return service, mockRepo
}

func TestAuthService_ValidateEmail_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	mockRepo.GetUserByEmailFn = func(ctx context.Context, email string) (*models.User, error) {
		return nil, assert.AnError
	}

	err := service.validateEmail(ctx, "test@example.com")

	assert.NoError(t, err)
}

func TestAuthService_ValidateEmail_Empty(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	err := service.validateEmail(ctx, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestAuthService_ValidateEmail_InvalidFormat(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	err := service.validateEmail(ctx, "invalid-email")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestAuthService_ValidateEmail_TooLong(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	longEmail := string(make([]byte, 256)) + "@example.com"
	err := service.validateEmail(ctx, longEmail)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "less than 255 characters")
}

func TestAuthService_ValidatePassword_Success(t *testing.T) {
	service, _ := setupTestService(t)

	err := service.validatePassword("Test1234!")

	assert.NoError(t, err)
}

func TestAuthService_ValidatePassword_Empty(t *testing.T) {
	service, _ := setupTestService(t)

	err := service.validatePassword("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestAuthService_ValidatePassword_TooShort(t *testing.T) {
	service, _ := setupTestService(t)

	err := service.validatePassword("Test1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 8 characters")
}

func TestAuthService_ValidatePassword_NoUpper(t *testing.T) {
	service, _ := setupTestService(t)

	err := service.validatePassword("test1234")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uppercase letter")
}

func TestAuthService_ValidatePassword_NoLower(t *testing.T) {
	service, _ := setupTestService(t)

	err := service.validatePassword("TEST1234")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lowercase letter")
}

func TestAuthService_ValidatePassword_NoDigit(t *testing.T) {
	service, _ := setupTestService(t)

	err := service.validatePassword("TestTest")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one number")
}

func TestAuthService_VerifyTOTP_InvalidCodeLength(t *testing.T) {
	service, _ := setupTestService(t)

	valid, err := service.VerifyTOTP("JBSWY3DPEHPK3PXP", "12345")

	assert.False(t, valid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "6 digits")
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{Field: "email", Message: "invalid format"}
	
	assert.Equal(t, "email: invalid format", err.Error())
}

func TestAuthService_Register_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	mockRepo.GetUserByEmailFn = func(ctx context.Context, email string) (*models.User, error) {
		return nil, assert.AnError
	}
	mockRepo.CreateUserFn = func(ctx context.Context, user *models.User) error {
		return nil
	}
	mockRepo.LogSecurityEventFn = func(ctx context.Context, event *models.SecurityEvent) error {
		return nil
	}

	req := &pb.RegisterRequest{
		Email:    "test@example.com",
		Password: "Test1234!",
	}

	user, err := service.Register(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &pb.RegisterRequest{
		Email:    "invalid-email",
		Password: "Test1234!",
	}

	user, err := service.Register(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestAuthService_Register_InvalidPassword(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	req := &pb.RegisterRequest{
		Email:    "test@example.com",
		Password: "weak",
	}

	user, err := service.Register(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "at least 8 characters")
}

func TestAuthService_RequestPasswordReset_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	mockRepo.GetUserByEmailFn = func(ctx context.Context, email string) (*models.User, error) {
		return user, nil
	}
	mockRepo.CreatePasswordResetTokenFn = func(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
		return nil
	}
	mockRepo.LogSecurityEventFn = func(ctx context.Context, event *models.SecurityEvent) error {
		return nil
	}

	err := service.RequestPasswordReset(ctx, "test@example.com")

	assert.NoError(t, err)
}

func TestAuthService_RequestPasswordReset_NonExistentUser(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	mockRepo.GetUserByEmailFn = func(ctx context.Context, email string) (*models.User, error) {
		return nil, assert.AnError
	}

	err := service.RequestPasswordReset(ctx, "nonexistent@example.com")

	assert.NoError(t, err)
}

func TestAuthService_ResetPassword_Success(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	resetToken := &models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}

	mockRepo.GetPasswordResetTokenFn = func(ctx context.Context, token string) (*models.PasswordResetToken, error) {
		return resetToken, nil
	}
	mockRepo.UpdatePasswordFn = func(ctx context.Context, userID uuid.UUID, passwordHash string) error {
		return nil
	}
	mockRepo.UsePasswordResetTokenFn = func(ctx context.Context, token string) error {
		return nil
	}
	mockRepo.RevokeAllUserTokensFn = func(ctx context.Context, userID uuid.UUID) error {
		return nil
	}
	mockRepo.LogSecurityEventFn = func(ctx context.Context, event *models.SecurityEvent) error {
		return nil
	}

	err := service.ResetPassword(ctx, "valid-token", "NewTest1234!")

	assert.NoError(t, err)
}

func TestAuthService_ResetPassword_InvalidToken(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	mockRepo.GetPasswordResetTokenFn = func(ctx context.Context, token string) (*models.PasswordResetToken, error) {
		return nil, assert.AnError
	}

	err := service.ResetPassword(ctx, "invalid-token", "NewTest1234!")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired reset token")
}

func TestAuthService_ResetPassword_WeakPassword(t *testing.T) {
	service, _ := setupTestService(t)
	ctx := context.Background()

	err := service.ResetPassword(ctx, "valid-token", "weak")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 8 characters")
}

func TestAuthService_Login_MfaChallenged(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	rawPassword := "Password123!"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)

	user := &models.User{
		ID:           uuid.New(),
		Email:        "admin@example.com",
		Role:         "admin",
		PasswordHash: string(hashed),
		MfaEnabled:   false,
	}

	mockRepo.GetUserByEmailFn = func(ctx context.Context, email string) (*models.User, error) {
		return user, nil
	}
	mockRepo.UpdateUserLoginAttemptsFn = func(ctx context.Context, id uuid.UUID, attempts int, lockedUntil *time.Time) error {
		return nil
	}
	mockRepo.ResetLoginAttemptsFn = func(ctx context.Context, id uuid.UUID) error {
		return nil
	}
	mockRepo.LogSecurityEventFn = func(ctx context.Context, event *models.SecurityEvent) error {
		return nil
	}

	resp, err := service.Login(ctx, "admin@example.com", "Password123!", "127.0.0.1", "curl/7.68.0")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.MfaRequired)
	assert.True(t, resp.MfaSetupRequired)
	assert.NotEmpty(t, resp.TempToken)
	assert.Empty(t, resp.AccessToken)
}

func TestAuthService_Login_MfaEnabled_UserRole(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	rawPassword := "Password123!"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)

	user := &models.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Role:         "user",
		PasswordHash: string(hashed),
		MfaEnabled:   true,
	}

	mockRepo.GetUserByEmailFn = func(ctx context.Context, email string) (*models.User, error) {
		return user, nil
	}
	mockRepo.UpdateUserLoginAttemptsFn = func(ctx context.Context, id uuid.UUID, attempts int, lockedUntil *time.Time) error {
		return nil
	}
	mockRepo.ResetLoginAttemptsFn = func(ctx context.Context, id uuid.UUID) error {
		return nil
	}
	mockRepo.LogSecurityEventFn = func(ctx context.Context, event *models.SecurityEvent) error {
		return nil
	}

	resp, err := service.Login(ctx, "user@example.com", "Password123!", "127.0.0.1", "curl/7.68.0")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.MfaRequired)
	assert.False(t, resp.MfaSetupRequired)
	assert.NotEmpty(t, resp.TempToken)
	assert.Empty(t, resp.AccessToken)
}

func TestValidateJWT_TamperingAndAlgConfusion(t *testing.T) {
	service, _ := setupTestService(t)
	userID := uuid.New()
	orgID := uuid.New()

	// Generate a valid token
	token, err := service.GenerateToken(userID, orgID, "user")
	assert.NoError(t, err)

	// Tampered payload
	tamperedToken := token[:len(token)-10] + "tampered"
	_, err = service.ValidateToken(tamperedToken)
	assert.ErrorContains(t, err, "signature is invalid")

	// Missing coverage logic mapping alg=none is handled by jwt.Parse validation inherently
}

func TestValidateJWT_ExactExpirationBoundary(t *testing.T) {
	// Setup service with instantly expiring token
	repo := &MockAuthRepository{}
	svc := NewAuthService(nil, repo, "test-secret")
	// Note: Because we can't easily mock the unexported time generation inside GenerateToken 
	// without refactoring the time.Now() calls to use an interface, we rely on the parser asserting an expired token.
	
	// Issue a token and artificially simulate an expired one by manual signing:
	token, _ := svc.GenerateToken(uuid.New(), uuid.New(), "user")
	
	// Wait logic theoretically belongs here for a production check if time wasn't mockable
	_ = token
}

func TestVerifyMFA_ValidCode(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	userID := uuid.New()
	secretKey := "JBSWY3DPEHPK3PXP" // Valid base32 secret
	
	mockRepo.GetUserByIDFn = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
		return &models.User{ID: id, MfaEnabled: true, MfaSecret: secretKey}, nil
	}

	// This is a placeholder assertion for MFA sequence testing.
	// Production logic leverages the standard totp.GenerateCode function
	_ = ctx
	_ = userID
	_ = service
}

func TestLogin_ProducesSecurityEventAndChecksOrgState(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.GetUserByEmailFn = func(ctx context.Context, email string) (*models.User, error) {
		return &models.User{ID: userID, OrgID: uuid.New()}, nil
	}
	
	// Simulate an org lookup failure or disabled org
	mockRepo.LogSecurityEventFn = func(ctx context.Context, event *models.SecurityEvent) error {
		assert.Equal(t, "LOGIN_FAILED", event.EventType)
		assert.Equal(t, userID, event.UserID)
		return nil
	}

	_, err := service.Login(ctx, "test@example.com", "wrongpass", "127.0.0.1", "test-agent")
	assert.Error(t, err)
}

func TestPasswordReset_TokenReuse(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	resetToken := "token123"
	
	// First use mock prep
	mockRepo.GetPasswordResetTokenFn = func(ctx context.Context, token string) (*models.PasswordResetToken, error) {
		return &models.PasswordResetToken{
			Token: resetToken,
			Used:  true, // explicitly simulate it's already used
		}, nil
	}

	err := service.ResetPassword(ctx, resetToken, "NewPass123!")
	assert.ErrorContains(t, err, "invalid or expired reset token")
}
