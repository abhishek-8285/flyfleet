package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	redis2 "github.com/flyfleet/fleet-core/internal/domains/iam/redis"
	"github.com/flyfleet/fleet-core/internal/domains/iam/repository"
	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrTOTPSetupFailed   = errors.New("totp_setup_failed")
	ErrTOTPInvalid       = errors.New("totp_invalid")
	ErrBackupCodeUsed    = errors.New("backup_code_already_used")
	ErrBackupCodeInvalid = errors.New("backup_code_invalid")
	ErrMFANotEnabled     = errors.New("mfa_not_enabled")
	ErrSMSNotConfigured  = errors.New("sms_not_configured")
)

const (
	totpIssuer      = "FlyFleet"
	totpDigits      = 6
	totpSecretSize  = 20
	backupCodeCount = 10
	backupCodeLen   = 8
)

type MFAService struct {
	userRepo *repository.UserRepository
	redis    *redis2.Client
	log      *zap.Logger
}

func NewMFAService(userRepo *repository.UserRepository, redis *redis2.Client, log *zap.Logger) *MFAService {
	return &MFAService{
		userRepo: userRepo,
		redis:    redis,
		log:      log,
	}
}

// SetupTOTP generates a TOTP secret and QR code URL for a user.
func (s *MFAService) SetupTOTP(ctx context.Context, userID uuid.UUID) (string, string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", ErrUserNotFound
	}

	// Generate a new TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: user.Email,
		SecretSize:  totpSecretSize,
		Digits:      otp.Digits(totpDigits),
	})
	if err != nil {
		s.log.Error("failed to generate totp key", zap.Error(err))
		return "", "", ErrTOTPSetupFailed
	}

	secret := key.Secret()
	qrURL := key.URL()

	return secret, qrURL, nil
}

// VerifyTOTP verifies a TOTP code against the user's stored secret and enables MFA.
func (s *MFAService) VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	if user.MFASecret == "" {
		return ErrTOTPSetupFailed
	}

	// Validate the TOTP code
	valid := totp.Validate(code, user.MFASecret)
	if !valid {
		// Try backup codes
		return s.verifyBackupCode(ctx, user, code)
	}

	// Enable MFA
	if err := s.userRepo.UpdateMFASecret(ctx, userID, user.MFASecret, true); err != nil {
		return err
	}

	s.log.Info("mfa_totp_enabled", zap.String("user_id", userID.String()))
	return nil
}

// VerifyTOTPCode checks a TOTP code against a given secret without modifying user state.
// Used by auth_service during login.
func VerifyTOTPCode(secret, code string) bool {
	if secret == "" || code == "" {
		return false
	}
	return totp.Validate(code, secret)
}

// GenerateBackupCodes creates a new set of one-time backup codes for the user.
// Codes are hashed with bcrypt before storage — plaintext codes are returned to caller only once.
func (s *MFAService) GenerateBackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	plaintextCodes := make([]string, backupCodeCount)
	hashedCodes := make([]string, backupCodeCount)
	for i := 0; i < backupCodeCount; i++ {
		code, err := randomAlphanumeric(backupCodeLen)
		if err != nil {
			return nil, err
		}
		plaintextCodes[i] = strings.ToUpper(code)
		// Hash each code with bcrypt (low cost since we only check 10 codes at a time)
		hash, err := bcrypt.GenerateFromPassword([]byte(strings.ToUpper(code)), bcrypt.MinCost)
		if err != nil {
			return nil, err
		}
		hashedCodes[i] = string(hash)
	}

	if err := s.userRepo.UpdateMFABackupCodes(ctx, userID, hashedCodes); err != nil {
		return nil, err
	}

	s.log.Info("mfa_backup_codes_generated", zap.String("user_id", userID.String()))
	return plaintextCodes, nil
}

