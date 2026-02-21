package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/flyfleet/auth-service/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo                 repository.AuthRepository
	jwtSecret            string
	logger               *zap.Logger
	fleetCoreServiceURL  string
	alertServiceURL      string
	httpClient           *http.Client
	ipAttemptTracker     map[string][]time.Time // Track login attempts by IP
	passwordResetTracker map[string][]time.Time // Track password reset attempts
	ipTrackerMutex       sync.RWMutex
}

const (
	actionPurposeVerifyEmail   = "verify_email"
	actionPurposePasswordReset = "password_reset"
)

func NewAuthService(repo repository.AuthRepository, jwtSecret, fleetCoreURL, alertURL string, logger *zap.Logger) *AuthService {
	authService := &AuthService{
		repo:                 repo,
		jwtSecret:            jwtSecret,
		logger:               logger,
		fleetCoreServiceURL:  fleetCoreURL,
		alertServiceURL:      alertURL,
		httpClient:           &http.Client{Timeout: 10 * time.Second},
		ipAttemptTracker:     make(map[string][]time.Time),
		passwordResetTracker: make(map[string][]time.Time),
	}

	// Start background cleanup for expired tokens
	go authService.startTokenCleanup()

	return authService
}

// startTokenCleanup starts a background process to clean up expired tokens
func (s *AuthService) startTokenCleanup() {
	ticker := time.NewTicker(30 * time.Minute) // Run cleanup every 30 minutes
	defer ticker.Stop()

	for range ticker.C {
		if err := s.cleanupExpiredTokens(); err != nil {
			s.logger.Error("Failed to cleanup expired tokens", zap.Error(err))
		}
	}
}

// cleanupExpiredTokens removes expired tokens from the database
func (s *AuthService) cleanupExpiredTokens() error {
	s.logger.Info("Running expired token cleanup")
	return s.repo.CleanupExpiredTokens(context.Background())
}

// checkIPRateLimit checks if an IP has exceeded the login attempt limit
func (s *AuthService) checkIPRateLimit(ip string) bool {
	s.ipTrackerMutex.RLock()
	attempts, exists := s.ipAttemptTracker[ip]
	s.ipTrackerMutex.RUnlock()

	if !exists {
		return true // IP not tracked, allow login
	}

	// Count attempts in the last 15 minutes
	now := time.Now()
	cutoff := now.Add(-15 * time.Minute)
	count := 0

	for _, attemptTime := range attempts {
		if attemptTime.After(cutoff) {
			count++
		}
	}

	// Allow maximum 20 login attempts per IP per 15 minutes
	return count < 20
}

// recordIPAttempt records a login attempt for an IP
func (s *AuthService) recordIPAttempt(ip string) {
	s.ipTrackerMutex.Lock()
	defer s.ipTrackerMutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-15 * time.Minute)

	// Clean up old attempts
	var cleanedAttempts []time.Time
	for _, attemptTime := range s.ipAttemptTracker[ip] {
		if attemptTime.After(cutoff) {
			cleanedAttempts = append(cleanedAttempts, attemptTime)
		}
	}

	// Add the new attempt
	cleanedAttempts = append(cleanedAttempts, now)
	s.ipAttemptTracker[ip] = cleanedAttempts
}

func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	return s.repo.DeleteToken(ctx, accessToken)
}

func (s *AuthService) GetSecurityEvents(ctx context.Context, userID uuid.UUID) ([]models.SecurityEvent, error) {
	return s.repo.GetSecurityEvents(ctx, userID)
}

func (s *AuthService) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]models.AuthToken, error) {
	tokens, err := s.repo.GetTokensByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Filter out expired tokens
	var activeTokens []models.AuthToken
	for _, token := range tokens {
		if token.ExpiresAt.After(time.Now()) {
			activeTokens = append(activeTokens, token)
		}
	}

	return activeTokens, nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteTokensByUserID(ctx, userID)
}

func (s *AuthService) ListAuditLogs(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.AuditLog, int, error) {
	return s.repo.ListAuditLogs(ctx, orgID, page, size, filters)
}

func (s *AuthService) GetAuditLogByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) {
	return s.repo.GetAuditLogByID(ctx, id)
}

func (s *AuthService) GetAuditSettings(ctx context.Context, orgID uuid.UUID) (*models.AuditSettings, error) {
	return s.repo.GetAuditSettings(ctx, orgID)
}

func (s *AuthService) UpdateAuditSettings(ctx context.Context, settings *models.AuditSettings) error {
	return s.repo.UpdateAuditSettings(ctx, settings)
}

func (s *AuthService) Login(ctx context.Context, email, password, ip, ua string) (*models.AuthResponse, error) {
	// Validate and sanitize the IP address
	sanitizedIP, err := validateAndSanitizeIP(ip)
	if err != nil {
		s.logger.Warn("Invalid IP address provided for login", zap.String("ip", ip), zap.Error(err))
		// Use a default value or deny login if IP is invalid
		sanitizedIP = "0.0.0.0" // Default fallback
	}

	// Check IP rate limit first to prevent DoS attacks
	if !s.checkIPRateLimit(sanitizedIP) {
		s.logger.Warn("IP rate limit exceeded", zap.String("ip", sanitizedIP))
		s.logLoginActivity(ctx, nil, "FAILED", "high", sanitizedIP, ua)
		return nil, errors.New("too many login attempts from this IP address")
	}

	// Record the IP attempt
	s.recordIPAttempt(sanitizedIP)

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Warn("Login attempt for non-existent user or DB error", zap.String("email", email), zap.Error(err))
		// For non-existent users, we still log the attempt to detect enumeration attempts
		// but we don't increment failed logins to prevent DoS attacks against non-existent accounts
		s.logLoginActivity(ctx, nil, "FAILED", "low", sanitizedIP, ua)

		// To prevent user enumeration and timing attacks, we'll simulate the delay that would happen with a real user
		// This makes timing attacks harder by making the response time consistent
		bcrypt.CompareHashAndPassword([]byte("$2a$10$dummyhash"), []byte(password))

		return nil, errors.New("invalid credentials")
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		s.logLoginActivity(ctx, user, "FAILED", "high", sanitizedIP, ua)
		return nil, errors.New("account is temporarily locked due to multiple failed attempts")
	}

	// Check if the user has reached the CAPTCHA threshold
	if user.FailedLogins >= 3 {
		s.logLoginActivity(ctx, user, "CHALLENGED", "medium", sanitizedIP, ua)
		return nil, errors.New("CAPTCHA_REQUIRED")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// Only increment failed logins for existing users to prevent DoS attacks
		count, _ := s.repo.IncrementFailedLogins(ctx, user.ID)
		s.logLoginActivity(ctx, user, "FAILED", "low", sanitizedIP, ua)

		// Lock account if threshold reached (e.g., 5 attempts)
		if count >= 5 {
			s.repo.LockAccount(ctx, user.ID, 15*time.Minute)
			// Create Threat Alert
			s.repo.CreateThreatAlert(ctx, &models.ThreatAlert{
				OrgID:       uuid.Nil, // Will be enriched if possible
				Type:        "BRUTE_FORCE",
				Severity:    "critical",
				IP:          sanitizedIP,
				UserID:      &user.ID,
				Description: fmt.Sprintf("Multiple failed login attempts for user %s", user.Email),
				Status:      "open",
			})
		}

		return nil, errors.New("invalid credentials")
	}

	// Reset failed logins on success
	s.repo.ResetFailedLogins(ctx, user.ID)
	s.logLoginActivity(ctx, user, "SUCCESS", "low", sanitizedIP, ua)

	if user.MFAEnabled {
		return &models.AuthResponse{
			MFAWait: true,
			User:    *user,
		}, nil
	}

	// Fetch user organizations to set default context
	orgs, err := s.repo.GetOrganizationsByUserID(ctx, user.ID)
	if err == nil && len(orgs) > 0 {
		return s.issueTokensWithOrg(ctx, user, orgs[0].ID, sanitizedIP, ua)
	}

	return s.issueTokens(ctx, user, sanitizedIP, ua)
}

