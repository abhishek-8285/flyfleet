package service

import (
	"context"
	"database/sql"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core-service/internal/domains/auth/models"
	pb "github.com/flyfleet/proto/auth/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"encoding/json"
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
	AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID, orgID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID, orgID uuid.UUID) ([]models.Role, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID, orgID uuid.UUID) ([]string, error)
	CheckPermission(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, permission string) (bool, error)
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
	ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]models.APIKey, error)
	RevokeAPIKey(ctx context.Context, id uuid.UUID) error
	UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error
	
	GetServiceAccount(ctx context.Context, id uuid.UUID) (*models.ServiceAccount, error)
	CreateServiceAccount(ctx context.Context, sa *models.ServiceAccount) error
	ListServiceAccounts(ctx context.Context, orgID uuid.UUID) ([]models.ServiceAccount, error)
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
	logger        logger.Logger
	repo          AuthRepositoryInterface
	identity      IdentityService
	jwtSecret     string
	redis         *redis.Client
	nats          *nats.Conn
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

func (s *AuthService) Login(ctx context.Context, email, password, orgIDStr, ipAddress, userAgent string) (*models.AuthResponse, error) {
	s.logger.Info("Login attempt started", zap.String("email", email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error("GetUserByEmail failed", zap.Error(err), zap.String("email", email))
		return nil, err
	}
	s.logger.Info("User found", zap.String("email", email), zap.String("userID", user.ID.String()))


	// Override org if provided
	if orgIDStr != "" {
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			return nil, errors.New("invalid org_id format")
		}
		
		orgs, err := s.repo.GetOrganizationsByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		
		found := false
		for _, o := range orgs {
			if o.ID == orgID {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("invalid organization")
		}
		user.OrgID = orgID
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

	// MFA Requirement (check if role requires it OR if risk score is high)
	if user.Role == "admin" || user.Role == "super_admin" || user.MfaEnabled || riskScore > 60 {
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

func (s *AuthService) issueTokens(ctx context.Context, user *models.User, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// Fetch permissions via IdentityService/Repository (New RBAC)
	permissions, _ := s.repo.GetUserPermissions(ctx, user.ID, user.OrgID)
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

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
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
		OrgID:  user.OrgID.String(),
		Role:   user.Role,
		Type:   "mfa_pending",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": claims.UserID,
		"org_id":  claims.OrgID,
		"role":    claims.Role,
		"type":    claims.Type,
		"exp":     time.Now().Add(5 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// Validate refresh token
	refreshToken, err := s.repo.GetRefreshToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, err
	}

	if refreshToken.Revoked {
		// ALARM: Potential token theft!
		_ = s.repo.RevokeAllUserTokens(ctx, refreshToken.UserID)
		_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
			UserID:    refreshToken.UserID,
			EventType: "REFRESH_TOKEN_REUSE_DETECTED",
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Metadata:  fmt.Sprintf("token_id: %s", refreshToken.ID),
		})
		return nil, errors.New("security alert: token reuse detected. all sessions revoked")
	}

	// Single-use: Revoke old refresh token
	_ = s.repo.RevokeRefreshToken(ctx, refreshTokenStr)

	// Get user
	user, err := s.repo.GetUserByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, err
	}

	// Issue new tokens
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
		// Log failed MFA verification
		_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
			UserID:    user.ID,
			EventType: "LOGIN_MFA_FAILED",
			IPAddress: "",
			UserAgent: "",
			Metadata:  fmt.Sprintf("type: %s", map[bool]string{true: "recovery", false: "totp"}[isRecovery]),
		})
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
	org, _ := s.repo.GetOrganizationByID(context.Background(), user.OrgID)
	tenantType := ""
	if org != nil {
		tenantType = org.Type
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":         user.ID.String(),
		"iss":         "FlyFleet",
		"aud":         "fleet-dashboard",
		"org_id":      user.OrgID.String(),
		"rol":         user.Role,
		"ten":         tenantType, // tenant_type claim
		"type":        "access",
		"prm":         permissions,
		"pmsk":        user.PermissionMask,
		"jti":         jti,
		"exp":         time.Now().Add(AccessTokenExpiry).Unix(),
		"iat":         time.Now().Unix(),
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
	var orgIDStr string
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
			"org_id":    orgIDStr,  // would be fetched if needed
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

	var orgID uuid.UUID
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

		orgID = invite.OrgID
		role = invite.Role

		// Mark invite as used
		_ = s.repo.MarkInviteUsed(ctx, invite.ID)

	} else {
		// Not joining via invite.
		if req.OrgName != "" {
			// Creating a new organization
			orgID = uuid.New()
			org := &models.Organization{
				ID:        orgID,
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
			// Or we could return an error if we strictly enforce org_id. The instructions say "public signup -> driver".
			// But org_id is NOT NULL in the DB. So we must put them somewhere. 
			// Wait, the plan says: "if creating org -> org_admin. If public signup (no invite, no org creation) -> driver."
			// Since org_id is NOT NULL, let's create a "Personal Organization" for them, or require org creation.
			// Let's create a personal org for drivers to satisfy the constraint.
			orgID = uuid.New()
			org := &models.Organization{
				ID:        orgID,
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
		OrgID:               orgID,
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

// EnableMFA enables MFA for a user and returns recovery codes
func (s *AuthService) EnableMFA(ctx context.Context, userID uuid.UUID, secret, code string) ([]string, error) {
	// Verify the code first
	valid, err := s.VerifyTOTP(secret, code)
	if err != nil || !valid {
		return nil, errors.New("invalid TOTP code")
	}

	// Update user in repository
	if err := s.repo.EnableMFA(ctx, userID, secret); err != nil {
		return nil, fmt.Errorf("failed to enable MFA: %w", err)
	}

	// Generate recovery codes
	recoveryCodes, err := s.GenerateRecoveryCodes(ctx, userID)
	if err != nil {
		s.logger.Error("failed to generate recovery codes during MFA enable", zap.Error(err), zap.String("user_id", userID.String()))
		// We could still return success but without codes, but it's better to log.
	}

	// Log security event
	_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    userID,
		EventType: "MFA_ENABLED",
		IPAddress: "",
		UserAgent: "",
		Metadata:  "method: totp",
	})

	return recoveryCodes, nil
}

// DisableMFA disables MFA for a user
func (s *AuthService) DisableMFA(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.MfaEnabled {
		return errors.New("MFA is not enabled")
	}

	// Verify current code
	secret := ""
	if user.MfaSecret.Valid {
		secret = user.MfaSecret.String
	}
	valid, err := s.VerifyTOTP(secret, code)
	if err != nil || !valid {
		return errors.New("invalid TOTP code")
	}

	// Update user
	if err := s.repo.DisableMFA(ctx, userID); err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	// Log security event
	_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    userID,
		EventType: "MFA_DISABLED",
		IPAddress: "",
		UserAgent: "",
		Metadata:  "",
	})

	return nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (*models.User, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid in token header")
		}

		key, err := s.repo.GetSigningKey(ctx, kid)
		if err != nil || key == nil {
			return nil, errors.New("signing key not found")
		}

		block, _ := pem.Decode([]byte(key.PublicKey))
		if block == nil {
			return nil, errors.New("failed to parse public key PEM")
		}
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
		}
		return pubKey, nil
	})
	
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	
	if iss, _ := claims["iss"].(string); iss != "FlyFleet" {
		return nil, errors.New("invalid token issuer")
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		userIDStr, _ = claims["user_id"].(string)
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user_id in claims")
	}

	return s.repo.GetUserByID(ctx, userID)
}

