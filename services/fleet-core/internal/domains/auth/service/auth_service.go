package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"database/sql"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core/internal/domains/auth/models"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	pb "github.com/flyfleet/proto/auth/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// AuthRepositoryInterface defines the interface for auth repository
type AuthRepositoryInterface interface {
	// User operations
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUserLoginAttempts(ctx context.Context, id uuid.UUID, attempts int, lockedUntil *time.Time) error
	ResetLoginAttempts(ctx context.Context, id uuid.UUID) error

	// Security events
	LogSecurityEvent(ctx context.Context, event *models.SecurityEvent) error

	// Token operations
	CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	// TryRevokeRefreshToken atomically revokes the token only if not already revoked.
	// Returns the token row if it was successfully revoked, nil if already revoked.
	TryRevokeRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error)
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
	ListActiveSessions(ctx context.Context, userID uuid.UUID) ([]models.RefreshToken, error)
	RevokeSessionByID(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error

	// MFA Operations
	UpdateMFASecret(ctx context.Context, userID uuid.UUID, secret string) error
	EnableMFA(ctx context.Context, userID uuid.UUID, secret string) error
	DisableMFA(ctx context.Context, userID uuid.UUID) error

	// Password reset
	CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	GetPasswordResetToken(ctx context.Context, token string) (*models.PasswordResetToken, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UsePasswordResetToken(ctx context.Context, token string) error

	// RBAC Operations
	CreateRole(ctx context.Context, role *models.Role) error
	AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID, tenantID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]models.Role, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]string, error)
	CheckPermission(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, permission string) (bool, error)
	GetRolePermissions(ctx context.Context, role string) ([]string, error) // Legacy support

	// MFA Recovery Codes
	SaveRecoveryCodes(ctx context.Context, userID uuid.UUID, codeHashes []string) error
	GetUnusedRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]models.MfaRecoveryCode, error)
	MarkRecoveryCodeUsed(ctx context.Context, codeID uuid.UUID) error

	// Risk Assessment
	GetLastLogin(ctx context.Context, userID uuid.UUID) (*models.LoginHistory, error)
	LogLoginAttempt(ctx context.Context, attempt *models.LoginHistory) error

	// SSO & Federation (OIDC)
	GetOidcProviderByName(ctx context.Context, name string) (*models.OidcProvider, error)
	GetUserByOidcLink(ctx context.Context, providerID uuid.UUID, providerUserID string) (*models.User, error)
	LinkUserToOidc(ctx context.Context, link *models.UserOidcLink) error
	UpdateOidcLinkLastLogin(ctx context.Context, linkID uuid.UUID) error

	// JWKS & Signing Keys
	GetActiveSigningKeys(ctx context.Context) ([]models.SigningKey, error)
	GetSigningKey(ctx context.Context, kid string) (*models.SigningKey, error)
	SaveSigningKey(ctx context.Context, key *models.SigningKey) error
	UpdateSigningKeyStatus(ctx context.Context, kid string, status string) error

	// API Keys & M2M
	SaveAPIKey(ctx context.Context, key *models.APIKey) error
	GetAPIKeyByPrefix(ctx context.Context, prefix string) (*models.APIKey, error)
	ListAPIKeys(ctx context.Context, tenantID uuid.UUID) ([]models.APIKey, error)
	RevokeAPIKey(ctx context.Context, id uuid.UUID) error
	UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error

	GetServiceAccount(ctx context.Context, id uuid.UUID) (*models.ServiceAccount, error)
	CreateServiceAccount(ctx context.Context, sa *models.ServiceAccount) error
	ListServiceAccounts(ctx context.Context, tenantID uuid.UUID) ([]models.ServiceAccount, error)
	DeleteServiceAccount(ctx context.Context, id uuid.UUID) error

	// Organization operations
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error)
	ListOrganizations(ctx context.Context, limit, offset int32) ([]models.Organization, int32, error)
	GetOrganizationsByEmail(ctx context.Context, email string) ([]models.Organization, error)
	CreateOrganization(ctx context.Context, org *models.Organization) error

	// Linked Accounts
	LinkAccounts(ctx context.Context, parentID, linkedID uuid.UUID) error
	GetLinkedAccounts(ctx context.Context, parentID uuid.UUID) ([]models.UserLink, error)
	IsAccountLinked(ctx context.Context, parentID, targetID uuid.UUID) bool

	// Invite operations
	GetInviteByCode(ctx context.Context, code string) (*models.Invite, error)
	MarkInviteUsed(ctx context.Context, id uuid.UUID) error
	CreateInvite(ctx context.Context, invite *models.Invite) error

	// Compliance & Admin
	RevokeAllTokensGlobal(ctx context.Context) error
	RevokeAllOrgTokens(ctx context.Context, tenantID uuid.UUID) error
	AnonymizeUser(ctx context.Context, userID uuid.UUID) error
}

const (
	AccessTokenExpiry  = 15 * time.Minute
	RefreshTokenExpiry = 7 * 24 * time.Hour
	MaxLoginAttempts   = 5
	LockoutDuration    = 15 * time.Minute
	MinPasswordLength  = 8
	TOTPIssuer         = "FlyFleet"
	TOTPPeriod         = 30
	TOTPDigits         = 6
)

// Argon2id parameters (RFC 9106)
const (
	Argon2Iterations  = 3
	Argon2Memory      = 64 * 1024 // 64 MB
	Argon2Parallelism = 4
	Argon2SaltLength  = 16
	Argon2KeyLength   = 32
)

// Validation regex patterns
var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type AuthService struct {
	logger    logger.Logger
	repo      AuthRepositoryInterface
	identity  IdentityService
	jwtSecret string
	redis     *redis.Client
	nats      *nats.Conn
}

