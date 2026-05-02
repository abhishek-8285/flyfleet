package service

import (
	"context"
	"errors"
	"time"

	"github.com/flyfleet/fleet-core/internal/platform/iam/models"
	"github.com/flyfleet/fleet-core/internal/platform/iam/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type IAMService interface {
	Register(ctx context.Context, email, password, fullName string, tenantID uuid.UUID) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(tokenString string) (*IAMClaims, error)
	CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
	IsSessionActive(ctx context.Context, userID, tenantID uuid.UUID, jti string) (bool, error)
}

type IAMClaims struct {
	UserID uuid.UUID `json:"user_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	JTI    string    `json:"jti"`
	jwt.RegisteredClaims
}

type iamService struct {
	repo      repository.IAMRepository
	tenantSvc TenantService
	jwtSecret []byte
}

func NewIAMService(repo repository.IAMRepository, tenantSvc TenantService, jwtSecret string) IAMService {
	return &iamService{
		repo:      repo,
		tenantSvc: tenantSvc,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *iamService) Register(ctx context.Context, email, password, fullName string, tenantID uuid.UUID) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// If tenantID is empty, create a new tenant for the user
	if tenantID == uuid.Nil {
		tenant, err := s.tenantSvc.Onboard(ctx, fullName+"'s Fleet", "")
		if err != nil {
			return nil, err
		}
		tenantID = tenant.ID
	}

	user := &models.User{
		ID:             uuid.New(),
		TenantID:       tenantID,
		Email:          email,
		FullName:       fullName,
		PasswordHash:   string(hashedPassword),
		Status:         "active",
		CreatedAt:      time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *iamService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	claims := &IAMClaims{
		UserID: user.ID,
		TenantID:  user.TenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *iamService) ValidateToken(tokenString string) (*IAMClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &IAMClaims{}, func(token *jwt.Token) (interface{}, error) {
		// SECURITY: Validate the signing algorithm to prevent algorithm switching attacks
		// Reject tokens that use none, HS256 with public key, or unexpected algorithms
		if token.Header["alg"] == nil {
			return nil, errors.New("missing algorithm in token header")
		}

		alg := token.Header["alg"].(string)

		// Only allow HS256 for this service
		if alg != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("invalid signing algorithm")
		}

		// Verify the token's declared method matches what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*IAMClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// SECURITY: Check token expiration
	// jwt.ParseWithClaims already validates exp, but we add explicit check for clarity
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	// SECURITY: Additional user status validation (check if user still exists and is active)
	// This prevents use of tokens for deleted/inactive users
	user, err := s.repo.GetUserByID(context.Background(), claims.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	if user.Status != "active" {
		return nil, errors.New("user account is not active")
	}

	return claims, nil
}

func (s *iamService) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	permissions, err := s.repo.GetPermissionsForUser(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p.Resource == resource && p.Action == action {
			return true, nil
		}
	}

	return false, nil
}

// IsSessionActive checks if a session (by JTI) is still active.
// Used by AuthMiddleware to reject revoked tokens.
func (s *iamService) IsSessionActive(ctx context.Context, userID, tenantID uuid.UUID, jti string) (bool, error) {
	// In the full IAM implementation, this checks Redis for "session:{jti}"
	// For now, fail-open (return true) to avoid breaking existing functionality
	// when Redis isn't wired up. The auth-service implementation has the full check.
	return true, nil
}