func (s *AuthService) SearchUsers(ctx context.Context, query string) ([]models.User, error) {
	// For now, this is a simple email search or ID search
	if uid, err := uuid.Parse(query); err == nil {
		user, err := s.repo.GetUserByID(ctx, uid)
		if err != nil {
			return nil, err
		}
		return []models.User{*user}, nil
	}
	
	user, err := s.repo.GetUserByEmail(ctx, query)
	if err != nil {
		// If not found by email, return empty instead of error for search
		return []models.User{}, nil
	}
	return []models.User{*user}, nil
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

// Password Reset Methods

// RequestPasswordReset generates a reset token and sends it to the user's email
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists - security best practice
		return nil
	}

	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Token expires in 1 hour
	expiresAt := time.Now().Add(1 * time.Hour)

	// Store token
	if err := s.repo.CreatePasswordResetToken(ctx, user.ID, token, expiresAt); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// In production, send email with reset link
	// resetLink := fmt.Sprintf("https://app.flyfleet.com/reset-password?token=%s", token)

	// Log security event
	_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    user.ID,
		EventType: "PASSWORD_RESET_REQUESTED",
		IPAddress: "",
		UserAgent: "",
		Metadata:  fmt.Sprintf("email: %s", email),
	})

	return nil
}

// ResetPassword validates the token and updates the password
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Get and validate reset token
	resetToken, err := s.repo.GetPasswordResetToken(ctx, token)
	if err != nil {
		return ValidationError{Field: "token", Message: "invalid or expired reset token"}
	}

	// Hash new password
	hashedPassword, err := s.generateArgon2Hash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.repo.UpdatePassword(ctx, resetToken.UserID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	_ = s.repo.UsePasswordResetToken(ctx, token)

	// Revoke all user tokens (force re-login)
	_ = s.repo.RevokeAllUserTokens(ctx, resetToken.UserID)

	// Log security event
	_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    resetToken.UserID,
		EventType: "PASSWORD_RESET_COMPLETED",
		IPAddress: "",
		UserAgent: "",
		Metadata:  "",
	})

	return nil
}