func NewAuthService(logger logger.Logger, repo AuthRepositoryInterface, identity IdentityService, jwtSecret string, rdb *redis.Client, nc *nats.Conn) *AuthService {
	return &AuthService{
		logger:    logger,
		repo:      repo,
		identity:  identity,
		jwtSecret: jwtSecret,
		redis:     rdb,
		nats:      nc,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password, tenantIDStr, ipAddress, userAgent string) (*models.AuthResponse, error) {
	s.logger.Info("Login attempt started", zap.String("email", email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error("GetUserByEmail failed", zap.Error(err), zap.String("email", email))
		return nil, err
	}
	s.logger.Info("User found", zap.String("email", email), zap.String("userID", user.ID.String()))

	// Override org if provided
	if tenantIDStr != "" {
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, errors.New("invalid tenant_id format")
		}

		orgs, err := s.repo.GetOrganizationsByEmail(ctx, email)
		if err != nil {
			return nil, err
		}

		found := false
		for _, o := range orgs {
			if o.ID == tenantID {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("invalid organization")
		}
		user.TenantID = tenantID
	}
	if err != nil {
		return nil, err
	}

	// 1. Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, fmt.Errorf("account locked until %v", user.LockedUntil)
	}

	// 2. Calculate Risk Score
	lastLogin, _ := s.repo.GetLastLogin(ctx, user.ID)
	country := s.getCountryFromIP(ipAddress)
	riskScore := s.calculateRiskScore(user, lastLogin, ipAddress, userAgent, country)

	// Distributed Rate Limiting (Checking via Redis would happen in Middleware/Handler,
	// but we'll ensure the service layer honors it if possible)

	// Handle high-risk actions
	if riskScore > 90 {
		_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
			UserID:    user.ID,
			EventType: "LOGIN_RISK_LOCKOUT",
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Metadata:  fmt.Sprintf("risk: %.2f", riskScore),
		})
		// Lock account temporarily
		until := time.Now().Add(LockoutDuration * 2)
		_ = s.repo.UpdateUserLoginAttempts(ctx, user.ID, MaxLoginAttempts, &until)
		return nil, errors.New("security check failed: unusual activity detected. account locked for your safety")
	}

	ok, err := s.comparePassword(password, user.PasswordHash)
	if err != nil || !ok {
		s.logger.Error("Password comparison failed", zap.Error(err), zap.Bool("ok", ok), zap.String("email", email))
		// ... existing failure logic ...
		newAttempts := int(user.FailedLoginAttempts.Int64) + 1
		var lockedUntil *time.Time
		if newAttempts >= MaxLoginAttempts {
			until := time.Now().Add(LockoutDuration)
			lockedUntil = &until
		}
		_ = s.repo.UpdateUserLoginAttempts(ctx, user.ID, newAttempts, lockedUntil)

		_ = s.repo.LogLoginAttempt(ctx, &models.LoginHistory{
			UserID:    user.ID,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Country:   country,
			Success:   false,
			RiskScore: riskScore,
		})

		return nil, errors.New("invalid credentials")
	}

	// Reset login attempts
	_ = s.repo.ResetLoginAttempts(ctx, user.ID)

	s.logger.Info("Login success, checking MFA requirements", 
		zap.String("email", user.Email),
		zap.String("role", user.Role),
		zap.Float64("risk_score", riskScore))

	// MFA Requirement
	if user.MfaEnabled || riskScore > 85 {
		s.logger.Info("MFA challenged", zap.String("email", user.Email), zap.Float64("risk_score", riskScore))
		tempToken, err := s.generateTempToken(user)
		if err != nil {
			return nil, err
		}

		_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
			UserID:    user.ID,
			EventType: "LOGIN_MFA_CHALLENGED",
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Metadata:  fmt.Sprintf("risk: %.2f", riskScore),
		})

		return &models.AuthResponse{
			MfaRequired:      true,
			MfaSetupRequired: !user.MfaEnabled,
			TempToken:        tempToken,
			User:             user,
		}, nil
	}

	s.logger.Info("Issuing tokens", zap.String("email", user.Email))
	// 4. Generate actual tokens
	authResponse, err := s.issueTokens(ctx, user, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	// Log successful login
	_ = s.repo.LogLoginAttempt(ctx, &models.LoginHistory{
		UserID:    user.ID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Country:   country,
		Success:   true,
		RiskScore: riskScore,
	})

	return authResponse, nil
}

func (s *AuthService) DevLogin(ctx context.Context, email string, ipAddress, userAgent string) (*models.AuthResponse, error) {
	if email == "" {
		email = "admin@flyfleet.com"
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Skip password and MFA checks in DevLogin
	s.logger.Info("DevLogin successful (Password Bypass)", zap.String("email", email))
	return s.issueTokens(ctx, user, ipAddress, userAgent)
}

func (s *AuthService) issueTokens(ctx context.Context, user *models.User, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// Fetch permissions via IdentityService/Repository (New RBAC)
	permissions, _ := s.repo.GetUserPermissions(ctx, user.ID, user.TenantID)
	if len(permissions) == 0 {
		// Fallback to legacy role-based permissions if no multi-roles assigned
		permissions, _ = s.repo.GetRolePermissions(ctx, user.Role)
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user, permissions)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(ctx, user.ID, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	// Audit session issuance with org context
	_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    user.ID,
		EventType: "SESSION_ISSUED",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Metadata:  fmt.Sprintf("tenant_id: %s, role: %s", user.TenantID.String(), user.Role),
	})

	return &models.AuthResponse{
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		TokenType:      "Bearer",
		ExpiresIn:      int64(AccessTokenExpiry.Seconds()),
		User:           user,
		Permissions:    permissions,
		PermissionMask: user.PermissionMask,
	}, nil
}

