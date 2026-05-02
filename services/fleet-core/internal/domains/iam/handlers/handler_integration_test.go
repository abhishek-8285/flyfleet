package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/flyfleet/fleet-core/internal/domains/iam/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ─── Mock services ───

type mockAuthService struct {
	identifyResp    *models.IdentifyResponse
	identifyErr     error
	loginResp       *models.AuthTokensResponse
	loginMFAResp    *models.MFARequiredResponse
	loginErr        error
	refreshResp     *models.AuthTokensResponse
	refreshErr      error
	registerResp    *models.AuthTokensResponse
	registerErr     error
	forgotErr       error
	resetErr        error
	logoutErr       error
	switchTenantResp *models.TenantSwitchResponse
	switchTenantErr  error
	tenants          []models.TenantInfo
	tenantsErr       error
}

func (m *mockAuthService) Identify(ctx context.Context, email string) (*models.IdentifyResponse, error) {
	return m.identifyResp, m.identifyErr
}
func (m *mockAuthService) Login(ctx context.Context, req models.LoginRequest, ip, userAgent string) (*models.AuthTokensResponse, *models.MFARequiredResponse, error) {
	return m.loginResp, m.loginMFAResp, m.loginErr
}
func (m *mockAuthService) VerifyMFA(ctx context.Context, req models.MFAVerifyRequest, ip, userAgent string) (*models.AuthTokensResponse, error) {
	return nil, nil
}
func (m *mockAuthService) Refresh(ctx context.Context, refreshToken, ip, ua string) (*models.AuthTokensResponse, error) {
	return m.refreshResp, m.refreshErr
}
func (m *mockAuthService) Register(ctx context.Context, email, password, name string, tenantID uuid.UUID) (*models.AuthTokensResponse, error) {
	return m.registerResp, m.registerErr
}
func (m *mockAuthService) ForgotPassword(ctx context.Context, email string) error {
	return m.forgotErr
}
func (m *mockAuthService) ResetPasswordWithToken(ctx context.Context, token, newPassword string) error {
	return m.resetErr
}
func (m *mockAuthService) Logout(ctx context.Context, userID uuid.UUID, tokenJTI string) error {
	return m.logoutErr
}
func (m *mockAuthService) SwitchTenant(ctx context.Context, userID uuid.UUID, targetTenantID uuid.UUID, ip, ua string) (*models.TenantSwitchResponse, error) {
	return m.switchTenantResp, m.switchTenantErr
}
func (m *mockAuthService) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]models.TenantInfo, error) {
	return m.tenants, m.tenantsErr
}
func (m *mockAuthService) VerifyEmail(ctx context.Context, token string) error { return nil }
func (m *mockAuthService) ResetPasswordForced(ctx context.Context, userID uuid.UUID, newPassword string) error {
	return m.resetErr
}

type mockPasswordService struct {
	validateErr error
	changeErr   error
	resetErr    error
	policy      models.PasswordPolicyResponse
}

func (m *mockPasswordService) Validate(ctx context.Context, userID uuid.UUID, password string) error {
	return m.validateErr
}
func (m *mockPasswordService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	return m.changeErr
}
func (m *mockPasswordService) ResetPassword(ctx context.Context, token, newPassword string) error {
	return m.resetErr
}
func (m *mockPasswordService) GetPolicy() models.PasswordPolicyResponse {
	return m.policy
}

type mockSessionService struct {
	sessions  []models.Session
	sessErr   error
	revokeErr error
}

func (m *mockSessionService) ListSessions(ctx context.Context, userID, tenantID uuid.UUID) ([]models.Session, error) {
	return m.sessions, m.sessErr
}
func (m *mockSessionService) RevokeSession(ctx context.Context, userID, tenantID uuid.UUID, sessionID string) error {
	return m.revokeErr
}
func (m *mockSessionService) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	return m.revokeErr
}

