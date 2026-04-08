package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/iam/config"
	"github.com/flyfleet/fleet-core/internal/domains/iam/email"
	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	redis2 "github.com/flyfleet/fleet-core/internal/domains/iam/redis"
	"github.com/flyfleet/fleet-core/internal/domains/iam/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrAccountLocked      = errors.New("account_locked")
	ErrMFARequired        = errors.New("mfa_required")
	ErrInvalidMFACode     = errors.New("invalid_mfa_code")
	ErrUserNotFound       = errors.New("user_not_found")
)

// JWTManager handles JWT token operations
type JWTManager interface {
	GenerateToken(claims jwt.MapClaims, expiration time.Duration) (string, error)
	VerifyToken(tokenString string) (*jwt.Token, error)
	ExtractClaims(tokenString string) (jwt.MapClaims, error)
}

type AuthService struct {
	userRepo     *repository.UserRepository
	orgRepo      *repository.OrgRepository
	roleRepo     *repository.RoleRepository
	sessionRepo  *repository.SessionRepository
	passwordRepo *repository.PasswordHistoryRepository
	redis        *redis2.Client
	jwtManager   JWTManager
	config       config.SecurityConfig
	log          *zap.Logger
	// publisher    *nats.Publisher // unused stub
	emailSender  *email.EmailService
	passwordSvc  *PasswordService
}

func NewAuthService(
	userRepo *repository.UserRepository,
	orgRepo *repository.OrgRepository,
	roleRepo *repository.RoleRepository,
	sessionRepo *repository.SessionRepository,
	passwordRepo *repository.PasswordHistoryRepository,
	redis *redis2.Client,
	jwtManager JWTManager,
	cfg config.SecurityConfig,
	log *zap.Logger,
	// publisher *nats.Publisher, // unused stub
	emailSender *email.EmailService,
	passwordSvc *PasswordService,
) *AuthService {
	return &AuthService{
		userRepo: userRepo, orgRepo: orgRepo, roleRepo: roleRepo,
		sessionRepo: sessionRepo, passwordRepo: passwordRepo,
		redis: redis, jwtManager: jwtManager, config: cfg,
		log: log, // publisher: publisher,
		emailSender: emailSender, passwordSvc: passwordSvc,
	}
}

// Identify checks if a user exists and returns org info + MFA status (anti-enumeration safe)
func (s *AuthService) Identify(ctx context.Context, email string) (*models.IdentifyResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	resp := &models.IdentifyResponse{
		Exists:       user != nil,
		MFARequired:  false,
		MFAMethods:   []string{},
		MFARemembered: false,
		SSOProviders: []string{"google", "microsoft"},
		AccountLocked: false,
	}

	if user == nil {
		// Anti-enumeration: do bcrypt compare anyway (constant time)
		dummyHash, _ := bcrypt.GenerateFromPassword([]byte("dummy"), 4)
		bcrypt.CompareHashAndPassword(dummyHash, []byte("dummy"))
		return resp, nil
	}

	resp.EmailVerified = user.EmailVerified

	// Check lockout
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		resp.AccountLocked = true
		return resp, nil
	}

	// Get user's organizations
	orgs, err := s.orgRepo.GetUserOrganizations(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	resp.Organizations = orgs

	// MFA status
	if user.MFAEnabled {
		resp.MFARequired = true
		resp.MFAMethods = []string{"totp"}
	}

	return resp, nil
}

// Login authenticates a user with email + password
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest, ip, userAgent string) (*models.AuthTokensResponse, *models.MFARequiredResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, err
	}

	// Anti-enumeration: if user not found, do dummy compare
	if user == nil {
		dummyHash, _ := bcrypt.GenerateFromPassword([]byte("dummy"), 4)
		bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
		return nil, nil, ErrInvalidCredentials
	}

	// Check lockout
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, nil, ErrAccountLocked
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.handleFailedLogin(ctx, user, ip)
		return nil, nil, ErrInvalidCredentials
	}

	// Successful password check — check if MFA needed
	if user.MFAEnabled {
		// Generate temp action token for MFA
		tempToken, err := s.generateActionToken(user.ID, req.OrgID, "mfa")
		if err != nil {
			return nil, nil, err
		}
		return nil, &models.MFARequiredResponse{
			MFARequired: true,
			TempToken:   tempToken,
			MFAMethods:  []string{"totp"},
			ExpiresIn:   int(s.config.JWTActionTTL.Seconds()),
		}, nil
	}

	// No MFA — generate full tokens
	tokens, err := s.generateLoginTokens(ctx, user, req.OrgID, ip, userAgent, req.RememberDevice)
	if err != nil {
		return nil, nil, err
	}
	return tokens, nil, nil
}