func (s *AuthService) generateTempToken(user *models.User) (string, error) {
	// Use consolidated logic later, but for now this works.
	// ... (no changes needed for temp token yet)
	// Actually, let's keep it simple.
	claims := models.TokenClaims{
		UserID: user.ID.String(),
		TenantID:  user.TenantID.String(),
		Role:   user.Role,
		Type:   "mfa_pending",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": claims.UserID,
		"tenant_id":  claims.TenantID,
		"role":    claims.Role,
		"type":    claims.Type,
		"exp":     time.Now().Add(5 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr, tenantIDStr, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// ATOMIC: Try to revoke the refresh token in a single DB operation.
	// This prevents the race condition where concurrent refresh requests all see revoked=false.
	revokedToken, err := s.repo.TryRevokeRefreshToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, err
	}

	if revokedToken == nil {
		// Token was already revoked or expired. Two scenarios:
		// 1. Benign: Concurrent refresh from same browser (same IP/UA) — race condition
		// 2. Malicious: Token theft — attacker replaying stolen token from different IP
		//
		// Look up the revoked token record to check if this is the same session.
		existingToken, lookupErr := s.repo.GetRefreshToken(ctx, refreshTokenStr)
		if lookupErr != nil || existingToken == nil {
			// Token doesn't exist at all — likely expired or invalid
			return nil, errors.New("invalid or expired refresh token")
		}

		// Check if the original token's IP matches the current request IP.
		// Same IP = benign concurrent refresh (race condition from multiple tabs).
		// Different IP = potential token theft — escalate with security alarm.
		originalIP := ""
		if existingToken.IPAddress != nil {
			originalIP = *existingToken.IPAddress
		}

		// SECURITY ENFORCEMENT: If we can't verify the location (both IPs empty), 
		// or if they don't match, we must treat it as a potential theft for safety.
		if originalIP == "" || ipAddress == "" || originalIP != ipAddress {
			// SECURITY ALERT: Token used from different IP or unverified location — likely theft
			_ = s.repo.RevokeAllUserTokens(ctx, existingToken.UserID)
			_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
				UserID:    existingToken.UserID,
				EventType: "REFRESH_TOKEN_REUSE_DETECTED",
				IPAddress: ipAddress,
				UserAgent: userAgent,
				Metadata: fmt.Sprintf("token_id: %s, original_ip: %s, current_ip: %s",
					existingToken.ID, originalIP, ipAddress),
			})
			return nil, errors.New("security alert: token reuse detected from different location. all sessions revoked")
		}

		// Same IP — benign concurrent refresh. Proceed to issue new tokens.
		s.logger.Debug("concurrent refresh detected (same IP), issuing new tokens",
			zap.String("ip_address", ipAddress))

		// Get user for this session
		user, err := s.repo.GetUserByID(ctx, existingToken.UserID)
		if err != nil {
			return nil, err
		}

		// Apply org context switch if requested
		if tenantIDStr != "" {
			tenantID, err := uuid.Parse(tenantIDStr)
			if err != nil {
				return nil, errors.New("invalid organization id format")
			}
			orgs, err := s.repo.GetOrganizationsByEmail(ctx, user.Email)
			if err != nil {
				return nil, err
			}
			found := false
			for _, o := range orgs {
				if o.ID == tenantID {
					found = true
					break
				}
			}
			if !found {
				return nil, errors.New("unauthorized: user does not belong to the requested organization")
			}
			if tenantID != user.TenantID {
				_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
					UserID:    user.ID,
					EventType: "ORG_CONTEXT_SWITCH",
					IPAddress: ipAddress,
					UserAgent: userAgent,
					Metadata:  fmt.Sprintf("old_org: %s, new_org: %s", user.TenantID.String(), tenantID.String()),
				})
				user.TenantID = tenantID
			}
		}

		return s.issueTokens(ctx, user, ipAddress, userAgent)
	}

	// Successfully revoked — this is the first (winner) of any concurrent requests.
	// Get user
	user, err := s.repo.GetUserByID(ctx, revokedToken.UserID)
	if err != nil {
		return nil, err
	}

	// Organization context switching logic
	if tenantIDStr != "" {
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, errors.New("invalid organization id format")
		}

		// Verify user belongs to the requested organization
		orgs, err := s.repo.GetOrganizationsByEmail(ctx, user.Email)
		if err != nil {
			return nil, err
		}

		found := false
		for _, o := range orgs {
			if o.ID == tenantID {
				found = true
				break
			}
		}

		if !found {
			return nil, errors.New("unauthorized: user does not belong to the requested organization")
		}

		// Audit context switch
		if tenantID != user.TenantID {
			_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
				UserID:    user.ID,
				EventType: "ORG_CONTEXT_SWITCH",
				IPAddress: ipAddress,
				UserAgent: userAgent,
				Metadata:  fmt.Sprintf("old_org: %s, new_org: %s", user.TenantID.String(), tenantID.String()),
			})
		}

		// Update user's active organization context for the new session
		user.TenantID = tenantID
	}

	// Issue new tokens (scoped to user.TenantID)
	return s.issueTokens(ctx, user, ipAddress, userAgent)
}

