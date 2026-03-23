package service

import (
	"context"
	"errors"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/platform/iam/models"
	"github.com/flyfleet/fleet-core-service/internal/platform/iam/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type IAMService interface {
	Register(ctx context.Context, email, password, fullName string, orgID uuid.UUID) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(tokenString string) (*IAMClaims, error)
	CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}

type IAMClaims struct {
	UserID uuid.UUID `json:"user_id"`
	OrgID  uuid.UUID `json:"org_id"`
	jwt.RegisteredClaims
}

type iamService struct {
	repo      repository.IAMRepository
	orgSvc    OrganizationService
	jwtSecret []byte
}

func NewIAMService(repo repository.IAMRepository, orgSvc OrganizationService, jwtSecret string) IAMService {
	return &iamService{
		repo:      repo,
		orgSvc:    orgSvc,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *iamService) Register(ctx context.Context, email, password, fullName string, orgID uuid.UUID) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// If orgID is empty, create a new organization for the user
	if orgID == uuid.Nil {
		org, err := s.orgSvc.Onboard(ctx, fullName+"'s Fleet", "")
		if err != nil {
			return nil, err
		}
		orgID = org.ID
	}

	user := &models.User{
		ID:             uuid.New(),
		OrganizationID: orgID,
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
		OrgID:  user.OrganizationID,
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
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*IAMClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
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
