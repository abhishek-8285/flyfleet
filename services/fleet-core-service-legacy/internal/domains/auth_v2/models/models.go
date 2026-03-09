package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID `json:"id"`
	OrgID              uuid.UUID `json:"org_id"`
	Email              string    `json:"email"`
	Role               string    `json:"role"`
	PasswordHash       string    `json:"-"`
	CreatedAt          time.Time `json:"created_at"`
	CreatedBy          uuid.UUID `json:"created_by"`
	UpdatedAt          time.Time `json:"updated_at"`
	UpdatedBy          uuid.UUID `json:"updated_by"`
	MfaEnabled         bool      `json:"mfa_enabled"`
	MfaMethods         []string  `json:"mfa_methods"`
	MfaSecret          string    `json:"-"`
	LastPasswordChange time.Time `json:"last_password_change"`
	FailedLoginAttempts int       `json:"-"`
	LockedUntil         *time.Time `json:"-"`
	Permissions         []string   `json:"permissions,omitempty"`
}

type Permission struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
}

type RolePermission struct {
	Role           string `json:"role"`
	PermissionName string `json:"permission_name"`
}

type SecurityEvent struct {
	UserID    uuid.UUID
	EventType string
	IPAddress string
	UserAgent string
	Metadata  string
}

type MfaRecoveryCode struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CodeHash  string    `json:"-"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginHistory struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Country   string    `json:"country"`
	Success   bool      `json:"success"`
	RiskScore float64   `json:"risk_score"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	ExpiresIn        int64  `json:"expires_in,omitempty"`
	User             *User    `json:"user,omitempty"`
	MfaRequired      bool     `json:"mfa_required,omitempty"`
	MfaSetupRequired bool     `json:"mfa_setup_required,omitempty"`
	TempToken        string   `json:"temp_token,omitempty"`
	Permissions      []string `json:"permissions,omitempty"`
	RecoveryCodes    []string `json:"recovery_codes,omitempty"`
}

type RefreshToken struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	Token      string     `json:"token"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	Revoked    bool       `json:"revoked"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	IPAddress  *string    `json:"ip_address,omitempty"`
	UserAgent  *string    `json:"user_agent,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type TokenClaims struct {
	UserID      string   `json:"user_id"`
	OrgID       string   `json:"org_id"`
	Role        string   `json:"role"`
	Type        string   `json:"type"` // "access" or "refresh"
	Permissions []string `json:"permissions,omitempty"`
}

type PasswordResetToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	Used      bool       `json:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

type PasswordResetRequest struct {
	Email string `json:"email"`
}

type PasswordResetConfirm struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type OidcProvider struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Issuer           string    `json:"issuer"`
	ClientID         string    `json:"client_id"`
	ClientSecretHash string    `json:"-"`
	Enabled          bool      `json:"enabled"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UserOidcLink struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	ProviderID     uuid.UUID `json:"provider_id"`
	ProviderUserID string    `json:"provider_user_id"`
	Email          string    `json:"email"`
	LastLoginAt    *time.Time `json:"last_login_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type SigningKey struct {
	KID           string     `json:"kid"`
	PrivateKeyEnc string     `json:"-"`
	PublicKey     string     `json:"public_key"`
	Status        string     `json:"status"` // active, previous, future
	Algorithm     string     `json:"algorithm"`
	CreatedAt     time.Time  `json:"created_at"`
	ActivatedAt   *time.Time `json:"activated_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}

type APIKey struct {
	ID           uuid.UUID `json:"id"`
	OrgID        uuid.UUID `json:"org_id"`
	UserID       uuid.UUID `json:"user_id"`
	Name         string    `json:"name"`
	Prefix       string    `json:"prefix"`
	SecretHash   string    `json:"-"`
	Permissions  []string  `json:"permissions"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	Revoked      bool      `json:"revoked"`
}

type Organization struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Status            string    `json:"status"` // trial, active, suspended, billing_hold, deleted
	MaxUsersLimit     int       `json:"max_users_limit"`
	MaxApiKeysLimit   int       `json:"max_api_keys_limit"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type OrganizationMembership struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type ServiceAccount struct {
	ID               uuid.UUID  `json:"id"`
	OrgID            uuid.UUID  `json:"org_id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	ClientID         string     `json:"client_id"`
	ClientSecretHash string     `json:"-"`
	Role             string     `json:"role"`
	Permissions      []string   `json:"permissions"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	LastUsedAt       *time.Time `json:"last_used_at,omitempty"`
	SecretHash       string     `json:"-"` // Deprecated or for legacy compatibility if any
}

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}