func (s *AuthService) VerifyLoginMFA(ctx context.Context, tempTokenStr, code, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// 1. Validate the temp token
	token, err := jwt.Parse(tempTokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired temp token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "mfa_pending" {
		return nil, errors.New("invalid token type")
	}

	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return nil, errors.New("invalid user id in token")
	}

	// 2. Fetch the user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.MfaEnabled {
		return nil, errors.New("MFA is not enabled for this user. Please complete setup first.")
	}

	// 3. Verify the code (TOTP or Recovery Code)
	mfaVerified := false
	isRecovery := len(code) > 6

	if isRecovery {
		verified, err := s.VerifyRecoveryCode(ctx, user.ID, code)
		if err == nil && verified {
			mfaVerified = true
		}
	} else {
		secret := ""
		if user.MfaSecret.Valid {
			secret = user.MfaSecret.String
		}
		valid, err := s.VerifyTOTP(secret, code)
		if err == nil && valid {
			mfaVerified = true
		}
	}

	if !mfaVerified {
		// Log failed MFA verification AND increment failed attempts
		_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
			UserID:    user.ID,
			EventType: "LOGIN_MFA_FAILED",
			IPAddress: "",
			UserAgent: "",
			Metadata:  fmt.Sprintf("type: %s", map[bool]string{true: "recovery", false: "totp"}[isRecovery]),
		})

		// SECURITY FIX: Increment failed MFA attempts to enable account lockout
		currentAttempts := int(user.FailedLoginAttempts.Int64)
		newAttempts := currentAttempts + 1
		// Update the failed attempts counter
		_ = s.repo.UpdateUserLoginAttempts(ctx, user.ID, newAttempts, nil)

		// Check if account should be locked (e.g., after 5 failed MFA attempts)
		if newAttempts >= 5 {
			lockoutDuration := 15 * time.Minute
			lockedUntil := time.Now().Add(lockoutDuration)
			_ = s.repo.UpdateUserLoginAttempts(ctx, user.ID, newAttempts, &lockedUntil)
			s.logger.Warn("User account locked due to excessive MFA failures",
				zap.String("user_id", user.ID.String()),
				zap.Int("attempts", newAttempts))
		}

		return nil, errors.New("invalid MFA code")
	}

	// 4. Issue tokens
	return s.issueTokens(ctx, user, ipAddress, userAgent)
}

func (s *AuthService) SetupAdminMFA(ctx context.Context, tempTokenStr, secret, code, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// 1. Validate the temp token
	token, err := jwt.Parse(tempTokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired temp token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "mfa_pending" {
		return nil, errors.New("invalid token type")
	}

	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return nil, errors.New("invalid user id in token")
	}

	// 2. Fetch the user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.MfaEnabled {
		return nil, errors.New("MFA is already enabled for this user")
	}

	// 3. Verify the TOTP code against the provided new secret
	valid, err := s.VerifyTOTP(secret, code)
	if err != nil || !valid {
		return nil, errors.New("invalid TOTP code")
	}

	// 4. In a hypothetical SetupAdminMFA scenario, we reuse the EnableMFA function logic
	// But first, we need to bypass token extraction since they aren't fully logged in.
	if err := s.repo.EnableMFA(ctx, user.ID, secret); err != nil {
		return nil, fmt.Errorf("failed to save MFA settings: %w", err)
	}

	// Reload user to get updated state
	user, _ = s.repo.GetUserByID(ctx, user.ID)

	// 5. Issue tokens
	return s.issueTokens(ctx, user, ipAddress, userAgent)
}

func (s *AuthService) generateAccessToken(user *models.User, permissions []string) (string, error) {
	// 1. Get active signing key
	keys, err := s.repo.GetActiveSigningKeys(context.Background())
	if err != nil || len(keys) == 0 {
		return "", errors.New("no active signing keys available")
	}

	// Use the first active key
	activeKey := keys[0]

	// 2. Parse private key
	block, _ := pem.Decode([]byte(activeKey.PrivateKeyEnc))
	if block == nil {
		return "", errors.New("failed to parse private key PEM")
	}

	var privKey *rsa.PrivateKey
	// Try PKCS1 first
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		privKey = key
	} else {
		// Try PKCS8
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse RSA private key (tried PKCS1 and PKCS8): %w", err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return "", errors.New("parsed key is not an RSA private key")
		}
		privKey = rsaKey
	}

	jti := uuid.New().String()

	// Fetch organization to get tenant_type
	org, _ := s.repo.GetOrganizationByID(context.Background(), user.TenantID)
	tenantType := ""
	if org != nil {
		tenantType = org.Type
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":    user.ID.String(),
		"iss":    "FlyFleet",
		"aud":    "fleet-dashboard",
		"tenant_id": user.TenantID.String(),
		"rol":    user.Role,
		"ten":    tenantType, // tenant_type claim
		"type":   "access",
		"prm":    permissions,
		"pmsk":   user.PermissionMask,
		"jti":    jti,
		"exp":    time.Now().Add(AccessTokenExpiry).Unix(),
		"iat":    time.Now().Unix(),
	})
	token.Header["kid"] = activeKey.KID

	return token.SignedString(privKey)
}

func (s *AuthService) ResolveOrgs(ctx context.Context, email string) ([]models.Organization, error) {
	orgs, err := s.repo.GetOrganizationsByEmail(ctx, email)
	if err != nil {
		// Generic success (empty list) on error/not found to prevent enumeration
		s.logger.Debug("resolve orgs failed or user not found", zap.String("email", email), zap.Error(err))
		return []models.Organization{}, nil
	}
	return orgs, nil
}

func (s *AuthService) generateRefreshToken(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) (string, error) {
	// Generate random token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(bytes)

	now := time.Now()
	var ip *string
	var ua *string
	if ipAddress != "" {
		ip = &ipAddress
	}
	if userAgent != "" {
		ua = &userAgent
	}

	refreshToken := &models.RefreshToken{
		ID:         uuid.New(),
		UserID:     userID,
		Token:      token,
		ExpiresAt:  now.Add(RefreshTokenExpiry),
		CreatedAt:  now,
		Revoked:    false,
		IPAddress:  ip,
		UserAgent:  ua,
		LastUsedAt: &now,
	}

	if err := s.repo.CreateRefreshToken(ctx, refreshToken); err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) RevokeToken(ctx context.Context, refreshTokenStr string) error {
	return s.repo.RevokeRefreshToken(ctx, refreshTokenStr)
}