// UpdateTokenLastActive updates the last active time for a token
func (s *AuthService) UpdateTokenLastActive(ctx context.Context, accessToken string) error {
	return s.repo.UpdateTokenLastActive(ctx, accessToken)
}

func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *AuthService) GetSecurityTelemetry(ctx context.Context) (map[string]any, error) {
	return s.repo.GetSecurityTelemetry(ctx)
}

// Security Configuration Methods

func (s *AuthService) Register(ctx context.Context, email, password, name, ip, ua string) (*models.User, error) {
	// Validate and sanitize the IP address
	sanitizedIP, err := validateAndSanitizeIP(ip)
	if err != nil {
		s.logger.Warn("Invalid IP address provided for registration", zap.String("ip", ip), zap.Error(err))
		sanitizedIP = "0.0.0.0" // Default fallback
	}

	// Use a higher cost for bcrypt to improve security (DefaultCost is typically 10, we'll use 12)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         &name,
		Status:       "active",
		Verified:     false, // Requires email verification
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    user.ID,
		EventType: "USER_REGISTERED",
		IPAddress: sanitizedIP,
		UserAgent: ua,
	})

	verificationToken, tokenErr := s.generateActionToken(user.ID, actionPurposeVerifyEmail, 24*time.Hour)
	if tokenErr != nil {
		s.logger.Warn("Failed to generate verification token", zap.Error(tokenErr), zap.String("user_id", user.ID.String()))
	} else {
		s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
			UserID:    user.ID,
			EventType: "EMAIL_VERIFICATION_REQUESTED",
			IPAddress: sanitizedIP,
			UserAgent: ua,
			Metadata:  fmt.Sprintf(`{"delivery":"pending","token":"%s","expires_in_hours":24}`, verificationToken),
		})
	}

	return user, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	userID, err := s.parseActionToken(token, actionPurposeVerifyEmail)
	if err != nil {
		legacyID, parseErr := uuid.Parse(token)
		if parseErr != nil {
			return errors.New("invalid verification token")
		}
		userID = legacyID
	}

	if err := s.repo.UpdateUserVerified(ctx, userID, true); err != nil {
		return err
	}

	s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    userID,
		EventType: "EMAIL_VERIFIED",
	})
	return nil
}

func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return nil // Don't reveal if user exists
	}
	if user.Verified {
		return errors.New("email already verified")
	}

	verificationToken, tokenErr := s.generateActionToken(user.ID, actionPurposeVerifyEmail, 24*time.Hour)
	if tokenErr != nil {
		return tokenErr
	}

	s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    user.ID,
		EventType: "EMAIL_VERIFICATION_RESENT",
		Metadata:  fmt.Sprintf(`{"delivery":"pending","token":"%s","expires_in_hours":24}`, verificationToken),
	})

	return nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return nil // Don't reveal
	}

	resetToken, tokenErr := s.generateActionToken(user.ID, actionPurposePasswordReset, 30*time.Minute)
	if tokenErr != nil {
		return tokenErr
	}

	s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    user.ID,
		EventType: "PASSWORD_RESET_REQUESTED",
		Metadata:  fmt.Sprintf(`{"delivery":"pending","token":"%s","expires_in_minutes":30}`, resetToken),
	})
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	userID, err := s.parseActionToken(token, actionPurposePasswordReset)
	if err != nil {
		legacyID, parseErr := uuid.Parse(token)
		if parseErr != nil {
			return errors.New("invalid reset token")
		}
		userID = legacyID
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = string(hashedPassword)

	if err := s.repo.UpdateUserPassword(ctx, user); err != nil {
		return err
	}

	s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    user.ID,
		EventType: "PASSWORD_RESET_SUCCESS",
	})
	return nil
}

// generateSecureCode generates a cryptographically secure random code
func generateSecureCode(length int) (string, error) {
	const charset = "0123456789"
	result := make([]byte, length)
	charsetLen := byte(len(charset))

	// Use crypto/rand for secure random generation
	_, err := rand.Read(result)
	if err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		result[i] = charset[result[i]%charsetLen]
	}

	return string(result), nil
}

// validateAndSanitizeIP validates and sanitizes an IP address
func validateAndSanitizeIP(ip string) (string, error) {
	// Remove port if present
	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		// If SplitHostPort fails, it might be because there's no port
		// In that case, use the original IP
		host = ip
	}

	// Parse the IP to validate it
	parsedIP := net.ParseIP(host)
	if parsedIP == nil {
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}

	// Return the normalized IP address
	return parsedIP.String(), nil
}

func (s *AuthService) SendMFA(ctx context.Context, userID uuid.UUID, method string) error {
	code, err := generateSecureCode(6) // 6-digit code
	if err != nil {
		return fmt.Errorf("failed to generate secure MFA code: %w", err)
	}

	mfa := &models.MFACode{
		ID:        uuid.New(),
		UserID:    userID,
		Code:      code, // This will be hashed in the repository layer
		Method:    method,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}
	return s.repo.SaveMFACode(ctx, mfa)
}

func (s *AuthService) VerifyMFA(ctx context.Context, userID uuid.UUID, code, ip, ua string) (*models.AuthResponse, error) {
	// Validate and sanitize the IP address
	sanitizedIP, err := validateAndSanitizeIP(ip)
	if err != nil {
		s.logger.Warn("Invalid IP address provided for MFA verification", zap.String("ip", ip), zap.Error(err))
		sanitizedIP = "0.0.0.0" // Default fallback
	}

	mfa, err := s.repo.GetMFACode(ctx, userID, code)
	if err != nil {
		return nil, err
	}
	if mfa == nil {
		return nil, errors.New("invalid or expired OTP")
	}

	// Delete the MFA code after successful verification to prevent replay attacks
	if err := s.repo.DeleteMFACode(ctx, mfa.ID); err != nil {
		s.logger.Error("Failed to delete MFA code after verification", zap.Error(err))
		// Don't return error to user, just log it
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user, sanitizedIP, ua)
}

func (s *AuthService) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]models.Organization, error) {
	return s.repo.GetOrganizationsByUserID(ctx, userID)
}