func (s *AuthService) generateArgon2Hash(password string) (string, error) {
	salt := make([]byte, Argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, uint32(Argon2Iterations), uint32(Argon2Memory), uint8(Argon2Parallelism), uint32(Argon2KeyLength))

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, Argon2Memory, Argon2Iterations, Argon2Parallelism, b64Salt, b64Hash)

	return encoded, nil
}

func (s *AuthService) comparePassword(password, encodedHash string) (bool, error) {
	// Support legacy Bcrypt
	if strings.HasPrefix(encodedHash, "$2a$") || strings.HasPrefix(encodedHash, "$2b$") {
		err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password))
		return err == nil, nil
	}

	// Handle Argon2id
	if !strings.HasPrefix(encodedHash, "$argon2id$") {
		return false, errors.New("invalid hash format")
	}

	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid argon2id hash format")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, errors.New("incompatible argon2id version")
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(decodedHash)))

	return hmac.Equal(hash, decodedHash), nil
}

func (s *AuthService) GenerateRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	codes := make([]string, 10)
	hashes := make([]string, 10)

	for i := 0; i < 10; i++ {
		// Generate 10-char random code
		bytes := make([]byte, 8)
		if _, err := rand.Read(bytes); err != nil {
			return nil, err
		}
		// Using a hex string for readability
		codes[i] = fmt.Sprintf("%x", bytes)[:10]
		
		hash, err := s.generateArgon2Hash(codes[i])
		if err != nil {
			return nil, err
		}
		hashes[i] = hash
	}

	if err := s.repo.SaveRecoveryCodes(ctx, userID, hashes); err != nil {
		return nil, err
	}

	return codes, nil
}

func (s *AuthService) VerifyRecoveryCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	unusedCodes, err := s.repo.GetUnusedRecoveryCodes(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, c := range unusedCodes {
		ok, err := s.comparePassword(code, c.CodeHash)
		if err == nil && ok {
			// Mark code as used
			_ = s.repo.MarkRecoveryCodeUsed(ctx, c.ID)
			return true, nil
		}
	}

	return false, nil
}

func (s *AuthService) calculateRiskScore(user *models.User, lastLogin *models.LoginHistory, currentIP, currentUA, currentCountry string) float64 {
	score := 0.0
	if lastLogin == nil {
		return 30.0
	}
	if lastLogin.Country != "" && currentCountry != "" && lastLogin.Country != currentCountry {
		if time.Since(lastLogin.CreatedAt) < 6*time.Hour {
			score += 60.0
		} else {
			score += 30.0
		}
	}
	if lastLogin.UserAgent != currentUA {
		score += 20.0
	}
	if lastLogin.IPAddress != currentIP {
		score += 10.0
	}
	if score > 100 {
		score = 100
	}
	return score
}