type mockUserService struct {
	usersResp   *models.ListUsersResponse
	usersErr    error
	userResp    *models.User
	userErr     error
	statusErr   error
	deleteErr   error
	auditResp   []models.AuditLog
	auditErr    error
}

func (m *mockUserService) ListUsers(ctx context.Context, params models.ListUsersParams) (*models.ListUsersResponse, error) {
	return m.usersResp, m.usersErr
}
func (m *mockUserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return m.userResp, m.userErr
}
func (m *mockUserService) UpdateUserStatus(ctx context.Context, userID, tenantID uuid.UUID, status string, actorID uuid.UUID) error {
	return m.statusErr
}
func (m *mockUserService) UpdateUserPreferences(ctx context.Context, userID, tenantID uuid.UUID, sidebarCollapsed *bool, language *string, actorID uuid.UUID) error {
	return nil
}
func (m *mockUserService) DeleteUser(ctx context.Context, userID, tenantID uuid.UUID, actorID uuid.UUID) error {
	return m.deleteErr
}
func (m *mockUserService) GetUserAuditLog(ctx context.Context, userID uuid.UUID, limit int) ([]models.AuditLog, error) {
	return m.auditResp, m.auditErr
}
func (m *mockUserService) CreateUser(ctx context.Context, tenantID uuid.UUID, req models.CreateUserRequest, actorID uuid.UUID) (*models.User, error) {
	return m.userResp, m.userErr
}

type mockInvitationService struct {
	inviteResp  *models.Invitation
	inviteErr   error
	acceptResp  *models.User
	acceptErr   error
	resendErr   error
	revokeErr   error
	validateResp *models.Invitation
	validateErr error
}

func (m *mockInvitationService) CreateInvite(ctx context.Context, tenantID, invitedBy uuid.UUID, email string, roleIDs []uuid.UUID, message string) (*models.Invitation, error) {
	return m.inviteResp, m.inviteErr
}
func (m *mockInvitationService) AcceptInvite(ctx context.Context, token, name, password string) (*models.User, error) {
	return m.acceptResp, m.acceptErr
}
func (m *mockInvitationService) ResendInvite(ctx context.Context, inviteID, actorID uuid.UUID) error {
	return m.resendErr
}
func (m *mockInvitationService) RevokeInvite(ctx context.Context, inviteID, actorID uuid.UUID) error {
	return m.revokeErr
}
func (m *mockInvitationService) ValidateInviteToken(ctx context.Context, token string) (*models.Invitation, error) {
	return m.validateResp, m.validateErr
}

type mockRBACService struct {
	roles    []models.Role
	rolesErr error
	roleResp *models.Role
	roleErr  error
	matrix   []models.RolePermissionMatrixEntry
	matrixErr error
	perms    []models.Permission
	permsErr error
	userRoles []models.Role
	userRolesErr error
	userPerms []string
	userPermsErr error
	createResp *models.Role
	createErr error
	updateErr error
	cloneResp *models.Role
	cloneErr error
	deleteErr error
	assignErr error
	bulkResp *service.BulkAssignResult
	unassignErr error
}