// VerifyMFA verifies an MFA code and returns full tokens
func (s *AuthService) VerifyMFA(ctx context.Context, req models.MFAVerifyRequest, ip, userAgent string) (*models.AuthTokensResponse, error) {
	// Verify temp token
	claims, err := s.jwtManager.ExtractClaims(req.TempToken)
	if err != nil {
		return nil, errors.New("invalid_token")
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "mfa" {
		return nil, errors.New("invalid_token")
	}

	userIDStr, _ := claims["sub"].(string)
	userID, _ := uuid.Parse(userIDStr)
	orgIDStr, _ := claims["org_id"].(string)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Verify TOTP code
	if !verifyTOTP(user.MFASecret, req.Code) {
		return nil, ErrInvalidMFACode
	}

	// Generate full tokens
	return s.generateLoginTokens(ctx, user, orgIDStr, ip, userAgent, req.RememberDevice)
}

// Refresh generates new access + refresh tokens with rotation-based replay protection.
func (s *AuthService) Refresh(ctx context.Context, refreshToken, ip, ua string) (*models.AuthTokensResponse, error) {
	// 1. Parse and verify refresh token
	claims, err := s.jwtManager.ExtractClaims(refreshToken)
	if err != nil {
		return nil, errors.New("invalid_token")
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return nil, errors.New("invalid_token")
	}

	userIDStr, _ := claims["sub"].(string)
	orgIDStr, _ := claims["org_id"].(string)
	jti, _ := claims["jti"].(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid_token")
	}

	// 2. Check for replay attack via rotation detection
	// If this refresh token's JTI has already been consumed, it's either a legitimate
	// concurrent mobile refresh (within a short grace window) or a stolen token being replayed.
	// Either way, we revoke ALL sessions and return an error.
	if jti != "" && s.redis != nil {
		used, err := s.redis.IsRefreshTokenUsed(ctx, jti)
		if err != nil {
			s.log.Debug("refresh token check failed", zap.Error(err))
		} else if used {
			// Token replay detected — revoke all sessions for security
			s.sessionRepo.RevokeAllTokensEverywhere(ctx, userID)
			s.log.Warn("refresh_token_replay_detected",
				zap.String("user_id", userIDStr),
				zap.String("jti", jti),
			)
			return nil, errors.New("token_revoked")
		}
	}

	// 3. Mark this refresh token as consumed (rotation)
	if jti != "" && s.redis != nil {
		s.redis.MarkRefreshTokenUsed(ctx, jti, s.config.JWTRefreshTTL)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check lockout
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, ErrAccountLocked
	}

	// 4. Generate new tokens (new JTI, so the old refresh token cannot be reused)
	return s.generateTokens(ctx, user, orgIDStr, ip, ua, "refresh")
}

// ── Internal helpers ──

func (s *AuthService) handleFailedLogin(ctx context.Context, user *models.User, ip string) {
	s.userRepo.IncrementFailedLogins(ctx, user.ID)

	newAttempts := user.FailedLogins + 1
	if newAttempts >= s.config.MaxLoginAttempts {
		lockedUntil := time.Now().Add(s.config.LockoutDuration)
		s.userRepo.LockAccount(ctx, user.ID, lockedUntil)
		s.redis.RecordFailedLogin(ctx, user.ID.String(), newAttempts, lockedUntil, s.config.LockoutDuration)
		s.log.Warn("account_locked",
			zap.String("user_id", user.ID.String()),
			zap.Int("failed_attempts", newAttempts),
			zap.Time("locked_until", lockedUntil),
		)

		// Send lockout notification email (async, non-blocking)
		if s.emailSender != nil && user.Email != "" {
			go func(email string) {
				emailCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if err := s.emailSender.SendAccountLocked(emailCtx, email); err != nil {
					s.log.Error("failed to send lockout email", zap.Error(err), zap.String("email", email))
				}
			}(user.Email)
		}
	}
}

func (s *AuthService) generateLoginTokens(ctx context.Context, user *models.User, orgID, ip, userAgent string, rememberDevice bool) (*models.AuthTokensResponse, error) {
	// Get org info
	orgs, err := s.orgRepo.GetUserOrganizations(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	var orgName string
	var userRole string
	for _, o := range orgs {
		if o.ID.String() == orgID {
			orgName = o.Name
			userRole = o.Role
			break
		}
	}

	// Compute permissions
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		s.log.Warn("invalid_org_id_in_login", zap.String("org_id", orgID))
		return nil, errors.New("invalid_org_id")
	}
	mask, err := s.roleRepo.ComputeEffectivePermissionMask(ctx, user.ID, orgUUID)
	if err != nil {
		s.log.Error("failed to compute permissions", zap.Error(err))
		mask = 0
	}

	// Cache permissions
	s.redis.CacheUserPermissions(ctx, user.ID.String(), orgID, mask, 5*time.Minute)

	// Update last login
	s.userRepo.UpdateLastLogin(ctx, user.ID, ip)
	s.userRepo.ResetFailedLogins(ctx, user.ID)

	// Generate tokens
	tokens, err := s.generateTokens(ctx, user, orgID, ip, userAgent, "login")
	if err != nil {
		return nil, err
	}

	// Set user info
	tokens.User.Role = userRole
	tokens.User.OrgName = orgName
	// Build permissions list
	tokens.User.Permissions = s.resolvePermissions(ctx, mask)

	return tokens, nil
}

func (s *AuthService) generateTokens(ctx context.Context, user *models.User, orgID, ip, ua, eventType string) (*models.AuthTokensResponse, error) {
	orgUUID, _ := uuid.Parse(orgID)
	jti := uuid.New().String()

	// Get permissions
	mask, err := s.roleRepo.ComputeEffectivePermissionMask(ctx, user.ID, orgUUID)
	if err != nil {
		mask = user.PermissionMask
	}

	// Get role
	roles, err := s.roleRepo.GetUserRoles(ctx, user.ID, orgUUID)
	role := "viewer"
	if len(roles) > 0 {
		role = roles[0].Name
	}

	// Access token claims
	accessClaims := jwt.MapClaims{
		"sub":                user.ID.String(),
		"user_id":            user.ID.String(),
		"org_id":             orgID,
		"role":               role,
		"permissions_mask":   mask,
		"type":               "access",
		"jti":                jti,
		"iss":                "flyfleet-auth-service",
		"aud":                "flyfleet-api",
		"device_fingerprint": deviceFingerprint(ip),
	}

	accessToken, err := s.jwtManager.GenerateToken(accessClaims, s.config.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	// Refresh token
	refreshJTI := uuid.New().String()
	refreshClaims := jwt.MapClaims{
		"sub":     user.ID.String(),
		"user_id": user.ID.String(),
		"org_id":  orgID,
		"type":    "refresh",
		"jti":     refreshJTI,
	}

	refreshToken, err := s.jwtManager.GenerateToken(refreshClaims, s.config.JWTRefreshTTL)
	if err != nil {
		return nil, err
	}

	// Store session in Redis
	sessionData := redis2.SessionData{
		UserID:     user.ID,
		OrgID:      orgUUID,
		Role:       role,
		Device:     ua,
		IP:         ip,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
	s.redis.SetSession(ctx, jti, sessionData, s.config.JWTAccessTTL)

	// Store in DB
	s.sessionRepo.StoreToken(ctx, &models.AuthToken{
		UserID:     user.ID,
		OrgID:      orgUUID,
		TokenHash:  accessToken,
		TokenType:  "access",
		DeviceInfo: ua,
		IPAddress:  ip,
		UserAgent:  ua,
		ExpiresAt:  time.Now().Add(s.config.JWTAccessTTL),
		LastActiveAt: time.Now(),
	})

	return &models.AuthTokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.config.JWTAccessTTL.Seconds()),
		User: models.UserInfo{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  role,
			OrgID: orgUUID,
		},
	}, nil
}

func (s *AuthService) generateActionToken(userID uuid.UUID, orgID, tokenType string) (string, error) {
	claims := jwt.MapClaims{
		"sub":     userID.String(),
		"org_id":  orgID,
		"type":    tokenType,
		"jti":     uuid.New().String(),
		"exp":     time.Now().Add(s.config.JWTActionTTL).Unix(),
		"iat":     time.Now().Unix(),
	}
	return s.jwtManager.GenerateToken(claims, s.config.JWTActionTTL)
}

// ForgotPassword initiates a password reset flow.
// Anti-enumeration: always returns nil even if user doesn't exist.
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.log.Error("forgot_password lookup failed", zap.Error(err), zap.String("email", email))
		return nil // Anti-enumeration: don't leak errors
	}

	if user == nil {
		// User doesn't exist — return success anyway (anti-enumeration)
		return nil
	}

	// Generate password reset action token
	token, err := s.generateActionToken(user.ID, "", "password_reset")
	if err != nil {
		s.log.Error("failed to generate reset token", zap.Error(err), zap.String("user_id", user.ID.String()))
		return nil
	}

	// Mark token as single-use in Redis
	claims, _ := s.jwtManager.ExtractClaims(token)
	jti, _ := claims["jti"].(string)
	if jti != "" {
		s.redis.MarkActionTokenUsed(ctx, jti, s.config.JWTActionTTL)
	}

	// Send reset email asynchronously
	if s.emailSender != nil {
		go func(emailAddr, tkn string) {
			emailCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := s.emailSender.SendPasswordReset(emailCtx, emailAddr, tkn); err != nil {
				s.log.Error("failed to send password reset email",
					zap.Error(err), zap.String("email", emailAddr))
			}
		}(user.Email, token)
	}

	s.log.Info("password_reset_initiated", zap.String("email", email))
	return nil
}