func (s *AuthService) SelectOrganization(ctx context.Context, userID, orgID uuid.UUID, ip, ua string) (*models.AuthResponse, error) {
	// Validate and sanitize the IP address
	sanitizedIP, err := validateAndSanitizeIP(ip)
	if err != nil {
		s.logger.Warn("Invalid IP address provided for organization selection", zap.String("ip", ip), zap.Error(err))
		sanitizedIP = "0.0.0.0" // Default fallback
	}

	orgs, err := s.repo.GetOrganizationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Validate that the user belongs to the requested organization
	valid := false
	for _, o := range orgs {
		if o.ID == orgID {
			valid = true
			break
		}
	}
	if !valid {
		s.logger.Warn("Unauthorized organization access attempt", zap.String("user_id", userID.String()), zap.String("org_id", orgID.String()))
		return nil, errors.New("unauthorized organization access")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.issueTokensWithOrg(ctx, user, orgID, sanitizedIP, ua)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string, ip, ua string) (*models.AuthResponse, error) {
	// Validate and sanitize the IP address
	sanitizedIP, err := validateAndSanitizeIP(ip)
	if err != nil {
		s.logger.Warn("Invalid IP address provided for token refresh", zap.String("ip", ip), zap.Error(err))
		sanitizedIP = "0.0.0.0" // Default fallback
	}

	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userIDStr, _ := claims["user_id"].(string)
	userID, _ := uuid.Parse(userIDStr)

	// Verify the refresh token exists in the database (for rotation)
	storedToken, err := s.repo.GetTokenByRefresh(ctx, refreshToken)
	if err != nil || storedToken == nil {
		// Invalid or used refresh token - possible token theft
		s.logger.Warn("Invalid refresh token used", zap.String("user_id", userIDStr), zap.String("ip", sanitizedIP))
		// For security, invalidate all tokens for this user
		s.repo.DeleteTokensByUserID(ctx, userID)
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Delete the used refresh token to prevent reuse (rotation)
	if err := s.repo.DeleteToken(ctx, storedToken.AccessToken); err != nil {
		s.logger.Error("Failed to delete used refresh token", zap.Error(err))
	}

	return s.issueTokens(ctx, user, sanitizedIP, ua)
}

func (s *AuthService) UpdateTheme(ctx context.Context, userID uuid.UUID, theme, ip, ua string) error {
	// Validate and sanitize the IP address
	sanitizedIP, err := validateAndSanitizeIP(ip)
	if err != nil {
		s.logger.Warn("Invalid IP address provided for theme update", zap.String("ip", ip), zap.Error(err))
		sanitizedIP = "0.0.0.0" // Default fallback
	}

	// 1. Validation
	validThemes := map[string]bool{"light": true, "dark": true, "system": true}
	if !validThemes[theme] {
		return errors.New("invalid theme: must be light, dark, or system")
	}

	// 2. Update DB
	if err := s.repo.UpdateTheme(ctx, userID, theme); err != nil {
		return err
	}

	// 3. Audit Log
	auditLog := &models.AuditLog{
		OrgID:     userID, // Placeholder
		UserID:    userID,
		Module:    "PROFILE",
		Action:    "USER_THEME_UPDATE",
		Resource:  "user_profile",
		EntityID:  userID.String(),
		Severity:  "info",
		Status:    "success",
		Details:   fmt.Sprintf(`{"new_theme": "%s"}`, theme),
		IPAddress: sanitizedIP,
		UserAgent: ua,
	}
	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		s.logger.Error("Failed to create audit log for theme update", zap.Error(err))
	}

	return nil
}

// UpdateUserPassword updates only the password for a user
func (s *AuthService) UpdateUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	// Hash the new password with increased cost
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Get the user to update only the password field
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)

	return s.repo.UpdateUserPassword(ctx, user)
}

// IsPasswordResetAllowed checks if a password reset is allowed based on rate limiting
func (s *AuthService) IsPasswordResetAllowed(ctx context.Context, email, ip string) bool {
	// Check both email-based and IP-based rate limits
	emailKey := "pwdreset:" + email
	ipKey := "pwdreset:ip:" + ip

	now := time.Now()
	emailCutoff := now.Add(-15 * time.Minute)
	ipCutoff := now.Add(-15 * time.Minute)

	// Check email-based attempts (max 3 per 15 minutes)
	emailAttempts := s.getPasswordResetAttempts(emailKey, emailCutoff)
	if emailAttempts >= 3 {
		return false
	}

	// Check IP-based attempts (max 5 per 15 minutes)
	ipAttempts := s.getPasswordResetAttempts(ipKey, ipCutoff)
	if ipAttempts >= 5 {
		return false
	}

	// Record the attempt
	s.recordPasswordResetAttempt(emailKey)
	s.recordPasswordResetAttempt(ipKey)

	return true
}

// getPasswordResetAttempts gets the number of password reset attempts within the time window
func (s *AuthService) getPasswordResetAttempts(key string, cutoff time.Time) int {
	s.ipTrackerMutex.RLock()
	defer s.ipTrackerMutex.RUnlock()

	attempts, exists := s.passwordResetTracker[key]
	if !exists {
		return 0
	}

	count := 0
	for _, attemptTime := range attempts {
		if attemptTime.After(cutoff) {
			count++
		}
	}
	return count
}

// recordPasswordResetAttempt records a password reset attempt
func (s *AuthService) recordPasswordResetAttempt(key string) {
	s.ipTrackerMutex.Lock()
	defer s.ipTrackerMutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-15 * time.Minute)

	// Clean up old attempts
	var cleanedAttempts []time.Time
	for _, attemptTime := range s.passwordResetTracker[key] {
		if attemptTime.After(cutoff) {
			cleanedAttempts = append(cleanedAttempts, attemptTime)
		}
	}

	// Add the new attempt
	cleanedAttempts = append(cleanedAttempts, now)
	s.passwordResetTracker[key] = cleanedAttempts
}

// Admin Operations
func (s *AuthService) CreateOrganization(ctx context.Context, name, domain, theme string, creatorID uuid.UUID) (*models.Organization, error) {
	org := &models.Organization{
		ID:           uuid.New(),
		Name:         name,
		Domain:       domain,
		DefaultTheme: theme,
		Status:       "active",
	}

	if err := s.repo.CreateOrganization(ctx, org); err != nil {
		return nil, err
	}

	// Audit Log
	_ = s.repo.CreateAuditLog(ctx, &models.AuditLog{
		OrgID:    org.ID,
		UserID:   creatorID,
		Action:   "ORG_CREATE",
		Resource: "organization",
		Severity: "info",
		Status:   "success",
		Details:  fmt.Sprintf(`{"name": "%s", "domain": "%s"}`, name, domain),
	})

	return org, nil
}

func (s *AuthService) UpdateOrganization(ctx context.Context, orgID uuid.UUID, name, domain, theme, status string, updaterID uuid.UUID) (*models.Organization, error) {
	org, err := s.repo.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	org.Name = name
	org.Domain = domain
	org.DefaultTheme = theme
	if status != "" {
		org.Status = status
	}

	if err := s.repo.UpdateOrganization(ctx, org); err != nil {
		return nil, err
	}

	// Audit Log
	_ = s.repo.CreateAuditLog(ctx, &models.AuditLog{
		OrgID:    org.ID,
		UserID:   updaterID,
		Action:   "ORG_UPDATE",
		Resource: "organization",
		Severity: "info",
		Status:   "success",
		Details:  fmt.Sprintf(`{"name": "%s", "status": "%s"}`, name, status),
	})

	return org, nil
}