func (m *mockRBACService) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]models.Role, error) {
	return m.roles, m.rolesErr
}
func (m *mockRBACService) GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	return m.roleResp, m.roleErr
}
func (m *mockRBACService) CreateRole(ctx context.Context, tenantID uuid.UUID, name, description string, perms []string, createdBy uuid.UUID) (*models.Role, error) {
	return m.createResp, m.createErr
}
func (m *mockRBACService) UpdateRole(ctx context.Context, id, tenantID uuid.UUID, name, description string) error {
	return m.updateErr
}
func (m *mockRBACService) UpdateRolePermissions(ctx context.Context, id, tenantID uuid.UUID, perms []string) error { return nil }
func (m *mockRBACService) CloneRole(ctx context.Context, sourceID, tenantID uuid.UUID, name, description string, createdBy uuid.UUID) (*models.Role, error) {
	return m.cloneResp, m.cloneErr
}
func (m *mockRBACService) DeleteRole(ctx context.Context, id, tenantID uuid.UUID) error {
	return m.deleteErr
}
func (m *mockRBACService) GetPermissionMatrix(ctx context.Context, tenantID uuid.UUID) ([]models.RolePermissionMatrixEntry, error) {
	return m.matrix, m.matrixErr
}
func (m *mockRBACService) AssignRole(ctx context.Context, userID, roleID, tenantID, assignedBy uuid.UUID) error {
	return m.assignErr
}
func (m *mockRBACService) BulkAssignRoles(ctx context.Context, userIDs, roleIDs []uuid.UUID, tenantID, assignedBy uuid.UUID) *service.BulkAssignResult {
	return m.bulkResp
}
func (m *mockRBACService) UnassignRole(ctx context.Context, userID, roleID, tenantID uuid.UUID) error {
	return m.unassignErr
}
func (m *mockRBACService) GetUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]models.Role, error) {
	return m.userRoles, m.userRolesErr
}
func (m *mockRBACService) GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
	return m.userPerms, m.userPermsErr
}
func (m *mockRBACService) ListAllPermissions(ctx context.Context) ([]models.Permission, error) {
	return m.perms, m.permsErr
}

type mockMFAService struct {
	setupErr   error
	verifyErr  error
	codes      []string
	codesErr   error
	disableErr error
}

func (m *mockMFAService) SetupTOTP(ctx context.Context, userID uuid.UUID) (string, string, error) {
	return "JBSWY3DPEHPK3PXP", "otpauth://totp/FlyFleet:test?secret=JBSWY3DPEHPK3PXP", m.setupErr
}
func (m *mockMFAService) VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	return m.verifyErr
}
func (m *mockMFAService) GenerateBackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return m.codes, m.codesErr
}
func (m *mockMFAService) DisableMFA(ctx context.Context, userID uuid.UUID, password string) error {
	return m.disableErr
}

type mockUserRepo struct {
	userResp  *models.User
	userErr   error
	updateErr error
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return m.userResp, m.userErr
}
func (m *mockUserRepo) UpdateProfile(ctx context.Context, id uuid.UUID, name, phone, avatarURL string) error {
	return m.updateErr
}
func (m *mockUserRepo) UpdatePreferences(ctx context.Context, id, tenantID uuid.UUID, sidebarCollapsed *bool, language *string) error {
	return nil
}
func (m *mockUserRepo) GetTenants(ctx context.Context, id uuid.UUID) ([]models.TenantInfo, error) {
	return nil, nil
}

// ─── Helpers ───

func newTestAuthHandler(authSvc authService, pwSvc passwordService) *AuthHandler {
	return NewAuthHandler(authSvc, pwSvc, zap.NewNop())
}

func newTestSessionHandler(svc sessionService) *SessionHandler {
	return NewSessionHandler(svc, zap.NewNop())
}

func newTestUserHandler(svc userService) *UserHandler {
	return NewUserHandler(svc, zap.NewNop())
}

func newTestInviteHandler(svc invitationService) *InviteHandler {
	return NewInviteHandler(svc, zap.NewNop())
}

func newTestRBACHandler(svc rbacService) *RBACHandler {
	return NewRBACHandler(svc, zap.NewNop())
}

func newTestMFAHandler(svc mfaService) *MFAHandler {
	return NewMFAHandler(svc, zap.NewNop())
}

func newTestProfileHandler(repo userRepo) *ProfileHandler {
	return NewProfileHandler(repo, zap.NewNop())
}