func (s *AuthService) RevokeAllTokens(ctx context.Context, userID uuid.UUID) error {
	return s.repo.RevokeAllUserTokens(ctx, userID)
}

func (s *AuthService) ListUserSessions(ctx context.Context, userID uuid.UUID) ([]models.RefreshToken, error) {
	return s.repo.ListActiveSessions(ctx, userID)
}

func (s *AuthService) RevokeUserSession(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error {
	return s.RevokeDeviceAccess(ctx, sessionID.String(), "user_revocation")
}

func (s *AuthService) RevokeDeviceAccess(ctx context.Context, deviceID string, reason string) error {
	// For session revocation, deviceID here is essentially the jti or session ID
	var jtiStr string
	var tenantIDStr string
	var userIDStr string

	if id, err := uuid.Parse(deviceID); err == nil {
		_ = s.repo.RevokeSessionByID(ctx, id, uuid.Nil)
		jtiStr = id.String() // assuming sessionID == jti
	} else {
		jtiStr = deviceID
	}

	key := fmt.Sprintf("blacklist:jti:%s", jtiStr)
	if s.redis != nil {
		// TTL matches AccessTokenExpiry (15m)
		s.redis.SetEx(ctx, key, "true", AccessTokenExpiry)
	}

	if s.nats != nil {
		event := map[string]string{
			"jti":       jtiStr,
			"user_id":   userIDStr, // would be fetched if needed
			"tenant_id":    tenantIDStr,  // would be fetched if needed
			"device_id": deviceID,
			"reason":    reason,
			"action":    "revoked",
		}
		data, _ := json.Marshal(event)
		s.nats.Publish("auth.events", data)
	}

	key = fmt.Sprintf("auth:revoked_session:%s", deviceID)
	if s.redis != nil {
		// TTL matches AccessTokenExpiry (15m)
		s.redis.SetEx(ctx, key, "revoked", AccessTokenExpiry)
	}

	if s.nats != nil {
		event := map[string]string{
			"device_id": deviceID,
			"reason":    reason,
			"action":    "revoked",
		}
		data, _ := json.Marshal(event)
		s.nats.Publish("auth.device.revoked", data)
	}

	return nil
}

// validateEmail validates email format and uniqueness
func (s *AuthService) validateEmail(ctx context.Context, email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return ValidationError{Field: "email", Message: "email is required"}
	}

	if len(email) > 255 {
		return ValidationError{Field: "email", Message: "email must be less than 255 characters"}
	}

	if !emailRegex.MatchString(email) {
		return ValidationError{Field: "email", Message: "invalid email format"}
	}

	// Check if email already exists
	_, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil {
		return ValidationError{Field: "email", Message: "email already registered"}
	}

	return nil
}

// validatePassword validates password strength
func (s *AuthService) validatePassword(password string) error {
	if password == "" {
		return ValidationError{Field: "password", Message: "password is required"}
	}

	if len(password) < MinPasswordLength {
		return ValidationError{Field: "password", Message: fmt.Sprintf("password must be at least %d characters", MinPasswordLength)}
	}

	if len(password) > 128 {
		return ValidationError{Field: "password", Message: "password must be less than 128 characters"}
	}

	// Check for at least one uppercase letter
	hasUpper := false
	hasLower := false
	hasDigit := false
	for _, c := range password {
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
		}
		if c >= 'a' && c <= 'z' {
			hasLower = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}

	if !hasUpper {
		return ValidationError{Field: "password", Message: "password must contain at least one uppercase letter"}
	}

	if !hasLower {
		return ValidationError{Field: "password", Message: "password must contain at least one lowercase letter"}
	}

	if !hasDigit {
		return ValidationError{Field: "password", Message: "password must contain at least one number"}
	}

	return nil
}

