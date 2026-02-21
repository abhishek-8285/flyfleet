package service

import (
	"context"
	"errors"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/auth-service/internal/models"
	"github.com/flyfleet/auth-service/internal/repository"
	pb "github.com/flyfleet/auth-service/proto/auth/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	logger    logger.Logger
	repo      *repository.AuthRepository
	jwtSecret string
}

func NewAuthService(logger logger.Logger, repo *repository.AuthRepository, jwtSecret string) *AuthService {
	return &AuthService{
		logger:    logger,
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Issue Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"org_id":  user.OrgID.String(),
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken: tokenString,
		User:        *user,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*models.User, error) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	
	now := time.Now()
	creatorID := s.getUserID(ctx)

	orgID := uuid.Nil

	user := &models.User{
		ID:           uuid.New(),
		OrgID:        orgID,
		Email:        req.Email,
		Role:         "user", // Default role
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		CreatedBy:    creatorID,
		UpdatedAt:    now,
		UpdatedBy:    creatorID,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (*models.User, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	userID, _ := uuid.Parse(claims["user_id"].(string))

	return s.repo.GetUserByID(ctx, userID)
}

func (s *AuthService) getUserID(ctx context.Context) uuid.UUID {
	val := ctx.Value("user_id")
	if val == nil {
		return uuid.Nil
	}
	uidStr, ok := val.(string)
	if !ok {
		return uuid.Nil
	}
	uid, _ := uuid.Parse(uidStr)
	return uid
}