func (s *AuthService) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	return s.repo.GetOrganizationByID(ctx, id)
}

func (s *AuthService) ListOrganizations(ctx context.Context, page, size int) ([]models.Organization, int, error) {
	return s.repo.ListOrganizations(ctx, page, size)
}

func (s *AuthService) ListUsers(ctx context.Context, filter models.UserFilter) (*models.UserListResponse, error) {
	// 1. Validation and Defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}

	// 2. Repository Call
	users, total, err := s.repo.ListUsers(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to list users", zap.Error(err), zap.Any("filter", filter))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// 3. Structure Response
	totalPages := 0
	if total > 0 {
		totalPages = (total + filter.Limit - 1) / filter.Limit
	}

	response := &models.UserListResponse{
		Data: users,
		Meta: models.UserMeta{
			Total:      total,
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalPages: totalPages,
			HasNext:    filter.Page < totalPages,
			HasPrev:    filter.Page > 1,
		},
	}

	return response, nil
}

func (s *AuthService) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	return s.repo.GetRoleByID(ctx, id)
}

func (s *AuthService) SoftDeleteOrganization(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.SoftDeleteOrganization(ctx, id); err != nil {
		return err
	}

	actorID, _ := ctx.Value("user_id").(uuid.UUID)
	s.repo.CreateAuditLog(ctx, &models.AuditLog{
		OrgID:     id,
		UserID:    actorID,
		Module:    "organizations",
		Action:    "soft_delete",
		Resource:  "organization",
		EntityID:  id.String(),
		Severity:  "high",
		Status:    "success",
		Details:   "Organization soft deleted",
		IPAddress: "0.0.0.0",
		UserAgent: "system",
	})
	return nil
}

func (s *AuthService) SoftDeleteUser(ctx context.Context, id uuid.UUID) error {
	user, _ := s.repo.GetUserByID(ctx, id)
	if err := s.repo.SoftDeleteUser(ctx, id); err != nil {
		return err
	}

	if err := s.repo.DeleteTokensByUserID(ctx, id); err != nil {
		s.logger.Warn("Failed to invalidate user sessions during soft delete", zap.Error(err), zap.String("user_id", id.String()))
	}

	actorID, _ := ctx.Value("user_id").(uuid.UUID)
	orgID := uuid.Nil
	if user != nil {
		orgID = user.OrgID
	}
	s.repo.CreateAuditLog(ctx, &models.AuditLog{
		OrgID:     orgID,
		UserID:    actorID,
		Module:    "users",
		Action:    "soft_delete",
		Resource:  "user",
		EntityID:  id.String(),
		Severity:  "high",
		Status:    "success",
		Details:   "User soft deleted and active sessions invalidated",
		IPAddress: "0.0.0.0",
		UserAgent: "system",
	})
	return nil
}

// Vehicles (Gateway)
// Vehicles (Direct DB)
func (s *AuthService) ListVehicles(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Vehicle, int, error) {
	return s.repo.ListVehicles(ctx, orgID, page, size)
}

func (s *AuthService) CreateVehicle(ctx context.Context, v *models.Vehicle) (*models.Vehicle, error) {
	v.ID = uuid.New()
	v.Status = "active"
	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()

	if err := s.repo.CreateVehicle(ctx, v); err != nil {
		return nil, err
	}
	// Audit
	// _ = s.repo.CreateAuditLog(...)
	return v, nil
}

func (s *AuthService) GetVehicleByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) {
	vehicle, err := s.repo.GetVehicleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify that the user has access to this vehicle
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, errors.New("user not authenticated")
	}

	// Check if user has access to this vehicle's organization
	userOrgs, err := s.repo.GetOrganizationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	hasAccess := false
	for _, org := range userOrgs {
		if org.ID == vehicle.OrgID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, errors.New("unauthorized access to vehicle")
	}

	return vehicle, nil
}

func (s *AuthService) UpdateVehicle(ctx context.Context, v *models.Vehicle) (*models.Vehicle, error) {
	// Verify that the user has access to this vehicle
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, errors.New("user not authenticated")
	}

	// Check if user has access to this vehicle's organization
	userOrgs, err := s.repo.GetOrganizationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	hasAccess := false
	for _, org := range userOrgs {
		if org.ID == v.OrgID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, errors.New("unauthorized access to vehicle")
	}

	if err := s.repo.UpdateVehicle(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *AuthService) UpdateVehicleStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.repo.UpdateVehicleStatus(ctx, id, status)
}

func (s *AuthService) SoftDeleteVehicle(ctx context.Context, id uuid.UUID) error {
	// Get the vehicle to check if user has access
	vehicle, err := s.repo.GetVehicleByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify that the user has access to this vehicle
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return errors.New("user not authenticated")
	}

	// Check if user has access to this vehicle's organization
	userOrgs, err := s.repo.GetOrganizationsByUserID(ctx, userID)
	if err != nil {
		return err
	}

	hasAccess := false
	for _, org := range userOrgs {
		if org.ID == vehicle.OrgID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return errors.New("unauthorized access to vehicle")
	}

	return s.repo.DeleteVehicle(ctx, id)
}