func (s *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*models.User, error) {
	// Validate email
	if err := s.validateEmail(ctx, req.Email); err != nil {
		return nil, err
	}

	// Validate password
	if err := s.validatePassword(req.Password); err != nil {
		return nil, err
	}

	hashedPassword, err := s.generateArgon2Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	creatorID := s.getUserID(ctx)

	var tenantID uuid.UUID
	var role string

	// Handle Organization & Role logic based on Invite vs Creation
	if req.InviteCode != "" {
		// User is joining via an invite
		invite, err := s.repo.GetInviteByCode(ctx, req.InviteCode)
		if err != nil {
			return nil, fmt.Errorf("invalid or expired invite code")
		}

		if invite.Used || now.After(invite.Expiry) {
			return nil, fmt.Errorf("invite code is expired or has already been used")
		}

		tenantID = invite.TenantID
		role = invite.Role

		// Mark invite as used
		_ = s.repo.MarkInviteUsed(ctx, invite.ID)

	} else {
		// Not joining via invite.
		if req.OrgName != "" {
			// Creating a new organization
			tenantID = uuid.New()
			org := &models.Organization{
				ID:        tenantID,
				Name:      req.OrgName,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := s.repo.CreateOrganization(ctx, org); err != nil {
				s.logger.Error("Failed to create new organization", zap.Error(err), zap.String("org_name", req.OrgName))
				return nil, fmt.Errorf("failed to create organization: %w", err)
			}
			role = "org_admin"
		} else {
			// Public signup with no invite and no org creation attempt -> fallback to restricted driver
			// Or we could return an error if we strictly enforce tenant_id. The instructions say "public signup -> driver".
			// But tenant_id is NOT NULL in the DB. So we must put them somewhere.
			// Wait, the plan says: "if creating org -> org_admin. If public signup (no invite, no org creation) -> driver."
			// Since tenant_id is NOT NULL, let's create a "Personal Organization" for them, or require org creation.
			// Let's create a personal org for drivers to satisfy the constraint.
			tenantID = uuid.New()
			org := &models.Organization{
				ID:        tenantID,
				Name:      fmt.Sprintf("%s's Workspace", req.Name),
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := s.repo.CreateOrganization(ctx, org); err != nil {
				return nil, fmt.Errorf("failed to create workspace: %w", err)
			}
			role = "driver"
		}
	}

	user := &models.User{
		ID:                  uuid.New(),
		TenantID:               tenantID,
		Email:               strings.ToLower(req.Email),
		Role:                role,
		PasswordHash:        string(hashedPassword),
		CreatedAt:           now,
		CreatedBy:           creatorID,
		UpdatedAt:           now,
		UpdatedBy:           creatorID,
		MfaEnabled:          false,
		MfaMethods:          []string{},
		LastPasswordChange:  now,
		FailedLoginAttempts: sql.NullInt64{Int64: 0, Valid: true},
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		s.logger.Error("Failed to insert user into database", zap.Error(err), zap.String("email", user.Email))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Log security event
	err = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    user.ID,
		EventType: "USER_REGISTERED",
		IPAddress: req.Ip,
		UserAgent: req.UserAgent,
		Metadata:  fmt.Sprintf("email: %s, role: %s", user.Email, role),
	})
	if err != nil {
		s.logger.Error("Failed to log security event for registration", zap.Error(err), zap.String("user_id", user.ID.String()))
	}

	return user, nil
}

// MFA Methods

// GenerateMFASecret generates a new TOTP secret for a user
func (s *AuthService) GenerateMFASecret(ctx context.Context, userID uuid.UUID) (string, string, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	// Generate random 20-byte secret (160 bits)
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", "", fmt.Errorf("failed to generate secret: %w", err)
	}

	secretBase32 := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)
	// Remove padding and non-alphanumeric chars for base32
	secretBase32 = strings.ReplaceAll(secretBase32, "=", "")
	secretBase32 = strings.ReplaceAll(secretBase32, "0", "")
	secretBase32 = strings.ReplaceAll(secretBase32, "1", "")
	secretBase32 = strings.ReplaceAll(secretBase32, "8", "")
	secretBase32 = strings.ReplaceAll(secretBase32, "9", "")

	// Generate QR code URI
	otpauthURI := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=%d&period=%d",
		TOTPIssuer, user.Email, secretBase32, TOTPIssuer, TOTPDigits, TOTPPeriod)

	// Save secret to user
	if err := s.repo.UpdateMFASecret(ctx, userID, secretBase32); err != nil {
		s.logger.Warn("failed to update MFA secret in repo", zap.Error(err), zap.String("user_id", userID.String()))
	}

	return secretBase32, otpauthURI, nil
}

// VerifyTOTP verifies a TOTP code
func (s *AuthService) VerifyTOTP(secret, code string) (bool, error) {
	// Normalize inputs
	secret = strings.ToUpper(strings.ReplaceAll(secret, " ", ""))
	code = strings.ReplaceAll(code, " ", "")

	// Validate code format
	if len(code) != 6 {
		return false, ValidationError{Field: "code", Message: "code must be 6 digits"}
	}

	// Get current time step
	unix := time.Now().Unix()
	timeStep := unix / TOTPPeriod

	// Decode secret
	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return false, fmt.Errorf("invalid secret: %w", err)
	}

	// Check current and adjacent time windows (to handle clock skew)
	for i := -1; i <= 1; i++ {
		if s.verifyCode(secretBytes, timeStep+int64(i), code) {
			return true, nil
		}
	}

	return false, nil
}

// verifyCode verifies a single TOTP code for a given time step
func (s *AuthService) verifyCode(secret []byte, timeStep int64, expectedCode string) bool {
	// Generate HMAC-SHA1 hash
	h := hmac.New(sha1.New, secret)

	// Time step as big-endian 8 bytes
	var timeBytes [8]byte
	for i := 7; i >= 0; i-- {
		timeBytes[i] = byte(timeStep)
		timeStep >>= 8
	}

	h.Write(timeBytes[:])
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := int(hash[len(hash)-1] & 0x0F)
	binary.BigEndian.PutUint32(timeBytes[:4], 0)
	codeInt := int(binary.BigEndian.Uint32(hash[offset:])) & 0x7FFFFFFF
	codeInt %= 1000000

	generatedCode := fmt.Sprintf("%06d", codeInt)
	return generatedCode == expectedCode
}

// Compliance & Admin

func (s *AuthService) RevokeAllTokensGlobal(ctx context.Context) error {
	return s.repo.RevokeAllTokensGlobal(ctx)
}

func (s *AuthService) RevokeAllOrgTokens(ctx context.Context, tenantID uuid.UUID) error {
	return s.repo.RevokeAllOrgTokens(ctx, tenantID)
}

func (s *AuthService) AnonymizeUser(ctx context.Context, userID uuid.UUID) error {
	return s.repo.AnonymizeUser(ctx, userID)
}

// GenerateToken is a public helper for generating tokens in tests
func (s *AuthService) GenerateToken(userID, tenantID uuid.UUID, role string) (string, error) {
	user := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Role:     role,
	}
	permissions, _ := s.repo.GetRolePermissions(context.Background(), role)
	return s.generateAccessToken(user, permissions)
}

func (s *AuthService) getCountryFromIP(ipAddress string) string {
	if ipAddress == "" || ipAddress == "127.0.0.1" || ipAddress == "::1" || strings.HasPrefix(ipAddress, "192.168.") || strings.HasPrefix(ipAddress, "10.") {
		return "LOCAL"
	}
	return "US"
}