// verifyTOTP validates a TOTP code against the user's secret using pquerna/otp
func verifyTOTP(secret, code string) bool {
	if secret == "" || code == "" || len(code) != 6 {
		return false
	}
	return VerifyTOTPCode(secret, code)
}

// GetUserOrganizations returns all organizations a user belongs to
func (s *AuthService) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]models.OrgInfo, error) {
	return s.orgRepo.GetUserOrganizations(ctx, userID)
}

// Register creates a new user account with the given credentials
func (s *AuthService) Register(ctx context.Context, email, password, name string, orgID uuid.UUID) (*models.AuthTokensResponse, error) {
	// Check if user already exists
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email_already_exists")
	}

	// Validate password
	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.config.BcryptCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		Name:         name,
		PasswordHash: string(hash),
		Status:       models.UserStatusActive,
		EmailVerified: false,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// If orgID provided, create user-org association
	// Note: This would typically be done via a separate user_orgs repo method
	// For now, generate tokens and let the user join org later

	// Generate tokens
	return s.generateTokens(ctx, user, orgID.String(), "", "", "register")
}

// Logout revokes the current session
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, tokenJTI string) error {
	// Remove session from Redis
	s.redis.DeleteSession(ctx, tokenJTI)
	s.log.Info("user_logout", zap.String("user_id", userID.String()))
	return nil
}

