package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/flyfleet/auth-service/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo              repository.AuthRepository
	jwtSecret         string
	logger            *zap.Logger
	driverServiceURL  string
	vehicleServiceURL string
	tripServiceURL    string
	alertServiceURL   string
	httpClient        *http.Client
}

func NewAuthService(repo repository.AuthRepository, jwtSecret, driverURL, vehicleURL, tripURL, alertURL string, logger *zap.Logger) *AuthService {
	return &AuthService{
		repo:              repo,
		jwtSecret:         jwtSecret,
		logger:            logger,
		driverServiceURL:  driverURL,
		vehicleServiceURL: vehicleURL,
		tripServiceURL:    tripURL,
		alertServiceURL:   alertURL,
		httpClient:        &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	return s.repo.DeleteToken(ctx, accessToken)
}

func (s *AuthService) GetSecurityEvents(ctx context.Context, userID uuid.UUID) ([]models.SecurityEvent, error) {
	return s.repo.GetSecurityEvents(ctx, userID)
}

func (s *AuthService) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]models.AuthToken, error) {
	return s.repo.GetTokensByUserID(ctx, userID)
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
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logLoginActivity(ctx, nil, "FAILED", "low", ip, ua)
		return nil, err
	}
	if user == nil {
		s.logLoginActivity(ctx, nil, "FAILED", "low", ip, ua)
		return nil, errors.New("invalid credentials")
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		s.logLoginActivity(ctx, user, "FAILED", "high", ip, ua)
		return nil, errors.New("account is temporarily locked due to multiple failed attempts")
	}

	if user.FailedLogins >= 3 {
		s.logLoginActivity(ctx, user, "CHALLENGED", "medium", ip, ua)
		return nil, errors.New("CAPTCHA_REQUIRED")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// Increment failed logins and log event
		count, _ := s.repo.IncrementFailedLogins(ctx, user.ID)
		s.logLoginActivity(ctx, user, "FAILED", "low", ip, ua)

		// Lock account if threshold reached (e.g., 5 attempts)
		if count >= 5 {
			s.repo.LockAccount(ctx, user.ID, 15*time.Minute)
			// Create Threat Alert
			s.repo.CreateThreatAlert(ctx, &models.ThreatAlert{
				OrgID:       uuid.Nil, // Will be enriched if possible
				Type:        "BRUTE_FORCE",
				Severity:    "critical",
				IP:          ip,
				UserID:      &user.ID,
				Description: fmt.Sprintf("Multiple failed login attempts for user %s", user.Email),
				Status:      "open",
			})
		}

		return nil, errors.New("invalid credentials")
	}

	// Reset failed logins on success
	s.repo.ResetFailedLogins(ctx, user.ID)
	s.logLoginActivity(ctx, user, "SUCCESS", "low", ip, ua)

	if user.MFAEnabled {
		return &models.AuthResponse{
			MFAWait: true,
			User:    *user,
		}, nil
	}

	// Fetch user organizations to set default context
	orgs, err := s.repo.GetOrganizationsByUserID(ctx, user.ID)
	if err == nil && len(orgs) > 0 {
		return s.issueTokensWithOrg(ctx, user, orgs[0].ID, ip, ua)
	}

	return s.issueTokens(ctx, user, ip, ua)
}

func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *AuthService) Register(ctx context.Context, email, password, name, ip, ua string) (*models.User, error) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
		Status:       "active",
		Verified:     false, // Requires email verification
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    user.ID,
		EventType: "USER_REGISTERED",
		IPAddress: ip,
		UserAgent: ua,
	})

	// TODO: Send verification email
	return user, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	// For simplicity, we'll assume the token is the user ID for now or a JWT
	// In a real app, you'd lookup the token in a verification_tokens table
	userID, err := uuid.Parse(token)
	if err != nil {
		return errors.New("invalid verification token")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	user.Verified = true
	// We need an UpdateUserVerified method or just use UpdateUser
	return s.repo.UpdateUser(ctx, user)
}

func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return nil // Don't reveal if user exists
	}
	if user.Verified {
		return errors.New("email already verified")
	}
	// TODO: Trigger email send logic
	return nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return nil // Don't reveal
	}
	// TODO: Generate reset token and send email
	s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    user.ID,
		EventType: "PASSWORD_RESET_REQUESTED",
	})
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// For simplicity, token is userID for now
	userID, err := uuid.Parse(token)
	if err != nil {
		return errors.New("invalid reset token")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	user.PasswordHash = string(hashedPassword)
	
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}

	s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:    user.ID,
		EventType: "PASSWORD_RESET_SUCCESS",
	})
	return nil
}