func (s *AuthService) calculateRiskScore(user *models.User, lastLogin *models.LoginHistory, ipAddress, userAgent, country string) float64 {
	if lastLogin == nil {
		return 0.0
	}
	score := 0.0
	if lastLogin.Country != "" && lastLogin.Country != country {
		score += 40.0
	}
	if lastLogin.IPAddress != "" && lastLogin.IPAddress != ipAddress {
		score += 25.0
	}
	if lastLogin.UserAgent != "" && lastLogin.UserAgent != userAgent {
		score += 22.0
	}
	if score > 100.0 {
		score = 100.0
	}
	return score
}

func (s *AuthService) comparePassword(password, hash string) (bool, error) {
	if strings.HasPrefix(hash, "$argon2id$") {
		parts := strings.Split(hash, "$")
		if len(parts) < 6 {
			return false, errors.New("invalid argon2id hash format")
		}
		var timeCost, memoryCost uint32
		var threads uint8
		_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memoryCost, &timeCost, &threads)
		if err != nil {
			return false, fmt.Errorf("invalid argon2id params: %w", err)
		}
		salt, err := base64.RawStdEncoding.DecodeString(parts[4])
		if err != nil {
			return false, fmt.Errorf("invalid argon2id salt: %w", err)
		}
		expectedKey, err := base64.RawStdEncoding.DecodeString(parts[5])
		if err != nil {
			return false, fmt.Errorf("invalid argon2id key: %w", err)
		}
		derivedKey := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, threads, uint32(len(expectedKey)))
		if subtle.ConstantTimeCompare(expectedKey, derivedKey) == 1 {
			return true, nil
		}
		return false, nil
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true, nil
	}
	return false, nil
}

func (s *AuthService) generateArgon2Hash(password string) (string, error) {
	salt := make([]byte, Argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, Argon2Iterations, Argon2Memory, Argon2Parallelism, Argon2KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)

	encodedHash := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		Argon2Memory, Argon2Iterations, Argon2Parallelism, b64Salt, b64Key)
	return encodedHash, nil
}

func (s *AuthService) getUserID(ctx context.Context) uuid.UUID {
	if uid, ok := ctxkeys.GetUserID(ctx); ok {
		return uid
	}
	return uuid.Nil
}

func (s *AuthService) VerifyRecoveryCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	codes, err := s.repo.GetUnusedRecoveryCodes(ctx, userID)
	if err != nil {
		return false, err
	}
	code = strings.TrimSpace(code)
	for _, rc := range codes {
		h := sha256.Sum256([]byte(code))
		hashedCodeHex := fmt.Sprintf("%x", h)
		if rc.CodeHash == hashedCodeHex || rc.CodeHash == code {
			_ = s.repo.MarkRecoveryCodeUsed(ctx, rc.ID)
			return true, nil
		}
		if err := bcrypt.CompareHashAndPassword([]byte(rc.CodeHash), []byte(code)); err == nil {
			_ = s.repo.MarkRecoveryCodeUsed(ctx, rc.ID)
			return true, nil
		}
	}
	return false, nil
}

func (s *AuthService) RegisterUser(ctx context.Context, req pb.RegisterRequest) (*models.User, error) {
	return s.Register(ctx, &req)
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (*models.User, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
			keys, err := s.repo.GetActiveSigningKeys(ctx)
			if err == nil && len(keys) > 0 {
				block, _ := pem.Decode([]byte(keys[0].PrivateKeyEnc))
				if block != nil {
					if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
						return &key.PublicKey, nil
					}
					if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
						if rsaKey, ok := key.(*rsa.PrivateKey); ok {
							return &rsaKey.PublicKey, nil
						}
					}
				}
			}
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		userIDStr, ok = claims["user_id"].(string)
	}
	if !ok {
		return nil, errors.New("user ID missing from claims")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID in claims")
	}

	return s.repo.GetUserByID(ctx, userID)
}

func (s *AuthService) CheckPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string) (bool, error) {
	return s.repo.CheckPermission(ctx, userID, tenantID, permission)
}

func (s *AuthService) ValidateInvite(ctx context.Context, code string) (*models.Invite, error) {
	invite, err := s.repo.GetInviteByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if invite == nil || invite.Used || time.Now().After(invite.Expiry) {
		return nil, errors.New("invite is invalid, used, or expired")
	}
	return invite, nil
}

func (s *AuthService) EnableMFA(ctx context.Context, userID uuid.UUID, secret, code string) ([]string, error) {
	valid, err := s.VerifyTOTP(secret, code)
	if err != nil || !valid {
		return nil, errors.New("invalid TOTP verification code")
	}

	if err := s.repo.EnableMFA(ctx, userID, secret); err != nil {
		return nil, fmt.Errorf("failed to enable MFA: %w", err)
	}

	codes := make([]string, 8)
	hashes := make([]string, 8)
	for i := 0; i < 8; i++ {
		rawCode := uuid.New().String()[:10]
		codes[i] = rawCode
		h := sha256.Sum256([]byte(rawCode))
		hashes[i] = fmt.Sprintf("%x", h)
	}

	if err := s.repo.SaveRecoveryCodes(ctx, userID, hashes); err != nil {
		s.logger.Error("failed to save recovery codes", zap.Error(err))
	}

	return codes, nil
}