func req(method, url string, body string) *http.Request {
	r := httptest.NewRequest(method, url, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func decodeResp(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	err := json.NewDecoder(w.Body).Decode(v)
	require.NoError(t, err)
}

// ─── Auth Handler Tests ───

func TestAuthHandler_Identify_Success(t *testing.T) {
	mockAuth := &mockAuthService{
		identifyResp: &models.IdentifyResponse{
			Exists:        true,
			EmailVerified: true,
			Tenants:       []models.TenantInfo{{Name: "Test Org", Role: "admin"}},
		},
	}
	h := newTestAuthHandler(mockAuth, nil)

	w := httptest.NewRecorder()
	h.Identify(w, req("POST", "/auth/identify", `{"email":"user@example.com"}`))

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.IdentifyResponse
	decodeResp(t, w, &resp)
	assert.True(t, resp.Exists)
	assert.Len(t, resp.Tenants, 1)
}

func TestAuthHandler_Identify_MissingEmail_400(t *testing.T) {
	h := newTestAuthHandler(&mockAuthService{}, nil)

	w := httptest.NewRecorder()
	h.Identify(w, req("POST", "/auth/identify", `{"email":""}`))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Identify_InvalidJSON_400(t *testing.T) {
	h := newTestAuthHandler(&mockAuthService{}, nil)

	w := httptest.NewRecorder()
	h.Identify(w, httptest.NewRequest("POST", "/auth/identify", strings.NewReader(`{bad`)))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Identify_ServiceError_500(t *testing.T) {
	mockAuth := &mockAuthService{identifyErr: assert.AnError}
	h := newTestAuthHandler(mockAuth, nil)

	w := httptest.NewRecorder()
	h.Identify(w, req("POST", "/auth/identify", `{"email":"user@example.com"}`))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthHandler_Login_Success_200(t *testing.T) {
	mockAuth := &mockAuthService{
		loginResp: &models.AuthTokensResponse{
			AccessToken:  "eyJ.access.token",
			RefreshToken: "refresh-token-123",
			TokenType:    "Bearer",
			ExpiresIn:    900,
		},
	}
	h := newTestAuthHandler(mockAuth, nil)

	w := httptest.NewRecorder()
	h.Login(w, req("POST", "/auth/login", `{"email":"user@example.com","password":"secret"}`))

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.AuthTokensResponse
	decodeResp(t, w, &resp)
	assert.Equal(t, "eyJ.access.token", resp.AccessToken)
	assert.Equal(t, 900, resp.ExpiresIn)
}

func TestAuthHandler_Login_MFARequired_202(t *testing.T) {
	mockAuth := &mockAuthService{
		loginMFAResp: &models.MFARequiredResponse{
			MFARequired: true,
			MFAMethods:  []string{"totp", "sms"},
		},
	}
	h := newTestAuthHandler(mockAuth, nil)

	w := httptest.NewRecorder()
	h.Login(w, req("POST", "/auth/login", `{"email":"user@example.com","password":"secret"}`))

	assert.Equal(t, http.StatusAccepted, w.Code)
	var resp models.MFARequiredResponse
	decodeResp(t, w, &resp)
	assert.True(t, resp.MFARequired)
}

func TestAuthHandler_Login_InvalidCredentials_401(t *testing.T) {
	mockAuth := &mockAuthService{loginErr: service.ErrInvalidCredentials}
	h := newTestAuthHandler(mockAuth, nil)

	w := httptest.NewRecorder()
	h.Login(w, req("POST", "/auth/login", `{"email":"user@example.com","password":"wrong"}`))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var errResp struct{ Error string }
	decodeResp(t, w, &errResp)
	assert.Equal(t, "invalid_credentials", errResp.Error)
}

func TestAuthHandler_Login_AccountLocked_403(t *testing.T) {
	mockAuth := &mockAuthService{loginErr: service.ErrAccountLocked}
	h := newTestAuthHandler(mockAuth, nil)

	w := httptest.NewRecorder()
	h.Login(w, req("POST", "/auth/login", `{"email":"user@example.com","password":"wrong"}`))

	assert.Equal(t, http.StatusForbidden, w.Code)
	var errResp struct{ Error string }
	decodeResp(t, w, &errResp)
	assert.Equal(t, "account_locked", errResp.Error)
}

func TestAuthHandler_Login_MissingCredentials_400(t *testing.T) {
	h := newTestAuthHandler(&mockAuthService{}, nil)

	w := httptest.NewRecorder()
	h.Login(w, req("POST", "/auth/login", `{"email":""}`))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_UserNotFound_401(t *testing.T) {
	mockAuth := &mockAuthService{loginErr: service.ErrUserNotFound}
	h := newTestAuthHandler(mockAuth, nil)

	w := httptest.NewRecorder()
	h.Login(w, req("POST", "/auth/login", `{"email":"nonexistent@example.com","password":"x"}`))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Login_InternalError_500(t *testing.T) {
	mockAuth := &mockAuthService{loginErr: assert.AnError}
	h := newTestAuthHandler(mockAuth, nil)

	w := httptest.NewRecorder()
	h.Login(w, req("POST", "/auth/login", `{"email":"user@example.com","password":"x"}`))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthHandler_ForgotPassword_AlwaysReturns200(t *testing.T) {
	mockAuth := &mockAuthService{}
	mockPw := &mockPasswordService{}
	h := newTestAuthHandler(mockAuth, mockPw)

	// Success case
	w := httptest.NewRecorder()
	h.ForgotPassword(w, req("POST", "/auth/forgot-password", `{"email":"user@example.com"}`))
	assert.Equal(t, http.StatusOK, w.Code)

	// Error case — still returns 200 to prevent email enumeration
	mockAuth.forgotErr = assert.AnError
	w2 := httptest.NewRecorder()
	h.ForgotPassword(w2, req("POST", "/auth/forgot-password", `{"email":"nonexistent@example.com"}`))
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestAuthHandler_ResetPassword_Success_200(t *testing.T) {
	h := newTestAuthHandler(&mockAuthService{}, &mockPasswordService{})

	w := httptest.NewRecorder()
	h.ResetPassword(w, req("POST", "/auth/reset-password", `{"token":"abc123","new_password":"New-Password-123!"}`))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_GetPolicy_Success(t *testing.T) {
	mockPw := &mockPasswordService{
		policy: models.PasswordPolicyResponse{
			MinLength: 12, RequireUpper: true, RequireLower: true,
			RequireDigit: true, RequireSpecial: true, HistoryLength: 5,
		},
	}
	h := newTestAuthHandler(&mockAuthService{}, mockPw)

	w := httptest.NewRecorder()
	h.GetPolicy(w, req("GET", "/auth/password-policy", ""))

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.PasswordPolicyResponse
	decodeResp(t, w, &resp)
	assert.Equal(t, 12, resp.MinLength)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	h := newTestAuthHandler(&mockAuthService{}, &mockPasswordService{})

	userID := uuid.New().String()
	r := req("POST", "/auth/logout", "")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyUserID, userID))
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyTokenJTI, "jti-456"))

	w := httptest.NewRecorder()
	h.Logout(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ─── Session Handler Tests ───

func TestSessionHandler_ListSessions_Success(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	mockSvc := &mockSessionService{
		sessions: []models.Session{
			{ID: uuid.New().String(), UserID: userID, Device: "browser", LastActive: now.Add(-2 * time.Minute)},
			{ID: uuid.New().String(), UserID: userID, Device: "mobile", LastActive: now.Add(-1 * time.Hour)},
		},
	}
	h := newTestSessionHandler(mockSvc)

	r := req("GET", "/auth/sessions", "")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyUserID, userID.String()))

	w := httptest.NewRecorder()
	h.ListSessions(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []models.Session
	decodeResp(t, w, &resp)
	assert.Len(t, resp, 2)
}

func TestSessionHandler_ListSessions_ServiceError_500(t *testing.T) {
	mockSvc := &mockSessionService{sessErr: assert.AnError}
	h := newTestSessionHandler(mockSvc)

	r := req("GET", "/auth/sessions", "")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyUserID, uuid.New().String()))

	w := httptest.NewRecorder()
	h.ListSessions(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSessionHandler_RevokeSession_Success(t *testing.T) {
	mockSvc := &mockSessionService{}
	h := newTestSessionHandler(mockSvc)

	r := httptest.NewRequest("DELETE", "/auth/sessions/"+uuid.New().String(), nil)
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyUserID, uuid.New().String()))
	w := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Delete("/auth/sessions/{id}", h.RevokeSession)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSessionHandler_RevokeAllSessions_Success(t *testing.T) {
	mockSvc := &mockSessionService{}
	h := newTestSessionHandler(mockSvc)

	r := req("DELETE", "/auth/sessions/all", "")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyUserID, uuid.New().String()))

	w := httptest.NewRecorder()
	h.RevokeAllSessions(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ─── User Handler Tests ───

func TestUserHandler_ListUsers_Success(t *testing.T) {
	mockSvc := &mockUserService{
		usersResp: &models.ListUsersResponse{
			Users: []models.UserListItem{
				{Email: "alice@example.com", Name: "Alice", Status: "active", Roles: []string{"admin"}},
				{Email: "bob@example.com", Name: "Bob", Status: "active", Roles: []string{"dispatcher"}},
			},
			Total:    2,
			Page:     1,
			PageSize: 20,
		},
	}
	h := newTestUserHandler(mockSvc)

	r := req("GET", "/admin/users?page=1&page_size=20", "")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyTenantID, uuid.New().String()))

	w := httptest.NewRecorder()
	h.ListUsers(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ListUsersResponse
	decodeResp(t, w, &resp)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Users, 2)
}

func TestUserHandler_ListUsers_WithFilters(t *testing.T) {
	mockSvc := &mockUserService{
		usersResp: &models.ListUsersResponse{Users: []models.UserListItem{}, Total: 0},
	}
	h := newTestUserHandler(mockSvc)

	r := req("GET", "/admin/users?page=2&page_size=50&search=john&status=active", "")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyTenantID, uuid.New().String()))

	w := httptest.NewRecorder()
	h.ListUsers(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_UpdateStatus_Success(t *testing.T) {
	mockSvc := &mockUserService{}
	h := newTestUserHandler(mockSvc)

	userID := uuid.New()
	body := `{"status":"deactivated","reason":"Policy violation"}`
	r := httptest.NewRequest("PUT", "/admin/users/"+userID.String()+"/status", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyUserID, uuid.New().String()))

	w := httptest.NewRecorder()
	router := chi.NewRouter()
	router.Put("/admin/users/{id}/status", h.UpdateStatus)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_GetUser_Success(t *testing.T) {
	userID := uuid.New()
	mockSvc := &mockUserService{
		userResp: &models.User{ID: userID, Email: "user@example.com", Name: "Test User", Status: "active"},
	}
	h := newTestUserHandler(mockSvc)

	r := httptest.NewRequest("GET", "/admin/users/"+userID.String(), nil)
	w := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Get("/admin/users/{id}", h.GetUser)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ─── Invitation Handler Tests ───

func TestInviteHandler_CreateInvite_Success(t *testing.T) {
	mockSvc := &mockInvitationService{
		inviteResp: &models.Invitation{
			Email:     "new@example.com",
			Status:    "pending",
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		},
	}
	h := newTestInviteHandler(mockSvc)

	roleID := uuid.New()
	tenantID := uuid.New()
	actorID := uuid.New()
	body := `{"email":"new@example.com","role_ids":["` + roleID.String() + `"],"message":"Welcome!"}`
	r := httptest.NewRequest("POST", "/admin/users/invite", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyTenantID, tenantID.String()))
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyUserID, actorID.String()))

	w := httptest.NewRecorder()
	h.CreateInvite(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestInviteHandler_ValidateToken_Success(t *testing.T) {
	mockSvc := &mockInvitationService{
		validateResp: &models.Invitation{Email: "new@example.com", Status: "pending"},
	}
	h := newTestInviteHandler(mockSvc)

	r := httptest.NewRequest("GET", "/auth/invite/abc123", nil)
	w := httptest.NewRecorder()

	// Use chi router to properly set URL params
	router := chi.NewRouter()
	router.Get("/auth/invite/{token}", h.ValidateInviteToken)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ─── RBAC Handler Tests ───

func TestRBACHandler_ListRoles_Success(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockRBACService{
		roles: []models.Role{
			{Name: "org_admin", IsSystem: true},
			{Name: "dispatcher", IsSystem: false},
		},
	}
	h := newTestRBACHandler(mockSvc)

	r := req("GET", "/admin/rbac/roles", "")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyTenantID, tenantID.String()))

	w := httptest.NewRecorder()
	h.ListRoles(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []models.Role
	decodeResp(t, w, &resp)
	assert.Len(t, resp, 2)
}

func TestRBACHandler_CloneRole_Success(t *testing.T) {
	mockSvc := &mockRBACService{
		cloneResp: &models.Role{Name: "dispatcher_copy", IsSystem: false},
	}
	h := newTestRBACHandler(mockSvc)

	sourceID := uuid.New()
	body := `{"name":"dispatcher_copy","description":"Copy"}`
	r := httptest.NewRequest("POST", "/admin/rbac/roles/"+sourceID.String()+"/clone", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyTenantID, uuid.New().String()))
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyUserID, uuid.New().String()))
	w := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Post("/admin/rbac/roles/{id}/clone", h.CloneRole)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp models.Role
	decodeResp(t, w, &resp)
	assert.Equal(t, "dispatcher_copy", resp.Name)
}

func TestRBACHandler_GetPermissionMatrix_Success(t *testing.T) {
	mockSvc := &mockRBACService{
		matrix: []models.RolePermissionMatrixEntry{
			{Name: "admin", Permissions: map[string]map[string]bool{"vehicle": {"read": true, "write": true}}},
		},
	}
	h := newTestRBACHandler(mockSvc)

	r := req("GET", "/admin/rbac/matrix", "")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyTenantID, uuid.New().String()))

	w := httptest.NewRecorder()
	h.GetPermissionMatrix(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ─── Profile Handler Tests ───

func TestProfileHandler_GetProfile_Success(t *testing.T) {
	userID := uuid.New()
	mockRepo := &mockUserRepo{
		userResp: &models.User{
			ID:    userID,
			Email: "user@example.com",
			Name:  "Test User",
		},
	}
	h := newTestProfileHandler(mockRepo)

	r := req("GET", "/auth/profile", "")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyUserID, userID.String()))

	w := httptest.NewRecorder()
	h.GetProfile(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	decodeResp(t, w, &resp)
	assert.Equal(t, "user@example.com", resp.Email)
	assert.Equal(t, "Test User", resp.Name)
}

func TestProfileHandler_UpdateProfile_Success(t *testing.T) {
	mockRepo := &mockUserRepo{}
	h := newTestProfileHandler(mockRepo)

	r := req("PUT", "/auth/profile", `{"name":"Updated Name","phone":"+91 98765 43210"}`)
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyUserID, uuid.New().String()))

	w := httptest.NewRecorder()
	h.UpdateProfile(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ─── MFA Handler Tests ───

func TestMFAHandler_SetupTOTP_Success(t *testing.T) {
	mockSvc := &mockMFAService{}
	h := newTestMFAHandler(mockSvc)

	r := req("POST", "/auth/mfa/setup/totp", "")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyUserID, uuid.New().String()))

	w := httptest.NewRecorder()
	h.SetupTOTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMFAHandler_GetBackupCodes_Success(t *testing.T) {
	mockSvc := &mockMFAService{
		codes: []string{"12345678", "23456789", "34567890"},
	}
	h := newTestMFAHandler(mockSvc)

	r := req("GET", "/auth/mfa/backup-codes", "")
	r = r.WithContext(context.WithValue(r.Context(), CtxKeyUserID, uuid.New().String()))

	w := httptest.NewRecorder()
	h.GetBackupCodes(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}