// VerifyEmail marks a user's email as verified given a valid JWT verification token.
// The token must be an "email_verification" action token generated by the system.
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	// 1. Parse and validate the JWT token
	parsedToken, err := s.jwtManager.VerifyToken(token)
	if err != nil {
		return errors.New("invalid_token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid_token")
	}

	// 2. Check token type is email_verification
	tokenType, _ := claims["type"].(string)
	if tokenType != "email_verification" {
		return errors.New("invalid_token")
	}

	// 3. Atomically consume the single-use token via Redis SET NX
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return errors.New("invalid_token")
	}
	consumed, err := s.redis.AtomicConsumeActionToken(ctx, jti, s.config.JWTActionTTL)
	if err != nil {
		s.log.Error("failed to atomically consume email verification token", zap.Error(err))
		return errors.New("invalid_token")
	}
	if !consumed {
		return errors.New("token_already_used")
	}

	// 4. Extract user_id from claims
	userIDStr, _ := claims["sub"].(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.New("invalid_token")
	}

	// 5. Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 6. Mark email as verified
	if err := s.userRepo.MarkEmailVerified(ctx, userID); err != nil {
		return err
	}

	// 7. Token already consumed atomically in step 3 — no need to mark again

	s.log.Info("email_verified", zap.String("user_id", userID.String()), zap.String("email", user.Email))
	return nil
}

// ResetPasswordWithToken completes a password reset using a valid reset token.
// Full flow: parse JWT → validate policy → check history → hash → save → revoke sessions → mark token used.
func (s *AuthService) ResetPasswordWithToken(ctx context.Context, token, newPassword string) error {
	// 1. Parse and validate the JWT token
	parsedToken, err := s.jwtManager.VerifyToken(token)
	if err != nil {
		return errors.New("invalid_token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid_token")
	}

	// 2. Check token type is password_reset
	tokenType, _ := claims["type"].(string)
	if tokenType != "password_reset" {
		return errors.New("invalid_token")
	}

	// 2b. Extract jti for atomic consume check (must be before step 4)
	jti, _ := claims["jti"].(string)

	// 3. Extract user_id from claims
	userIDStr, _ := claims["sub"].(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.New("invalid_token")
	}

	// 4. Atomically consume the single-use token via Redis SET NX
	// This is atomic — no TOCTOU race between check and mark
	if jti == "" {
		return errors.New("invalid_token")
	}
	consumed, err := s.redis.AtomicConsumeActionToken(ctx, jti, s.config.JWTActionTTL)
	if err != nil {
		s.log.Error("failed to atomically consume reset token", zap.Error(err))
		return errors.New("invalid_token")
	}
	if !consumed {
		return errors.New("token_already_used")
	}

	// 5. Get user by ID
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 6. Validate password against policy
	if s.passwordSvc != nil {
		if err := s.passwordSvc.Validate(ctx, userID, newPassword); err != nil {
			return err
		}
	}

	// 7. Save old hash to password history
	if err := s.passwordRepo.Add(ctx, userID, user.PasswordHash); err != nil {
		s.log.Error("failed to save password history", zap.Error(err))
	}

	// 8. Hash new password and update user
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.config.BcryptCost)
	if err != nil {
		return err
	}
	if err := s.userRepo.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return err
	}

	// 9. Revoke ALL sessions everywhere
	if err := s.sessionRepo.RevokeAllTokensEverywhere(ctx, userID); err != nil {
		s.log.Error("failed to revoke sessions after password reset", zap.Error(err))
	}

	// 10. Token already consumed atomically in step 4 — no need to mark again

	s.log.Info("password_reset_completed", zap.String("user_id", userID.String()))
	return nil
}