func (s *AuthService) SendMFA(ctx context.Context, userID uuid.UUID, method string) error {
	code := "123456" // Mock OTP
	mfa := &models.MFACode{
		UserID:    userID,
		Code:      code,
		Method:    method,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	return s.repo.SaveMFACode(ctx, mfa)
}

func (s *AuthService) VerifyMFA(ctx context.Context, userID uuid.UUID, code, ip, ua string) (*models.AuthResponse, error) {
	mfa, err := s.repo.GetMFACode(ctx, userID, code)
	if err != nil {
		return nil, err
	}
	if mfa == nil {
		return nil, errors.New("invalid or expired OTP")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user, ip, ua)
}

func (s *AuthService) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]models.Organization, error) {
	return s.repo.GetOrganizationsByUserID(ctx, userID)
}

func (s *AuthService) SelectOrganization(ctx context.Context, userID, orgID uuid.UUID, ip, ua string) (*models.AuthResponse, error) {
	orgs, err := s.repo.GetOrganizationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	valid := false
	for _, o := range orgs {
		if o.ID == orgID {
			valid = true
			break
		}
	}
	if !valid {
		return nil, errors.New("unauthorized organization access")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.issueTokensWithOrg(ctx, user, orgID, ip, ua)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string, ip, ua string) (*models.AuthResponse, error) {
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

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user, ip, ua)
}

func (s *AuthService) UpdateTheme(ctx context.Context, userID uuid.UUID, theme, ip, ua string) error {
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
		IPAddress: ip,
		UserAgent: ua,
	}
	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		s.logger.Error("Failed to create audit log for theme update", zap.Error(err))
	}

	return nil
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

func (s *AuthService) ListUsers(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.User, int, error) {
	return s.repo.ListUsers(ctx, orgID, page, size)
}

func (s *AuthService) SoftDeleteOrganization(ctx context.Context, id uuid.UUID) error {
	// TODO: Audit Log
	return s.repo.SoftDeleteOrganization(ctx, id)
}

func (s *AuthService) SoftDeleteUser(ctx context.Context, id uuid.UUID) error {
	// TODO: Audit Log, Invalidate tokens
	return s.repo.SoftDeleteUser(ctx, id)
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
	return s.repo.GetVehicleByID(ctx, id)
}

func (s *AuthService) UpdateVehicle(ctx context.Context, v *models.Vehicle) (*models.Vehicle, error) {
	if err := s.repo.UpdateVehicle(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *AuthService) UpdateVehicleStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.repo.UpdateVehicleStatus(ctx, id, status)
}

func (s *AuthService) SoftDeleteVehicle(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteVehicle(ctx, id)
}

