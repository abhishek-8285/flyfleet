package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Request parsing tests ───

func TestAuthHandler_Identify_InvalidJSON(t *testing.T) {
	body := bytes.NewBufferString(`{invalid json}`)
	var parsed models.IdentifyRequest
	err := json.NewDecoder(body).Decode(&parsed)
	require.Error(t, err)
}

func TestAuthHandler_Identify_MissingEmail(t *testing.T) {
	body := bytes.NewBufferString(`{"email":""}`)
	var parsed models.IdentifyRequest
	err := json.NewDecoder(body).Decode(&parsed)
	require.NoError(t, err)
	assert.Equal(t, "", parsed.Email)
}

func TestAuthHandler_Login_RequestParsing(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		expectEmail string
		expectPass  string
		expectTenantID string
		wantErr     bool
	}{
		{
			name:        "valid_login",
			body:        `{"email":"user@example.com","password":"secret123","tenant_id":"org-001"}`,
			expectEmail: "user@example.com",
			expectPass:  "secret123",
			expectTenantID: "org-001",
			wantErr:     false,
		},
		{
			name:        "missing_tenant_id",
			body:        `{"email":"user@example.com","password":"secret123"}`,
			expectEmail: "user@example.com",
			expectPass:  "secret123",
			expectTenantID: "",
			wantErr:     false,
		},
		{
			name:    "invalid_json",
			body:    `{bad`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parsed models.LoginRequest
			err := json.Unmarshal([]byte(tt.body), &parsed)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectEmail, parsed.Email)
			assert.Equal(t, tt.expectPass, parsed.Password)
			assert.Equal(t, tt.expectTenantID, parsed.TenantID)
		})
	}
}

func TestAuthHandler_Register_RequestParsing(t *testing.T) {
	body := `{"email":"new@example.com","password":"MyP@ssw0rd!","name":"New User","tenant_id":"org-001"}`

	var parsed struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		TenantID    string `json:"tenant_id"`
	}
	err := json.Unmarshal([]byte(body), &parsed)
	require.NoError(t, err)
	assert.Equal(t, "new@example.com", parsed.Email)
	assert.Equal(t, "MyP@ssw0rd!", parsed.Password)
	assert.Equal(t, "New User", parsed.Name)
	assert.Equal(t, "org-001", parsed.TenantID)
}

func TestAuthHandler_OrgSwitch_PathParsing(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/orgs/org-beta-002/switch", nil)
	assert.Contains(t, req.URL.Path, "org-beta-002")
}

func TestRBACHandler_CloneRole_RequestParsing(t *testing.T) {
	body := `{"name":"Fleet Manager (Copy)","description":"Cloned from Fleet Manager"}`

	var parsed models.CloneRoleRequest
	err := json.Unmarshal([]byte(body), &parsed)
	require.NoError(t, err)
	assert.Equal(t, "Fleet Manager (Copy)", parsed.Name)
	assert.Equal(t, "Cloned from Fleet Manager", parsed.Description)
}

func TestRBACHandler_BulkAssign_RequestParsing(t *testing.T) {
	body := `{
		"user_ids": ["550e8400-e29b-41d4-a716-446655440001","550e8400-e29b-41d4-a716-446655440002"],
		"role_ids": ["550e8400-e29b-41d4-a716-446655440003"],
		"tenant_id": "550e8400-e29b-41d4-a716-446655440004"
	}`

	var parsed struct {
		UserIDs []string `json:"user_ids"`
		RoleIDs []string `json:"role_ids"`
		TenantID   string   `json:"tenant_id"`
	}
	err := json.Unmarshal([]byte(body), &parsed)
	require.NoError(t, err)
	assert.Len(t, parsed.UserIDs, 2)
	assert.Len(t, parsed.RoleIDs, 1)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440004", parsed.TenantID)
}

func TestInviteHandler_CreateInvite_RequestParsing(t *testing.T) {
	body := `{"email":"new@example.com","role_ids":["550e8400-e29b-41d4-a716-446655440001"],"message":"Welcome!"}`

	var parsed models.InviteUserRequest
	err := json.Unmarshal([]byte(body), &parsed)
	require.NoError(t, err)
	assert.Equal(t, "new@example.com", parsed.Email)
	assert.Len(t, parsed.RoleIDs, 1)
	assert.Equal(t, "Welcome!", parsed.Message)
}

func TestInviteHandler_AcceptInvite_RequestParsing(t *testing.T) {
	body := `{"token":"abc123","name":"New User","password":"MyP@ssw0rd!"}`

	var parsed models.InviteAcceptRequest
	err := json.Unmarshal([]byte(body), &parsed)
	require.NoError(t, err)
	assert.Equal(t, "abc123", parsed.Token)
	assert.Equal(t, "New User", parsed.Name)
	assert.Equal(t, "MyP@ssw0rd!", parsed.Password)
}