func (s *AuthService) DisableMFA(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.MfaSecret.Valid && user.MfaSecret.String != "" && code != "" {
		valid, err := s.VerifyTOTP(user.MfaSecret.String, code)
		if err != nil || !valid {
			return errors.New("invalid MFA code")
		}
	}

	return s.repo.DisableMFA(ctx, userID)
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return nil
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour)

	if err := s.repo.CreatePasswordResetToken(ctx, user.ID, token, expiresAt); err != nil {
		return err
	}

	_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    user.ID,
		EventType: "PASSWORD_RESET_REQUESTED",
		Metadata:  "token_issued",
	})
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	resetToken, err := s.repo.GetPasswordResetToken(ctx, token)
	if err != nil || resetToken == nil {
		return errors.New("invalid or expired reset token")
	}

	if resetToken.Used || time.Now().After(resetToken.ExpiresAt) {
		return errors.New("invalid or expired reset token")
	}

	hash, err := s.generateArgon2Hash(newPassword)
	if err != nil {
		return err
	}

	if err := s.repo.UpdatePassword(ctx, resetToken.UserID, hash); err != nil {
		return err
	}

	_ = s.repo.UsePasswordResetToken(ctx, token)
	_ = s.repo.RevokeAllUserTokens(ctx, resetToken.UserID)

	_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    resetToken.UserID,
		EventType: "PASSWORD_RESET_COMPLETED",
	})
	return nil
}

func (s *AuthService) LoginWithOIDC(ctx context.Context, providerName, code, state, ipAddress, userAgent string) (*models.AuthResponse, error) {
	provider, err := s.repo.GetOidcProviderByName(ctx, providerName)
	if err != nil || provider == nil || !provider.Enabled {
		return nil, errors.New("invalid or disabled OIDC provider")
	}

	user, err := s.repo.GetUserByOidcLink(ctx, provider.ID, code)
	if err != nil || user == nil {
		return nil, errors.New("user not linked to OIDC provider")
	}

	return s.issueTokens(ctx, user, ipAddress, userAgent)
}

func (s *AuthService) GetJWKS(ctx context.Context) (*models.JWKS, error) {
	keys, err := s.repo.GetActiveSigningKeys(ctx)
	if err != nil {
		return nil, err
	}

	jwks := &models.JWKS{
		Keys: make([]models.JWK, 0, len(keys)),
	}

	for _, key := range keys {
		block, _ := pem.Decode([]byte(key.PrivateKeyEnc))
		var nStr, eStr string
		if block != nil {
			if privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
				nStr = base64.RawURLEncoding.EncodeToString(privKey.N.Bytes())
				eBytes := make([]byte, 4)
				binary.BigEndian.PutUint32(eBytes, uint32(privKey.E))
				eBytes = bytes.TrimLeft(eBytes, "\x00")
				eStr = base64.RawURLEncoding.EncodeToString(eBytes)
			}
		}
		jwks.Keys = append(jwks.Keys, models.JWK{
			Kty: "RSA",
			Kid: key.KID,
			Use: "sig",
			Alg: "RS256",
			N:   nStr,
			E:   eStr,
		})
	}
	return jwks, nil
}

func (s *AuthService) CreateAPIKey(ctx context.Context, userID, tenantID uuid.UUID, name string, expiresAt *time.Time) (string, error) {
	rawSecret := uuid.New().String()
	prefix := "ff_" + uuid.New().String()[:8]
	fullKey := prefix + "." + rawSecret

	h := sha256.Sum256([]byte(rawSecret))
	secretHash := fmt.Sprintf("%x", h)

	apiKey := &models.APIKey{
		ID:          uuid.New(),
		TenantID:    tenantID,
		UserID:      userID,
		Name:        name,
		Prefix:      prefix,
		SecretHash:  secretHash,
		Permissions: []string{},
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
		Revoked:     false,
	}

	if err := s.repo.SaveAPIKey(ctx, apiKey); err != nil {
		return "", err
	}
	return fullKey, nil
}

func (s *AuthService) ListAPIKeys(ctx context.Context, tenantID uuid.UUID) ([]models.APIKey, error) {
	return s.repo.ListAPIKeys(ctx, tenantID)
}

func (s *AuthService) RevokeAPIKey(ctx context.Context, id uuid.UUID) error {
	return s.repo.RevokeAPIKey(ctx, id)
}

func (s *AuthService) CreateServiceAccount(ctx context.Context, tenantID uuid.UUID, name, description string) (*models.ServiceAccount, string, error) {
	clientID := "sa_" + uuid.New().String()[:8]
	clientSecret := uuid.New().String()
	h := sha256.Sum256([]byte(clientSecret))
	secretHash := fmt.Sprintf("%x", h)

	sa := &models.ServiceAccount{
		ID:               uuid.New(),
		TenantID:         tenantID,
		Name:             name,
		Description:      description,
		ClientID:         clientID,
		ClientSecretHash: secretHash,
		Role:             "service_account",
		Permissions:      []string{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.CreateServiceAccount(ctx, sa); err != nil {
		return nil, "", err
	}
	return sa, clientSecret, nil
}

func (s *AuthService) ListServiceAccounts(ctx context.Context, tenantID uuid.UUID) ([]models.ServiceAccount, error) {
	return s.repo.ListServiceAccounts(ctx, tenantID)
}

func (s *AuthService) EmergencyRevoke(ctx context.Context) error {
	return s.repo.RevokeAllTokensGlobal(ctx)
}

func (s *AuthService) SearchUsers(ctx context.Context, query string) ([]*models.User, error) {
	if query == "" {
		return []*models.User{}, nil
	}
	user, err := s.repo.GetUserByEmail(ctx, query)
	if err == nil && user != nil {
		return []*models.User{user}, nil
	}
	if uID, err := uuid.Parse(query); err == nil {
		user, err := s.repo.GetUserByID(ctx, uID)
		if err == nil && user != nil {
			return []*models.User{user}, nil
		}
	}
	return []*models.User{}, nil
}