// SwitchOrg validates the user belongs to the target org and issues new tokens
func (s *AuthService) SwitchOrg(ctx context.Context, userID uuid.UUID, targetOrgID uuid.UUID, ip, ua string) (*models.OrgSwitchResponse, error) {
	// Verify user has access to target org
	orgs, err := s.orgRepo.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, err
	}

	var found bool
	var orgInfo models.OrgInfo
	var userRole string
	for _, o := range orgs {
		if o.ID == targetOrgID {
			found = true
			orgInfo = o
			userRole = o.Role
			break
		}
	}
	if !found {
		return nil, errors.New("org_switch_failed")
	}

	// Compute permissions for new org
	mask, err := s.roleRepo.ComputeEffectivePermissionMask(ctx, userID, targetOrgID)
	if err != nil {
		s.log.Error("failed to compute permissions for org switch", zap.Error(err))
		mask = 0
	}

	// Cache new permissions
	s.redis.CacheUserPermissions(ctx, userID.String(), targetOrgID.String(), mask, 5*time.Minute)

	// Generate new JWT with new org context
	jti := uuid.New().String()
	accessClaims := jwt.MapClaims{
		"sub":              userID.String(),
		"user_id":          userID.String(),
		"org_id":           targetOrgID.String(),
		"role":             userRole,
		"permissions_mask": mask,
		"type":             "access",
		"jti":              jti,
		"iss":              "flyfleet-auth-service",
		"aud":              "flyfleet-api",
		"device_fingerprint": deviceFingerprint(ip),
	}

	accessToken, err := s.jwtManager.GenerateToken(accessClaims, s.config.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	refreshJTI := uuid.New().String()
	refreshClaims := jwt.MapClaims{
		"sub":     userID.String(),
		"user_id": userID.String(),
		"org_id":  targetOrgID.String(),
		"type":    "refresh",
		"jti":     refreshJTI,
	}
	refreshToken, err := s.jwtManager.GenerateToken(refreshClaims, s.config.JWTRefreshTTL)
	if err != nil {
		return nil, err
	}

	// Update session cache
	sessionData := redis2.SessionData{
		UserID:     userID,
		OrgID:      targetOrgID,
		Role:       userRole,
		Device:     ua,
		IP:         ip,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
	s.redis.SetSession(ctx, jti, sessionData, s.config.JWTAccessTTL)

	// Log the switch
	s.orgRepo.LogOrgSwitch(ctx, userID, uuid.Nil, targetOrgID, ip, ua)

	s.log.Info("org_switched",
		zap.String("user_id", userID.String()),
		zap.String("to_org_id", targetOrgID.String()),
	)

	return &models.OrgSwitchResponse{
		Success:      true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.config.JWTAccessTTL.Seconds()),
		Organization: orgInfo,
	}, nil
}

// resolvePermissions returns the list of resource:action strings for a given permission mask
func (s *AuthService) resolvePermissions(ctx context.Context, mask int64) []string {
	perms, err := s.roleRepo.GetAllPermissions(ctx)
	if err != nil {
		s.log.Error("failed to resolve permissions", zap.Error(err))
		return []string{}
	}
	var result []string
	for _, p := range perms {
		if mask&(1<<p.BitPosition) != 0 {
			result = append(result, p.Resource+":"+p.Action)
		}
	}
	return result
}

func deviceFingerprint(ip string) string {
	h := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(h[:])
}

func extractIP(ip string) string {
	if host, _, err := net.SplitHostPort(ip); err == nil {
		return host
	}
	return ip
}