func (s *AuthService) getCountryFromIP(ip string) string {
	if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "127.0.") || ip == "::1" || ip == "" {
		return "US"
	}
	h := sha1.New()
	h.Write([]byte(ip))
	sum := h.Sum(nil)
	countries := []string{"US", "GB", "IN", "DE", "FR", "CN", "BR"}
	return countries[int(sum[0])%len(countries)]
}

func (s *AuthService) LoginWithOIDC(ctx context.Context, providerName, code, state, ipAddress, userAgent string) (*models.AuthResponse, error) {
	provider, err := s.repo.GetOidcProviderByName(ctx, providerName)
	if err != nil || provider == nil {
		return nil, errors.New("provider not found or disabled")
	}

	sub, email, err := s.verifyOidcToken(ctx, provider, code)
	if err != nil {
		return nil, fmt.Errorf("OIDC verification failed: %w", err)
	}

	user, err := s.repo.GetUserByOidcLink(ctx, provider.ID, sub)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = s.repo.GetUserByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, errors.New("user not found and auto-provisioning disabled")
		}
		err = s.repo.LinkUserToOidc(ctx, &models.UserOidcLink{
			UserID:         user.ID,
			ProviderID:     provider.ID,
			ProviderUserID: sub,
			Email:          email,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to link OIDC account: %w", err)
		}
	}

	lastLogin, _ := s.repo.GetLastLogin(ctx, user.ID)
	country := s.getCountryFromIP(ipAddress)
	riskScore := s.calculateRiskScore(user, lastLogin, ipAddress, userAgent, country)

	if user.MfaEnabled || riskScore > 80 {
		tempToken, err := s.generateTempToken(user)
		if err != nil {
			return nil, err
		}
		return &models.AuthResponse{
			MfaRequired: true,
			TempToken:   tempToken,
			User:        user,
		}, nil
	}

	permissions, _ := s.repo.GetRolePermissions(ctx, user.Role)
	accessToken, err := s.generateAccessToken(user, permissions)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.generateRefreshToken(ctx, user.ID, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	_ = s.repo.LogLoginAttempt(ctx, &models.LoginHistory{
		UserID:    user.ID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Country:   country,
		Success:   true,
		RiskScore: riskScore,
	})

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(AccessTokenExpiry.Seconds()),
		User:         user,
	}, nil
}

func (s *AuthService) verifyOidcToken(ctx context.Context, provider *models.OidcProvider, code string) (sub, email string, err error) {
	// TODO: Implement actual OIDC verification using provider details
	return "", "", errors.New("OIDC verification not implemented")
}


func (s *AuthService) RotateSigningKeys(ctx context.Context) error {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})
	
	pubDER, _ := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	})

	now := time.Now()
	newKey := &models.SigningKey{
		KID:           uuid.New().String(),
		PrivateKeyEnc: string(privPEM),
		PublicKey:     string(pubPEM),
		Status:        "active",
		Algorithm:     "RS256",
		CreatedAt:     now,
		ActivatedAt:   &now,
	}

	activeKeys, _ := s.repo.GetActiveSigningKeys(ctx)
	for _, k := range activeKeys {
		if k.Status == "active" {
			_ = s.repo.UpdateSigningKeyStatus(ctx, k.KID, "previous")
		}
	}

	return s.repo.SaveSigningKey(ctx, newKey)
}



func (s *AuthService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// GDPR: Right to be Forgotten
	// 1. Revoke all tokens
	_ = s.repo.RevokeAllUserTokens(ctx, userID)
	
	// 2. Anonymize/Delete user data
	return s.repo.AnonymizeUser(ctx, userID)
}

func (s *AuthService) AuthenticateServiceAccount(ctx context.Context, id uuid.UUID, secret string) (string, error) {
	sa, err := s.repo.GetServiceAccount(ctx, id)
	if err != nil || sa == nil {
		return "", errors.New("service account not found")
	}

	ok, err := s.comparePassword(secret, sa.SecretHash)
	if err != nil || !ok {
		return "", errors.New("invalid service account secret")
	}

	// Generate M2M Access Token (usually longer expiry and contains permissions)
	return s.generateAccessToken(&models.User{
		ID:    sa.ID,
		OrgID: sa.OrgID,
		Role:  "service_account",
	}, sa.Permissions)
}

