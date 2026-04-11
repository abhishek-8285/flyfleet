package models

import "github.com/google/uuid"

// ──────────────────────────────────────────────
// Request DTOs
// ──────────────────────────────────────────────

type IdentifyRequest struct {
	Email string `json:"email"`
}

type LoginRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	OrgID          string `json:"org_id"`
	RememberDevice bool   `json:"remember_device"`
}

type MFAVerifyRequest struct {
	TempToken       string `json:"temp_token"`
	Code            string `json:"code"`
	Method          string `json:"method"`
	RememberDevice  bool   `json:"remember_device"`
}

type MFASendRequest struct {
	Email  string `json:"email"`
	Method string `json:"method"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type SSOLoginRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type InviteValidateRequest struct {
	Token string `json:"token"`
}

type InviteAcceptRequest struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type InviteUserRequest struct {
	Email    string      `json:"email"`
	RoleIDs  []uuid.UUID `json:"role_ids"`
	Message  string      `json:"message,omitempty"`
}

type BulkInviteRequest struct {
	CSV string `json:"csv"`
}

type UpdateUserStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type UpdateUserPreferencesRequest struct {
	SidebarCollapsed *bool   `json:"sidebar_collapsed,omitempty"`
	Language         *string `json:"language,omitempty"`
}

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CloneRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRolePermissionsRequest struct {
	Permissions []string `json:"permissions"` // resource:action format
}

type AssignRoleRequest struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
	OrgID  uuid.UUID `json:"org_id"`
}

type UnassignRoleRequest struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
	OrgID  uuid.UUID `json:"org_id"`
}

type SetupTOTPRequest struct {
}

type VerifyTOTPRequest struct {
	Code string `json:"code"`
}

type SetupSMSRequest struct {
	Phone string `json:"phone"`
}

type VerifySMSMFARequest struct {
	Code string `json:"code"`
}

type DisableMFARequest struct {
	Password string `json:"password"`
}

type SwitchOrgRequest struct {
	OrgID string `json:"org_id"`
}

type ProfileUpdateRequest struct {
	Name  string `json:"name,omitempty"`
	Phone string `json:"phone,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type ListUsersParams struct {
	Page     int
	PageSize int
	Search   string
	Status   string
	RoleID   *uuid.UUID
}

// ──────────────────────────────────────────────
// Response DTOs
// ──────────────────────────────────────────────

type IdentifyResponse struct {
	Exists          bool         `json:"exists"`
	EmailVerified   bool         `json:"email_verified"`
	Organizations   []OrgInfo    `json:"organizations"`
	MFARequired     bool         `json:"mfa_required"`
	MFAMethods      []string     `json:"mfa_methods"`
	MFARemembered   bool         `json:"mfa_remembered"`
	SSOProviders    []string     `json:"sso_providers"`
	AccountLocked   bool         `json:"account_locked"`
}

type AuthTokensResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int        `json:"expires_in"`
	User         UserInfo   `json:"user"`
}

type MFARequiredResponse struct {
	MFARequired bool     `json:"mfa_required"`
	TempToken   string   `json:"temp_token"`
	MFAMethods  []string `json:"mfa_methods"`
	ExpiresIn   int      `json:"expires_in"`
}

type UserInfo struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	Role         string    `json:"role"`
	OrgID        uuid.UUID `json:"org_id"`
	OrgName      string    `json:"org_name"`
	Permissions  []string  `json:"permissions"`
}

type ProfileResponse struct {
	ID            uuid.UUID   `json:"id"`
	Email         string      `json:"email"`
	Name          string      `json:"name"`
	Role          string      `json:"role"`
	Phone         string      `json:"phone,omitempty"`
	AvatarURL     string      `json:"avatar_url,omitempty"`
	MFAEnabled    bool        `json:"mfa_enabled"`
	EmailVerified bool        `json:"email_verified"`
	CurrentOrg    *OrgInfo    `json:"current_org,omitempty"`
	Organizations []OrgInfo   `json:"organizations"`
	Permissions   []string    `json:"permissions"`
	LastLoginAt   *string     `json:"last_login_at,omitempty"`
}

type OrgSwitchResponse struct {
	Success      bool     `json:"success"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"`
	Organization OrgInfo  `json:"organization"`
}

type InviteResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	ExpiresAt string    `json:"expires_at"`
	InvitedBy uuid.UUID `json:"invited_by"`
	InvitedAt string    `json:"invited_at"`
}

type RoleResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	IsSystem       bool      `json:"is_system"`
	PermissionMask int64     `json:"permission_mask"`
	CreatedAt      string    `json:"created_at"`
}

type PermissionMatrixResponse struct {
	Roles []RolePermissionMatrixEntry `json:"roles"`
}

type RolePermissionMatrixEntry struct {
	ID          uuid.UUID                        `json:"id"`
	Name        string                           `json:"name"`
	IsSystem    bool                             `json:"is_system"`
	Permissions map[string]map[string]bool       `json:"permissions"`
}

type ErrorResponse struct {
	Error     string      `json:"error"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

type ListUsersResponse struct {
	Users      []UserListItem `json:"users"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}

type UserListItem struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Roles       []string   `json:"roles"`
	LastLoginAt *string    `json:"last_login_at,omitempty"`
	CreatedAt   string     `json:"created_at"`
}

type TOTPSecretResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

type BackupCodesResponse struct {
	Codes []string `json:"codes"`
}

type PasswordPolicyResponse struct {
	MinLength      int      `json:"min_length"`
	RequireUpper   bool     `json:"require_upper"`
	RequireLower   bool     `json:"require_lower"`
	RequireDigit   bool     `json:"require_digit"`
	RequireSpecial bool     `json:"require_special"`
	HistoryLength  int      `json:"history_length"`
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Services  map[string]string `json:"services"`
	Timestamp string            `json:"timestamp"`
}