// Drivers (Gateway)
func (s *AuthService) ListDrivers(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Driver, int, error) {
	url := fmt.Sprintf("%s/drivers?page=%d&page_size=%d&org_id=%s", s.fleetCoreServiceURL, page, size, orgID)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("service returned %d", resp.StatusCode)
	}

	var result struct {
		Drivers []models.Driver `json:"drivers"`
		Total   int             `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}
	return result.Drivers, result.Total, nil
}

func (s *AuthService) CreateDriver(ctx context.Context, name, phone string, orgID uuid.UUID) (*models.Driver, error) {
	reqBody := map[string]interface{}{
		"first_name": name, "last_name": ".", "phone": phone, "org_id": orgID,
		"email":          fmt.Sprintf("%s@flyfleet.com", uuid.New().String()[:8]), // Placeholder
		"license_number": "TEMP-" + uuid.New().String()[:8],
	}
	body, _ := json.Marshal(reqBody)

	resp, err := s.httpClient.Post(s.fleetCoreServiceURL+"/drivers", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service returned %d", resp.StatusCode)
	}

	var d models.Driver
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *AuthService) SoftDeleteDriver(ctx context.Context, id uuid.UUID) error {
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/drivers/%s", s.fleetCoreServiceURL, id), nil)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to delete")
	}
	return nil
}

// Devices (Gateway) - Placeholder for when Device Service is ready
func (s *AuthService) ListDevices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Device, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	offset := (page - 1) * size
	url := fmt.Sprintf("%s/api/v1/diagnostics/devices?limit=%d&offset=%d", s.fleetCoreServiceURL, size, offset)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err == nil {
		req.Header.Set("X-Organization-ID", orgID.String())
		resp, doErr := s.httpClient.Do(req)
		if doErr == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var payload []struct {
					ID           uuid.UUID  `json:"id"`
					OrgID        uuid.UUID  `json:"org_id"`
					VehicleID    *uuid.UUID `json:"vehicle_id"`
					SerialNumber string     `json:"serial_number"`
					Type         string     `json:"type"`
					Status       string     `json:"status"`
					LastPingAt   *time.Time `json:"last_ping_at"`
					IsDeleted    bool       `json:"is_deleted"`
					CreatedAt    time.Time  `json:"created_at"`
					UpdatedAt    time.Time  `json:"updated_at"`
				}
				if decErr := json.NewDecoder(resp.Body).Decode(&payload); decErr == nil {
					devices := make([]models.Device, 0, len(payload))
					for _, d := range payload {
						devices = append(devices, models.Device{
							ID:           d.ID,
							OrgID:        d.OrgID,
							VehicleID:    d.VehicleID,
							SerialNumber: d.SerialNumber,
							Type:         d.Type,
							Status:       d.Status,
							LastPingAt:   d.LastPingAt,
							IsDeleted:    d.IsDeleted,
							CreatedAt:    d.CreatedAt,
							UpdatedAt:    d.UpdatedAt,
						})
					}
					return devices, len(devices), nil
				}
			}
		}
	}

	// Fallback to auth DB-backed device listing when diagnostics endpoint is unavailable.
	return s.repo.ListDevices(ctx, orgID, page, size)
}

func (s *AuthService) CreateDevice(ctx context.Context, serial, devType string, orgID uuid.UUID, vehicleID *uuid.UUID) (*models.Device, error) {
	// Fallback to DB
	d := &models.Device{
		ID:           uuid.New(),
		OrgID:        orgID,
		VehicleID:    vehicleID,
		SerialNumber: serial,
		Type:         devType,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.repo.CreateDevice(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *AuthService) GetDeviceByID(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	return s.repo.GetDeviceByID(ctx, id)
}

func (s *AuthService) AssignDeviceToVehicle(ctx context.Context, deviceID, vehicleID uuid.UUID) error {
	return s.repo.AssignDeviceToVehicle(ctx, deviceID, vehicleID)
}

func (s *AuthService) UnassignDevice(ctx context.Context, deviceID uuid.UUID) error {
	return s.repo.UnassignDevice(ctx, deviceID)
}

func (s *AuthService) AssignDriverToVehicle(ctx context.Context, driverID, vehicleID uuid.UUID) error {
	return s.repo.AssignDriverToVehicle(ctx, driverID, vehicleID)
}

func (s *AuthService) SoftDeleteDevice(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteDevice(ctx, id)
}

// Documents
func (s *AuthService) CreateDocument(ctx context.Context, d *models.Document) (*models.Document, error) {
	d.ID = uuid.New()
	if d.Status == "" {
		d.Status = "pending"
	}
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()

	if err := s.repo.CreateDocument(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *AuthService) GetDocumentByID(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	doc, err := s.repo.GetDocumentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify that the user has access to this document
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, errors.New("user not authenticated")
	}

	// Check if user has access to this document's vehicle's organization
	userOrgs, err := s.repo.GetOrganizationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	hasAccess := false
	for _, org := range userOrgs {
		if org.ID == doc.OrgID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, errors.New("unauthorized access to document")
	}

	return doc, nil
}

func (s *AuthService) ListDocumentsByVehicleID(ctx context.Context, vehicleID uuid.UUID) ([]models.Document, error) {
	return s.repo.ListDocumentsByVehicleID(ctx, vehicleID)
}

func (s *AuthService) ListDocuments(ctx context.Context, orgID uuid.UUID, filter *models.DocumentFilter) (*models.DocumentListResponse, error) {
	docs, total, err := s.repo.ListDocuments(ctx, orgID, filter)
	if err != nil {
		return nil, err
	}
	if docs == nil {
		docs = []models.Document{}
	}
	return &models.DocumentListResponse{
		Data: docs,
		Meta: models.DocumentListMeta{
			Total: total,
			Page:  filter.Page,
			Limit: filter.Limit,
		},
	}, nil
}

func (s *AuthService) UpdateDocument(ctx context.Context, d *models.Document) (*models.Document, error) {
	d.UpdatedAt = time.Now()
	if err := s.repo.UpdateDocument(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *AuthService) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteDocument(ctx, id)
}

func (s *AuthService) issueTokens(ctx context.Context, user *models.User, ip, ua string) (*models.AuthResponse, error) {
	// Validate and sanitize the IP address
	sanitizedIP, err := validateAndSanitizeIP(ip)
	if err != nil {
		s.logger.Warn("Invalid IP address provided for token issuance", zap.String("ip", ip), zap.Error(err))
		sanitizedIP = "0.0.0.0" // Default fallback
	}

	// Check concurrent session limits
	activeSessions, err := s.repo.GetTokensByUserID(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to check active sessions", zap.Error(err))
		// Continue anyway to not block login
	} else if len(activeSessions) >= 5 { // Max 5 concurrent sessions
		// Remove oldest session to stay within limit
		oldest := activeSessions[0]
		for _, session := range activeSessions {
			if session.CreatedAt.Before(oldest.CreatedAt) {
				oldest = session
			}
		}
		s.repo.DeleteToken(ctx, oldest.AccessToken)
	}

	accessToken, _ := s.generateJWT(user.ID, 15*time.Minute)
	refreshToken, _ := s.generateJWT(user.ID, 7*24*time.Hour)

	token := &models.AuthToken{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IPAddress:    sanitizedIP,
		UserAgent:    ua,
		LastActive:   time.Now(),
		IsActive:     true,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	if err := s.repo.SaveToken(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) issueTokensWithOrg(ctx context.Context, user *models.User, orgID uuid.UUID, ip, ua string) (*models.AuthResponse, error) {
	// Validate and sanitize the IP address
	sanitizedIP, err := validateAndSanitizeIP(ip)
	if err != nil {
		s.logger.Warn("Invalid IP address provided for token issuance", zap.String("ip", ip), zap.Error(err))
		sanitizedIP = "0.0.0.0" // Default fallback
	}

	// Fetch user role in this organization
	role, err := s.repo.GetUserRoleInOrg(ctx, user.ID, orgID)
	if err != nil {
		role = "member" // Fallback
	}

	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"org_id":  orgID.String(),
		"role":    role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	accessToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
	refreshToken, _ := s.generateJWT(user.ID, 7*24*time.Hour)

	// Save token with enrichment
	s.repo.SaveToken(ctx, &models.AuthToken{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IPAddress:    sanitizedIP,
		UserAgent:    ua,
		LastActive:   time.Now(),
		IsActive:     true,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	})

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) generateJWT(userID uuid.UUID, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) generateActionToken(userID uuid.UUID, purpose string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":     userID.String(),
		"purpose": purpose,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) parseActionToken(tokenString, expectedPurpose string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid claims")
	}

	purpose, _ := claims["purpose"].(string)
	if purpose != expectedPurpose {
		return uuid.Nil, errors.New("invalid token purpose")
	}

	subject, _ := claims["sub"].(string)
	if subject == "" {
		// Backward compatibility for earlier format.
		subject, _ = claims["user_id"].(string)
	}
	if subject == "" {
		return uuid.Nil, errors.New("missing subject")
	}

	return uuid.Parse(subject)
}

// Trips (Gateway)
func (s *AuthService) ListTrips(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Trip, int, error) {
	url := fmt.Sprintf("%s/api/v1/trips?page=%d&page_size=%d&org_id=%s", s.fleetCoreServiceURL, page, size, orgID)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("service returned %d", resp.StatusCode)
	}

	var result struct {
		Trips []models.Trip `json:"trips"`
		Total int           `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}
	return result.Trips, result.Total, nil
}