// DisableMFA disables all MFA for a user after verifying their password.
func (s *MFAService) DisableMFA(ctx context.Context, userID uuid.UUID, password string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}

	// Disable MFA
	if err := s.userRepo.UpdateMFASecret(ctx, userID, "", false); err != nil {
		return err
	}

	// Clear backup codes
	if err := s.userRepo.UpdateMFABackupCodes(ctx, userID, []string{}); err != nil {
		s.log.Warn("failed to clear backup codes on MFA disable", zap.Error(err))
	}

	s.log.Info("mfa_disabled", zap.String("user_id", userID.String()))
	return nil
}

// SetupSMS generates and stores an SMS verification code for a user.
// In production, this would trigger an SMS send via Twilio or similar provider.
func (s *MFAService) SetupSMS(ctx context.Context, userID uuid.UUID, phone string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Generate a 6-digit verification code
	code, err := randomNumeric(6)
	if err != nil {
		return err
	}

	// Store the code in Redis with a 5-minute TTL
	smsKey := "sms_verify:" + userID.String()
	if s.redis != nil {
		s.redis.RDB().Set(ctx, smsKey, code, 5*time.Minute)
	}

	// Update user's phone number
	if err := s.userRepo.UpdateProfile(ctx, userID, "", phone, "", "", ""); err != nil {
		return err
	}

	s.log.Info("mfa_sms_setup",
		zap.String("user_id", userID.String()),
		zap.String("phone", phone),
	)
	return nil
}

// VerifySMS enables SMS MFA after the user verifies an SMS code.
// The code must match the server-side stored value (not just be non-empty).
func (s *MFAService) VerifySMS(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	if user.Phone == "" {
		return ErrSMSNotConfigured
	}
	if code == "" {
		return ErrTOTPInvalid
	}

	// Verify the code against the server-side stored value in Redis
	smsKey := "sms_verify:" + userID.String()
	if s.redis != nil {
		storedCode, err := s.redis.RDB().Get(ctx, smsKey).Result()
		if err != nil || storedCode == "" {
			s.log.Warn("sms_code_not_found", zap.String("user_id", userID.String()))
			return ErrTOTPInvalid
		}
		if !strings.EqualFold(code, storedCode) {
			return ErrTOTPInvalid
		}
		// Consume the code (delete from Redis)
		s.redis.RDB().Del(ctx, smsKey)
	}

	// Enable MFA
	if user.MFASecret == "" {
		secret, err := randomAlphanumeric(20)
		if err != nil {
			return err
		}
		user.MFASecret = secret
	}

	if err := s.userRepo.UpdateMFASecret(ctx, userID, user.MFASecret, true); err != nil {
		return err
	}

	s.log.Info("mfa_sms_enabled", zap.String("user_id", userID.String()))
	return nil
}

// ── Internal helpers ──

func (s *MFAService) verifyBackupCode(ctx context.Context, user *models.User, code string) error {
	if len(user.MFABackupCodes) == 0 {
		return ErrTOTPInvalid
	}

	code = strings.ToUpper(strings.TrimSpace(code))

	// Find the matching hashed code
	for i, storedHash := range user.MFABackupCodes {
		// Compare plaintext input against stored bcrypt hash
		err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(code))
		if err == nil {
			// Match found — remove the used code
			user.MFABackupCodes = append(user.MFABackupCodes[:i], user.MFABackupCodes[i+1:]...)
			if err := s.userRepo.UpdateMFABackupCodes(ctx, user.ID, user.MFABackupCodes); err != nil {
				s.log.Error("failed to remove used backup code", zap.Error(err))
			}

			s.log.Info("mfa_backup_code_used",
				zap.String("user_id", user.ID.String()),
				zap.Int("remaining", len(user.MFABackupCodes)),
			)
			return nil
		}
	}

	return ErrBackupCodeInvalid
}

func randomAlphanumeric(n int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b), nil
}

func randomNumeric(n int) (string, error) {
	const charset = "0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b), nil
}