func (s *AuthService) ValidateAPIKey(ctx context.Context, keyStr string) (*models.APIKey, error) {
	// keyStr format: prefix.secret
	parts := strings.Split(keyStr, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid api key format")
	}

	prefix := parts[0]
	secret := parts[1]

	apiKey, err := s.repo.GetAPIKeyByPrefix(ctx, prefix)
	if err != nil || apiKey == nil {
		return nil, errors.New("api key not found")
	}

	if apiKey.Revoked {
		return nil, errors.New("api key revoked")
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("api key expired")
	}

	ok, err := s.comparePassword(secret, apiKey.SecretHash)
	if err != nil || !ok {
		return nil, errors.New("invalid api key secret")
	}

	// Update last used
	_ = s.repo.UpdateAPIKeyLastUsed(ctx, apiKey.ID)

	return apiKey, nil
}

func (s *AuthService) CreateAPIKey(ctx context.Context, userID, orgID uuid.UUID, name string, expiresAt *time.Time) (string, error) {
	// Generate random secret
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	secret := base64.URLEncoding.EncodeToString(bytes)
	
	hash, err := s.generateArgon2Hash(secret)
	if err != nil {
		return "", err
	}

	prefix := strings.ReplaceAll(uuid.New().String(), "-", "")[:12]

	apiKey := &models.APIKey{
		ID:          uuid.New(),
		OrgID:       orgID,
		UserID:      userID,
		Name:        name,
		Prefix:      prefix,
		SecretHash:  string(hash),
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		Revoked:     false,
	}

	if err := s.repo.SaveAPIKey(ctx, apiKey); err != nil {
		return "", err
	}

	// Format: prefix.secret
	return fmt.Sprintf("%s.%s", prefix, secret), nil
}

func (s *AuthService) ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]models.APIKey, error) {
	return s.repo.ListAPIKeys(ctx, orgID)
}

func (s *AuthService) RevokeAPIKey(ctx context.Context, id uuid.UUID) error {
	return s.repo.RevokeAPIKey(ctx, id)
}

func (s *AuthService) CreateServiceAccount(ctx context.Context, orgID uuid.UUID, name, description string) (*models.ServiceAccount, string, error) {
	// Generate random secret
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, "", err
	}
	secret := base64.URLEncoding.EncodeToString(bytes)
	
	hash, err := s.generateArgon2Hash(secret)
	if err != nil {
		return nil, "", err
	}

	sa := &models.ServiceAccount{
		ID:               uuid.New(),
		OrgID:            orgID,
		Name:             name,
		Description:      description,
		ClientID:         uuid.New().String(),
		ClientSecretHash: string(hash),
		Role:             "service_account",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.CreateServiceAccount(ctx, sa); err != nil {
		return nil, "", err
	}

	return sa, secret, nil
}

func (s *AuthService) ListServiceAccounts(ctx context.Context, orgID uuid.UUID) ([]models.ServiceAccount, error) {
	return s.repo.ListServiceAccounts(ctx, orgID)
}

func (s *AuthService) EmergencyRevoke(ctx context.Context) error {
	return s.repo.RevokeAllTokensGlobal(ctx)
}

// InviteValidationResult is a secure subset of the Invite model sent to the frontend
type InviteValidationResult struct {
	Valid   bool   `json:"valid"`
	OrgID   string `json:"org_id"`
	Role    string `json:"role"`
	OrgName string `json:"org_name"`
}

// ValidateInvite verifies an invite code's mathematical presence and expiration without marking it used.
func (s *AuthService) ValidateInvite(ctx context.Context, code string) (*InviteValidationResult, error) {
	invite, err := s.repo.GetInviteByCode(ctx, code)
	if err != nil {
		return &InviteValidationResult{Valid: false}, nil
	}
	
	if invite.Used || time.Now().After(invite.Expiry) {
		return &InviteValidationResult{Valid: false}, nil
	}

	org, err := s.repo.GetOrganizationByID(ctx, invite.OrgID)
	orgName := ""
	if err == nil && org != nil {
		orgName = org.Name
	}

	return &InviteValidationResult{
		Valid:   true,
		OrgID:   invite.OrgID.String(),
		Role:    invite.Role,
		OrgName: orgName,
	}, nil
}