func TestUserHandler_ListUsers_QueryParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?page=2&page_size=50&search=john&status=active", nil)

	page := req.URL.Query().Get("page")
	pageSize := req.URL.Query().Get("page_size")
	search := req.URL.Query().Get("search")
	status := req.URL.Query().Get("status")

	assert.Equal(t, "2", page)
	assert.Equal(t, "50", pageSize)
	assert.Equal(t, "john", search)
	assert.Equal(t, "active", status)
}

func TestUserHandler_UpdateStatus_RequestParsing(t *testing.T) {
	body := `{"status":"disabled","reason":"Policy violation"}`

	var parsed models.UpdateUserStatusRequest
	err := json.Unmarshal([]byte(body), &parsed)
	require.NoError(t, err)
	assert.Equal(t, "disabled", parsed.Status)
	assert.Equal(t, "Policy violation", parsed.Reason)
}

// ─── Response structure tests ───

func TestResponse_AuthTokensResponse(t *testing.T) {
	resp := models.AuthTokensResponse{
		AccessToken:  "eyJ...",
		RefreshToken: "dGhpcyBpcyBhIHRlc3Q...",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User: models.UserInfo{
			Email:       "user@example.com",
			Name:        "Test User",
			OrgName:     "Acme Corp",
			Permissions: []string{"vehicle:read", "driver:read"},
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var parsed models.AuthTokensResponse
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "eyJ...", parsed.AccessToken)
	assert.Equal(t, "user@example.com", parsed.User.Email)
	assert.Len(t, parsed.User.Permissions, 2)
}

func TestResponse_OrgSwitchResponse(t *testing.T) {
	resp := models.OrgSwitchResponse{
		Success:      true,
		AccessToken:  "new-token",
		RefreshToken: "new-refresh",
		ExpiresIn:    3600,
		Organization: models.OrgInfo{
			Name:         "Beta Fleet",
			Role:         "fleet_manager",
			VehicleCount: 12,
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var parsed models.OrgSwitchResponse
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.True(t, parsed.Success)
	assert.Equal(t, "Beta Fleet", parsed.Organization.Name)
	assert.Equal(t, 12, parsed.Organization.VehicleCount)
}

func TestResponse_IdentifyResponse(t *testing.T) {
	resp := models.IdentifyResponse{
		Exists:        true,
		EmailVerified: true,
		Organizations: []models.OrgInfo{
			{Name: "Alpha Fleet", Role: "admin", VehicleCount: 45},
			{Name: "Beta Fleet", Role: "fleet_manager", VehicleCount: 12},
		},
		MFARequired:   false,
		MFAMethods:    []string{"totp", "sms"},
		MFARemembered: false,
		AccountLocked: false,
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var parsed models.IdentifyResponse
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.True(t, parsed.Exists)
	assert.True(t, parsed.EmailVerified)
	assert.Len(t, parsed.Organizations, 2)
	assert.Equal(t, "Alpha Fleet", parsed.Organizations[0].Name)
	assert.Len(t, parsed.MFAMethods, 2)
}

func TestResponse_ListUsersResponse(t *testing.T) {
	resp := models.ListUsersResponse{
		Users: []models.UserListItem{
			{Email: "alice@example.com", Name: "Alice", Status: "active", Roles: []string{"admin"}},
			{Email: "bob@example.com", Name: "Bob", Status: "active", Roles: []string{"fleet_manager"}},
		},
		Total:    2,
		Page:     1,
		PageSize: 20,
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var parsed models.ListUsersResponse
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Users, 2)
	assert.Equal(t, "admin", parsed.Users[0].Roles[0])
}

func TestResponse_PasswordPolicyResponse(t *testing.T) {
	resp := models.PasswordPolicyResponse{
		MinLength:      12,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
		HistoryLength:  5,
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var parsed models.PasswordPolicyResponse
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, 12, parsed.MinLength)
	assert.True(t, parsed.RequireUpper)
	assert.Equal(t, 5, parsed.HistoryLength)
}

func TestResponse_ErrorResponse(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
		status  int
	}{
		{"invalid_credentials", "invalid_credentials", "Invalid email or password", http.StatusUnauthorized},
		{"account_locked", "account_locked", "Account is locked", http.StatusForbidden},
		{"missing_email", "missing_email", "Email is required", http.StatusBadRequest},
		{"role_not_found", "role_not_found", "Role not found", http.StatusNotFound},
		{"internal_error", "internal_error", "An internal error occurred", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := models.ErrorResponse{
				Error:   tt.code,
				Message: tt.message,
			}

			data, err := json.Marshal(resp)
			require.NoError(t, err)

			var parsed models.ErrorResponse
			err = json.Unmarshal(data, &parsed)
			require.NoError(t, err)
			assert.Equal(t, tt.code, parsed.Error)
			assert.Equal(t, tt.message, parsed.Message)
		})
	}
}