// Drivers (Gateway)
func (s *AuthService) ListDrivers(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Driver, int, error) {
	url := fmt.Sprintf("%s/drivers?page=%d&page_size=%d&org_id=%s", s.driverServiceURL, page, size, orgID)
	resp, err := s.httpClient.Get(url)
	if err != nil { return nil, 0, err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { return nil, 0, fmt.Errorf("service returned %d", resp.StatusCode) }

	var result struct {
		Drivers []models.Driver `json:"drivers"`
		Total   int             `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { return nil, 0, err }
	return result.Drivers, result.Total, nil
}

func (s *AuthService) CreateDriver(ctx context.Context, name, phone string, orgID uuid.UUID) (*models.Driver, error) {
	reqBody := map[string]interface{}{
		"first_name": name, "last_name": ".", "phone": phone, "org_id": orgID,
		"email": fmt.Sprintf("%s@flyfleet.com", uuid.New().String()[:8]), // Placeholder
		"license_number": "TEMP-" + uuid.New().String()[:8],
	}
	body, _ := json.Marshal(reqBody)
	
	resp, err := s.httpClient.Post(s.driverServiceURL+"/drivers", "application/json", bytes.NewBuffer(body))
	if err != nil { return nil, err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service returned %d", resp.StatusCode) 
	}

	var d models.Driver
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil { return nil, err }
	return &d, nil
}

func (s *AuthService) SoftDeleteDriver(ctx context.Context, id uuid.UUID) error {
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/drivers/%s", s.driverServiceURL, id), nil)
	resp, err := s.httpClient.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode >= 400 { return fmt.Errorf("failed to delete") }
	return nil
}

// Devices (Gateway) - Placeholder for when Device Service is ready
func (s *AuthService) ListDevices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Device, int, error) {
	// TODO: Point to actual Device Service (using Vehicle Service for now as placeholder or separate)
	// For now, fallback to DB if device service not up, OR return empty
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
	if err := s.repo.CreateDevice(ctx, d); err != nil { return nil, err }
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
	return s.repo.GetDocumentByID(ctx, id)
}

func (s *AuthService) ListDocumentsByVehicleID(ctx context.Context, vehicleID uuid.UUID) ([]models.Document, error) {
	return s.repo.ListDocumentsByVehicleID(ctx, vehicleID)
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
	accessToken, _ := s.generateJWT(user.ID, 15*time.Minute)
	refreshToken, _ := s.generateJWT(user.ID, 7*24*time.Hour)

	s.repo.SaveToken(ctx, &models.AuthToken{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IPAddress:    ip,
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

func (s *AuthService) issueTokensWithOrg(ctx context.Context, user *models.User, orgID uuid.UUID, ip, ua string) (*models.AuthResponse, error) {
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
		IPAddress:    ip,
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

// Trips (Gateway)
func (s *AuthService) ListTrips(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Trip, int, error) {
	url := fmt.Sprintf("%s/trips?page=%d&page_size=%d&org_id=%s", s.tripServiceURL, page, size, orgID)
	resp, err := s.httpClient.Get(url)
	if err != nil { return nil, 0, err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { return nil, 0, fmt.Errorf("service returned %d", resp.StatusCode) }

	var result struct {
		Trips []models.Trip `json:"trips"`
		Total int           `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { return nil, 0, err }
	return result.Trips, result.Total, nil
}

func (s *AuthService) CreateTrip(ctx context.Context, trip *models.Trip) (*models.Trip, error) {
	body, _ := json.Marshal(trip)
	resp, err := s.httpClient.Post(s.tripServiceURL+"/trips", "application/json", bytes.NewBuffer(body))
	if err != nil { return nil, err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service returned %d", resp.StatusCode)
	}

	var t models.Trip
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil { return nil, err }
	return &t, nil
}

// Alerts (Gateway)
func (s *AuthService) ListAlerts(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Alert, int, error) {
	url := fmt.Sprintf("%s/alerts?page=%d&page_size=%d&org_id=%s", s.alertServiceURL, page, size, orgID)
	resp, err := s.httpClient.Get(url)
	if err != nil { return nil, 0, err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { return nil, 0, fmt.Errorf("service returned %d", resp.StatusCode) }

	var result struct {
		Alerts []models.Alert `json:"alerts"`
		Total  int            `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { return nil, 0, err }
	return result.Alerts, result.Total, nil
}

func (s *AuthService) GetAlertSummary(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/alerts/stats/summary?org_id=%s", s.alertServiceURL, orgID)
	resp, err := s.httpClient.Get(url)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("service returned %d", resp.StatusCode) }

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { return nil, err }
	return result, nil
}

// Analytics (Gateway)
func (s *AuthService) GetDriverStats(ctx context.Context, orgID uuid.UUID) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/stats/drivers?org_id=%s", s.tripServiceURL, orgID)
	resp, err := s.httpClient.Get(url)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("service returned %d", resp.StatusCode) }

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { return nil, err }
	return result, nil
}

func (s *AuthService) GetTripStats(ctx context.Context, orgID uuid.UUID) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/stats/trips?org_id=%s", s.tripServiceURL, orgID)
	resp, err := s.httpClient.Get(url)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("service returned %d", resp.StatusCode) }

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { return nil, err }
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

func (s *AuthService) GetLockedUsers(ctx context.Context, orgID uuid.UUID) ([]models.User, error) {
	return s.repo.GetLockedUsers(ctx, orgID)
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
			OrgID:                orgID,
			RequireMFA:           false,
			AllowedMethods:       []string{"totp", "sms"},
			MaxAttempts:          5,
			LockoutMinutes:       30,
			GracePeriodDays:      7,
			RememberDeviceDays:   30,
			EnforceAtLogin:       true,
			AdminOnlyMode:        false,
			RememberDeviceAllowed: true,
			MFAChallengeTimeout:  60,
			Version:              0,
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
	if err != nil { return err }

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
	if err != nil { return err }

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
	if err != nil { return err }

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
	if err != nil { return err }

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

func (s *AuthService) GetMFAExceptions(ctx context.Context, orgID uuid.UUID) ([]models.MFAException, error) {
	return s.repo.GetMFAExceptions(ctx, orgID)
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
	if role.OrganizationID != nil { orgID = *role.OrganizationID }
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
	if role.OrganizationID != nil { orgID = *role.OrganizationID }
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
	if role.OrganizationID != nil { orgID = *role.OrganizationID }
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
	if role.OrganizationID != nil { orgID = *role.OrganizationID }
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
	if source.OrganizationID != nil { orgID = *source.OrganizationID }
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
		IPAddress: ip,
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
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now(),
	})
}