// GenerateInvite creates a mathematically secure UUIDv4 invite code for a given organization and role.
func (s *AuthService) GenerateInvite(ctx context.Context, orgID uuid.UUID, role string, expiryHours int) (string, error) {
	creatorID := s.getUserID(ctx)
	
	// ENFORCEMENT: Never use low-entropy alphanumeric tokens. Force UUIDv4 for all invites
	// to mathematically blockade credential stuffing and brute force scanners.
	code := uuid.New().String()
	
	invite := &models.Invite{
		ID:        uuid.New(),
		Code:      code,
		OrgID:     orgID,
		Role:      role,
		Expiry:    time.Now().Add(time.Duration(expiryHours) * time.Hour),
		CreatedAt: time.Now(),
		CreatedBy: creatorID,
		Used:      false,
	}

	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return "", fmt.Errorf("failed to create invite: %w", err)
	}

	return code, nil
}

func (s *AuthService) GetJWKS(ctx context.Context) (*models.JWKS, error) {
	keys, err := s.repo.GetActiveSigningKeys(ctx)
	if err != nil {
		return nil, err
	}

	var jwks models.JWKS
	for _, k := range keys {
		// Parse public key to get N and E
		block, _ := pem.Decode([]byte(k.PublicKey))
		if block == nil {
			continue
		}
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			continue
		}

		rsaPubKey, ok := pubKey.(*rsa.PublicKey)
		if !ok {
			continue
		}

		jwks.Keys = append(jwks.Keys, models.JWK{
			Kty: "RSA",
			Kid: k.KID,
			Use: "sig",
			Alg: k.Algorithm,
			N:   base64.RawURLEncoding.EncodeToString(rsaPubKey.N.Bytes()),
			E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(rsaPubKey.E)).Bytes()),
		})
	}

	return &jwks, nil
}

type RegisterRequest struct {
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Password   string  `json:"password"`
	InviteCode *string `json:"invite_code"`
	OrgName    *string `json:"org_name"`
}

type RegisterResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

// RegisterUser implements the production-ready logic for user registration handling multi-tenant logic.
func (s *AuthService) RegisterUser(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	// STEP 1 - Validate Input
	if err := s.validateEmail(ctx, req.Email); err != nil {
		return nil, err
	}
	if err := s.validatePassword(req.Password); err != nil {
		return nil, err
	}
	if req.InviteCode == nil && req.OrgName == nil {
		return nil, fmt.Errorf("either InviteCode or OrgName must be provided")
	}

	now := time.Now()
	creatorID := s.getUserID(ctx)
	var orgID uuid.UUID
	var role string

	// STEP 2 - Handle Invite Flow
	if req.InviteCode != nil && *req.InviteCode != "" {
		invite, err := s.repo.GetInviteByCode(ctx, *req.InviteCode)
		if err != nil {
			return nil, fmt.Errorf("invalid or expired invite code")
		}
		if invite.Used || now.After(invite.Expiry) {
			return nil, fmt.Errorf("invite code is expired or has already been used")
		}

		orgID = invite.OrgID
		role = invite.Role

		// Mark invite as used
		_ = s.repo.MarkInviteUsed(ctx, invite.ID)
	} else {
		// STEP 3 - Handle New Organization Flow
		if req.OrgName == nil || *req.OrgName == "" {
			return nil, fmt.Errorf("OrgName is required when not using an invite code")
		}

		orgID = uuid.New()
		org := &models.Organization{
			ID:        orgID,
			Name:      *req.OrgName,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.repo.CreateOrganization(ctx, org); err != nil {
			s.logger.Error("Failed to create new organization", zap.Error(err), zap.String("org_name", *req.OrgName))
			return nil, fmt.Errorf("failed to create organization: %w", err)
		}
		role = "org_admin"
	}

	// STEP 4 - Create User
	hashedPassword, err := s.generateArgon2Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:                  uuid.New(),
		OrgID:               orgID,
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

	// STEP 5 - Generate Tokens
	// issueTokens generates access_token and refresh_token
	authResp, err := s.issueTokens(ctx, user, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// STEP 6 - Return Response
	return &RegisterResponse{
		User:         user,
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
	}, nil
}