func (s *AuthService) CreateTrip(ctx context.Context, trip *models.Trip) (*models.Trip, error) {
	body, _ := json.Marshal(trip)
	resp, err := s.httpClient.Post(s.fleetCoreServiceURL+"/api/v1/trips", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service returned %d", resp.StatusCode)
	}

	var t models.Trip
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Alerts (Gateway)
func (s *AuthService) ListAlerts(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Alert, int, error) {
	url := fmt.Sprintf("%s/alerts?page=%d&page_size=%d&org_id=%s", s.alertServiceURL, page, size, orgID)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("service returned %d", resp.StatusCode)
	}

	var result struct {
		Alerts []models.Alert `json:"alerts"`
		Total  int            `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}
	return result.Alerts, result.Total, nil
}

func (s *AuthService) GetAlertSummary(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/alerts/stats/summary?org_id=%s", s.alertServiceURL, orgID)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service returned %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// Analytics (Gateway)
func (s *AuthService) GetDriverStats(ctx context.Context, orgID uuid.UUID) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/stats/drivers?org_id=%s", s.fleetCoreServiceURL, orgID)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service returned %d", resp.StatusCode)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *AuthService) GetTripStats(ctx context.Context, orgID uuid.UUID) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/stats/trips?org_id=%s", s.fleetCoreServiceURL, orgID)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service returned %d", resp.StatusCode)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// Security Suite Service Implementation

func (s *AuthService) GetMFADashboardSummary(ctx context.Context, orgID uuid.UUID) (*models.MFADashboardSummary, error) {
	return s.repo.GetMFADashboardSummary(ctx, orgID)
}

func (s *AuthService) GetMFAAdoptionTrend(ctx context.Context, orgID uuid.UUID, days int) ([]map[string]any, error) {
	return s.repo.GetMFAAdoptionTrend(ctx, orgID, days)
}

func (s *AuthService) GetMFAMethodDistribution(ctx context.Context, orgID uuid.UUID) (map[string]int, error) {
	return s.repo.GetMFAMethodDistribution(ctx, orgID)
}

func (s *AuthService) GetMFAEvents(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.MFAEvent, int, error) {
	return s.repo.GetMFAEvents(ctx, orgID, page, size, filters)
}

func (s *AuthService) GetTopFailureLocations(ctx context.Context, orgID uuid.UUID) ([]map[string]any, error) {
	return s.repo.GetTopFailureLocations(ctx, orgID)
}

func (s *AuthService) GetTopFailureUsers(ctx context.Context, orgID uuid.UUID) ([]map[string]any, error) {
	return s.repo.GetTopFailureUsers(ctx, orgID)
}

func (s *AuthService) GetLockedUsers(ctx context.Context, orgID uuid.UUID, page, limit int) ([]models.User, int, error) {
	return s.repo.GetLockedUsers(ctx, orgID, page, limit)
}

func (s *AuthService) UnlockUser(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("Unlocking user account", zap.String("userID", userID.String()))
	return s.repo.UnlockUser(ctx, userID)
}

func (s *AuthService) ResetMFA(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("Resetting user MFA", zap.String("userID", userID.String()))
	return s.repo.ResetMFA(ctx, userID)
}

func (s *AuthService) GetMFAPolicy(ctx context.Context, orgID uuid.UUID) (*models.MFAPolicy, error) {
	policy, err := s.repo.GetMFAPolicy(ctx, orgID)
	if err != nil {
		// Return default policy if not found
		return &models.MFAPolicy{
			OrgID:                 orgID,
			RequireMFA:            false,
			AllowedMethods:        []string{"totp", "sms"},
			MaxAttempts:           5,
			LockoutMinutes:        30,
			GracePeriodDays:       7,
			RememberDeviceDays:    30,
			EnforceAtLogin:        true,
			AdminOnlyMode:         false,
			RememberDeviceAllowed: true,
			MFAChallengeTimeout:   60,
			Version:               0,
		}, nil
	}
	return policy, nil
}

func (s *AuthService) UpdateMFAPolicy(ctx context.Context, p *models.MFAPolicy, updaterID uuid.UUID, reason string) error {
	s.logger.Info("Updating MFA policy",
		zap.String("orgID", p.OrgID.String()),
		zap.String("updaterID", updaterID.String()),
	)

	err := s.repo.UpdateMFAPolicy(ctx, p, updaterID, reason)
	if err != nil {
		return err
	}

	// Log Audit
	details, _ := json.Marshal(p)
	_ = s.repo.CreateAuditLog(ctx, &models.AuditLog{
		OrgID:     p.OrgID,
		UserID:    updaterID,
		Action:    "MFA_POLICY_UPDATE",
		Resource:  "mfa_policy",
		Severity:  "warning",
		Status:    "success",
		Details:   string(details),
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *AuthService) GetMFAPolicyHistory(ctx context.Context, orgID uuid.UUID) ([]models.MFAPolicyHistory, error) {
	return s.repo.GetMFAPolicyHistory(ctx, orgID)
}

func (s *AuthService) RollbackMFAPolicy(ctx context.Context, historyID uuid.UUID, updaterID uuid.UUID) error {
	s.logger.Info("Rolling back MFA policy",
		zap.String("historyID", historyID.String()),
		zap.String("updaterID", updaterID.String()),
	)

	err := s.repo.RollbackMFAPolicy(ctx, historyID)
	if err != nil {
		return err
	}

	// Log Audit
	_ = s.repo.CreateAuditLog(ctx, &models.AuditLog{
		OrgID:     uuid.Nil, // Should fetch from historyID if needed, but for now Nil or from history
		UserID:    updaterID,
		Action:    "MFA_POLICY_ROLLBACK",
		Resource:  "mfa_policy",
		Severity:  "critical",
		Status:    "success",
		Details:   fmt.Sprintf(`{"history_id": "%s"}`, historyID.String()),
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *AuthService) GetMFAOrgOverride(ctx context.Context, orgID uuid.UUID) (*models.MFAOrgOverride, error) {
	return s.repo.GetMFAOrgOverride(ctx, orgID)
}

func (s *AuthService) UpdateMFAOrgOverride(ctx context.Context, o *models.MFAOrgOverride, updaterID uuid.UUID) error {
	s.logger.Info("Updating MFA Org Override",
		zap.String("orgID", o.OrgID.String()),
		zap.String("updaterID", updaterID.String()),
	)

	err := s.repo.UpdateMFAOrgOverride(ctx, o)
	if err != nil {
		return err
	}

	// Log Audit
	details, _ := json.Marshal(o)
	_ = s.repo.CreateAuditLog(ctx, &models.AuditLog{
		OrgID:     o.OrgID,
		UserID:    updaterID,
		Action:    "MFA_ORG_OVERRIDE_UPDATE",
		Resource:  "mfa_org_override",
		Severity:  "warning",
		Status:    "success",
		Details:   string(details),
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *AuthService) AddMFAException(ctx context.Context, userID, orgID, grantedBy uuid.UUID, reason string, days int) error {
	expiresAt := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	exc := &models.MFAException{
		UserID:    userID,
		OrgID:     orgID,
		Reason:    reason,
		GrantedBy: grantedBy,
		ExpiresAt: &expiresAt,
	}

	err := s.repo.AddMFAException(ctx, exc)
	if err != nil {
		return err
	}

	// Log Audit
	details, _ := json.Marshal(exc)
	_ = s.repo.CreateAuditLog(ctx, &models.AuditLog{
		OrgID:     orgID,
		UserID:    grantedBy,
		Action:    "MFA_EXCEPTION_ADD",
		Resource:  "mfa_exception",
		Severity:  "warning",
		Status:    "success",
		Details:   string(details),
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *AuthService) GetMFAExceptions(ctx context.Context, orgID uuid.UUID, page, limit int) ([]models.MFAException, int, error) {
	return s.repo.GetMFAExceptions(ctx, orgID, page, limit)
}

func (s *AuthService) DeleteMFAException(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteMFAException(ctx, id)
}

// RBAC Operations

func (s *AuthService) CreateRole(ctx context.Context, name, description string, orgID uuid.UUID, isSystem bool, actorID uuid.UUID) (*models.Role, error) {
	// Unique Name Check
	existing, err := s.repo.GetRoleByName(ctx, orgID, name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("role with this name already exists")
	}

	var orgIDPtr *uuid.UUID
	if orgID != uuid.Nil {
		orgIDPtr = &orgID
	}

	role := &models.Role{
		ID:             uuid.New(),
		Name:           name,
		Description:    description,
		OrganizationID: orgIDPtr,
		IsSystem:       isSystem,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	// Audit Log
	s.logRBACAudit(ctx, actorID, "ROLE_CREATE", "role", fmt.Sprintf(`{"role_id": "%s", "name": "%s", "org_id": "%s"}`, role.ID, name, orgID), orgID)

	return role, nil
}

func (s *AuthService) GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	return s.repo.GetRoleByID(ctx, id)
}

func (s *AuthService) UpdateRole(ctx context.Context, id uuid.UUID, name, description string, actorID uuid.UUID) (*models.Role, error) {
	role, err := s.repo.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role.IsSystem {
		return nil, errors.New("cannot update system roles")
	}

	// Unique Name Check
	if role.OrganizationID != nil {
		existing, err := s.repo.GetRoleByName(ctx, *role.OrganizationID, name)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, errors.New("role with this name already exists")
		}
	}

	oldName := role.Name
	role.Name = name
	role.Description = description
	role.UpdatedAt = time.Now()

	if err := s.repo.UpdateRole(ctx, role); err != nil {
		return nil, err
	}

	// Audit Log
	orgID := uuid.Nil
	if role.OrganizationID != nil {
		orgID = *role.OrganizationID
	}
	s.logRBACAudit(ctx, actorID, "ROLE_UPDATE", "role", fmt.Sprintf(`{"role_id": "%s", "old_name": "%s", "new_name": "%s"}`, id, oldName, name), orgID)

	return role, nil
}

func (s *AuthService) DeleteRole(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error {
	role, err := s.repo.GetRoleByID(ctx, id)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return errors.New("cannot delete system roles")
	}

	if err := s.repo.DeleteRole(ctx, id); err != nil {
		return err
	}

	// Audit Log
	orgID := uuid.Nil
	if role.OrganizationID != nil {
		orgID = *role.OrganizationID
	}
	s.logRBACAudit(ctx, actorID, "ROLE_DELETE", "role", fmt.Sprintf(`{"role_id": "%s", "name": "%s"}`, id, role.Name), orgID)

	return nil
}

func (s *AuthService) ListRolesByOrg(ctx context.Context, orgID uuid.UUID) ([]models.Role, error) {
	return s.repo.ListRolesByOrg(ctx, orgID)
}

func (s *AuthService) ListPermissions(ctx context.Context) ([]models.PermissionDefinition, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *AuthService) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]models.PermissionDefinition, error) {
	return s.repo.GetRolePermissions(ctx, roleID)
}

func (s *AuthService) UpdateRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID, actorID uuid.UUID) error {
	role, err := s.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role.IsSystem && actorID != uuid.Nil {
		return errors.New("cannot modify permissions of system roles")
	}

	if err := s.repo.UpdateRolePermissions(ctx, roleID, permissionIDs); err != nil {
		return err
	}

	// Audit Log
	orgID := uuid.Nil
	if role.OrganizationID != nil {
		orgID = *role.OrganizationID
	}
	s.logRBACAudit(ctx, actorID, "ROLE_PERMISSIONS_UPDATE", "role_permissions", fmt.Sprintf(`{"role_id": "%s", "permissions_count": %d}`, roleID, len(permissionIDs)), orgID)

	return nil
}

func (s *AuthService) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	if err := s.repo.AssignRoleToUser(ctx, userID, roleID, assignedBy); err != nil {
		return err
	}

	// Audit Log
	// Fetch role to get details for log? Or just log IDs
	s.logRBACAudit(ctx, assignedBy, "USER_ROLE_ASSIGN", "user_role", fmt.Sprintf(`{"user_id": "%s", "role_id": "%s"}`, userID, roleID), uuid.Nil)
	return nil
}

func (s *AuthService) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID, actorID uuid.UUID) error { // Added actorID
	// 1. Fetch role to check if it's a critical role
	role, err := s.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}

	// 2. Safeguard: Prevent removing last SuperAdmin or OrgAdmin
	// In a real system, we'd have a more robust policy engine, but hardcoding critical roles is safe for now
	criticalRoles := map[string]bool{"super_admin": true, "org_admin": true}
	if criticalRoles[role.Name] {
		count, err := s.repo.CountUsersWithRole(ctx, role.Name, role.OrganizationID)
		if err != nil {
			return err
		}
		if count <= 1 {
			return errors.New("cannot remove the last administrator")
		}
	}

	if err := s.repo.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		return err
	}

	// Audit Log
	orgID := uuid.Nil
	if role.OrganizationID != nil {
		orgID = *role.OrganizationID
	}
	s.logRBACAudit(ctx, actorID, "USER_ROLE_REMOVE", "user_role", fmt.Sprintf(`{"user_id": "%s", "role_id": "%s"}`, userID, roleID), orgID)

	return nil
}

func (s *AuthService) CloneRole(ctx context.Context, sourceRoleID uuid.UUID, newRoleName, newRoleDescription string, actorID uuid.UUID) (*models.Role, error) {
	// 1. Validate new name uniqueness (optimization: check before DB call)
	// Need to fetch OrgID from source role first or let CloneRole handle it?
	// The repo method CloneRole does fetch the source role.
	// But to check uniqueness here, I need OrgID.
	// Let's fetch source role first.
	source, err := s.repo.GetRoleByID(ctx, sourceRoleID)
	if err != nil {
		return nil, err
	}
	if source.OrganizationID != nil {
		existing, err := s.repo.GetRoleByName(ctx, *source.OrganizationID, newRoleName)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, errors.New("role with this name already exists")
		}
	}

	newRole, err := s.repo.CloneRole(ctx, sourceRoleID, newRoleName, newRoleDescription)
	if err != nil {
		return nil, err
	}

	// Audit Log
	orgID := uuid.Nil
	if source.OrganizationID != nil {
		orgID = *source.OrganizationID
	}
	s.logRBACAudit(ctx, actorID, "ROLE_CLONE", "role", fmt.Sprintf(`{"source_role_id": "%s", "new_role_id": "%s"}`, sourceRoleID, newRole.ID), orgID)

	return newRole, nil
}

// Helper for RBAC Audit
func (s *AuthService) logRBACAudit(ctx context.Context, actorID uuid.UUID, action, resource, details string, orgID uuid.UUID) {
	_ = s.repo.CreateAuditLog(ctx, &models.AuditLog{
		OrgID:     orgID,
		UserID:    actorID,
		Module:    "RBAC",
		Action:    action,
		Resource:  resource,
		Severity:  "info",
		Status:    "success",
		Details:   details,
		CreatedAt: time.Now(),
	})
}

func (s *AuthService) GetSystemStats(ctx context.Context) (*models.SystemStats, error) {
	stats, err := s.repo.GetSystemStats(ctx)
	if err != nil {
		return nil, err
	}

	// Enrich with Driver stats from Driver Service if possible
	// For now, we'll try to fetch total drivers across all orgs?
	// The driver service endpoint ListDrivers is per-org.
	// We might need a global driver count endpoint in Driver Service or just mock it/leave it 0.
	// Let's leave it 0 or use the approximation from earlier BFF logic (which was mock).
	// Real implementation would be:
	// drivers, _, err := s.ListDrivers(ctx, uuid.Nil, 1, 1) -> if supported by driver service to list all.
	// Assuming driver service supports global list if orgID is nil or specific flag.

	return stats, nil
}

// Security Configuration Methods

func (s *AuthService) GetSecuritySettings(ctx context.Context, orgID uuid.UUID) (*models.SecuritySettings, error) {
	return s.repo.GetSecuritySettings(ctx, orgID)
}

func (s *AuthService) UpdateSecuritySettings(ctx context.Context, orgID uuid.UUID, settings *models.SecuritySettings) error {
	settings.OrgID = orgID
	settings.ID = uuid.New()
	settings.UpdatedAt = time.Now()

	return s.repo.UpdateSecuritySettings(ctx, settings)
}

func (s *AuthService) GetSecurityMetrics(ctx context.Context, orgID uuid.UUID) (*models.SecurityMetrics, error) {
	return s.repo.GetSecurityMetrics(ctx, orgID)
}

func (s *AuthService) GetSecurityPostureScore(ctx context.Context, orgID uuid.UUID) (float64, error) {
	metrics, err := s.repo.GetSecurityMetrics(ctx, orgID)
	if err != nil {
		return 0, err
	}
	return metrics.SecurityPostureScore, nil
}

func (s *AuthService) GetVulnerabilityReport(ctx context.Context, orgID uuid.UUID) (*models.VulnerabilityReport, error) {
	return s.repo.GetVulnerabilityReport(ctx, orgID)
}

func (s *AuthService) GetSecurityComplianceReport(ctx context.Context, orgID uuid.UUID, reportType string, startDate, endDate time.Time) (*models.ComplianceReport, error) {
	return s.repo.GetComplianceReport(ctx, orgID, reportType, startDate, endDate)
}

func (s *AuthService) ListSecurityEvents(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]interface{}) ([]models.SecurityEvent, int, error) {
	return s.repo.ListSecurityEvents(ctx, orgID, page, size, filters)
}

func (s *AuthService) ListEntityAuditHistory(ctx context.Context, entityType, entityID string) ([]models.EntityAuditHistory, error) {
	return s.repo.ListEntityAuditHistory(ctx, entityType, entityID)
}

func (s *AuthService) CreateEvidenceExport(ctx context.Context, orgID uuid.UUID, module string, start, end time.Time, format string, requestedBy uuid.UUID) (*models.EvidenceExport, error) {
	export := &models.EvidenceExport{
		OrgID:          orgID,
		Module:         module,
		DateRangeStart: start,
		DateRangeEnd:   end,
		Format:         format,
		Status:         "processing",
		RequestedBy:    requestedBy,
	}
	if err := s.repo.CreateEvidenceExport(ctx, export); err != nil {
		return nil, err
	}
	// In production, trigger a background worker here
	return export, nil
}

func (s *AuthService) ListEvidenceExports(ctx context.Context, orgID uuid.UUID) ([]models.EvidenceExport, error) {
	return s.repo.ListEvidenceExports(ctx, orgID)
}

func (s *AuthService) GetEvidenceExportByID(ctx context.Context, id uuid.UUID) (*models.EvidenceExport, error) {
	return s.repo.GetEvidenceExportByID(ctx, id)
}

func (s *AuthService) GetAuditPolicy(ctx context.Context, orgID uuid.UUID) (*models.AuditPolicy, error) {
	return s.repo.GetAuditPolicy(ctx, orgID)
}

func (s *AuthService) UpdateAuditPolicy(ctx context.Context, policy *models.AuditPolicy) error {
	return s.repo.UpdateAuditPolicy(ctx, policy)
}

func (s *AuthService) GetEffectivePermissions(ctx context.Context, userID uuid.UUID) ([]models.PermissionDefinition, error) {
	return s.repo.GetEffectivePermissions(ctx, userID)
}

func (s *AuthService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	return s.repo.GetUserRoles(ctx, userID)
}

func (s *AuthService) logLoginActivity(ctx context.Context, user *models.User, status, risk, ip, ua string) {
	// Validate and sanitize the IP address
	sanitizedIP, err := validateAndSanitizeIP(ip)
	if err != nil {
		s.logger.Warn("Invalid IP address provided for logging", zap.String("ip", ip), zap.Error(err))
		// Use a default value or skip logging if IP is invalid
		sanitizedIP = "0.0.0.0" // Default fallback
	}

	orgID := uuid.Nil
	userID := uuid.Nil
	email := ""
	if user != nil {
		userID = user.ID
		email = user.Email
		// Fetch orgID if possible
		orgs, _ := s.repo.GetOrganizationsByUserID(ctx, user.ID)
		if len(orgs) > 0 {
			orgID = orgs[0].ID
		}
	}

	event := &models.LoginEvent{
		OrgID:     orgID,
		UserID:    &userID,
		Email:     email,
		IPAddress: sanitizedIP,
		Device:    "Unknown", // In production, use a UA parser
		Browser:   ua,
		Status:    status,
		RiskLevel: risk,
		CreatedAt: time.Now(),
	}
	_ = s.repo.LogLoginEvent(ctx, event)

	// Also log to security_events for compatibility
	_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    userID,
		EventType: "LOGIN_" + status,
		IPAddress: sanitizedIP,
		UserAgent: ua,
		CreatedAt: time.Now(),
	})
}
