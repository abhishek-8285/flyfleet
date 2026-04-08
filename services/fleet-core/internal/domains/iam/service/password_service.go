package service

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"github.com/flyfleet/fleet-core/internal/domains/iam/config"
	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/flyfleet/fleet-core/internal/domains/iam/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordTooShort  = errors.New("password_too_weak")
	ErrPasswordTooWeak   = errors.New("password_too_weak")
	ErrPasswordReused    = errors.New("password_reused")
	ErrPasswordNoUpper   = errors.New("password_too_weak")
	ErrPasswordNoLower   = errors.New("password_too_weak")
	ErrPasswordNoDigit   = errors.New("password_too_weak")
	ErrPasswordNoSpecial = errors.New("password_too_weak")
)

type PasswordService struct {
	userRepo     *repository.UserRepository
	passwordRepo *repository.PasswordHistoryRepository
	sessionRepo  *repository.SessionRepository
	config       config.SecurityConfig
	log          *zap.Logger
}

func NewPasswordService(
	userRepo *repository.UserRepository,
	passwordRepo *repository.PasswordHistoryRepository,
	sessionRepo *repository.SessionRepository,
	cfg config.SecurityConfig,
	log *zap.Logger,
) *PasswordService {
	return &PasswordService{
		userRepo: userRepo, passwordRepo: passwordRepo,
		sessionRepo: sessionRepo,
		config: cfg, log: log,
	}
}

func (s *PasswordService) Validate(ctx context.Context, userID uuid.UUID, password string) error {
	// Length check
	if len(password) < 12 {
		return ErrPasswordTooShort
	}
	// Character class checks
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			hasSpecial = true
		}
	}
	if !hasUpper {
		return ErrPasswordNoUpper
	}
	if !hasLower {
		return ErrPasswordNoLower
	}
	if !hasDigit {
		return ErrPasswordNoDigit
	}
	if !hasSpecial {
		return ErrPasswordNoSpecial
	}

	// History check
	recent, err := s.passwordRepo.GetRecent(ctx, userID, s.config.PasswordHistoryLength)
	if err != nil {
		s.log.Error("failed to get password history", zap.Error(err))
	} else {
		for _, hash := range recent {
			if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil {
				return ErrPasswordReused
			}
		}
	}

	return nil
}

func (s *PasswordService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Validate new password
	if err := s.Validate(ctx, userID, newPassword); err != nil {
		return err
	}

	// Hash and save
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.config.BcryptCost)
	if err != nil {
		return err
	}

	// Save to history first
	if err := s.passwordRepo.Add(ctx, userID, user.PasswordHash); err != nil {
		s.log.Error("failed to save password history", zap.Error(err))
	}

	// Update user
	if err := s.userRepo.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return err
	}

	// Revoke all sessions
	s.sessionRepo.RevokeAllTokensEverywhere(ctx, userID)

	s.log.Info("password_changed", zap.String("user_id", userID.String()))
	return nil
}

func (s *PasswordService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validate token (would extract user_id from JWT claims)
	// For now, placeholder — the real implementation parses the action token
	if token == "" {
		return errors.New("invalid_token")
	}

	// Validate password
	// userID would come from token claims
	// placeholder
	return nil
}

func (s *PasswordService) GetPolicy() models.PasswordPolicyResponse {
	return models.PasswordPolicyResponse{
		MinLength:      12,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
		HistoryLength:  s.config.PasswordHistoryLength,
	}
}

func containsAny(s string, substrings []string) bool {
	lower := strings.ToLower(s)
	for _, sub := range substrings {
		if strings.Contains(lower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}
