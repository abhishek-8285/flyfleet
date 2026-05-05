package models

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
	UserStatusLocked   UserStatus = "locked"
)

type User struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Email           string     `json:"email" db:"email" index:"unique"`
	Name            string     `json:"name" db:"name"`
	Phone           string     `json:"phone,omitempty" db:"phone"`
	AvatarURL       string     `json:"avatar_url,omitempty" db:"avatar_url"`
	PasswordHash    string     `json:"-" db:"password_hash"`
	MFAEnabled      bool       `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecret       string     `json:"-" db:"mfa_secret"`
	MFABackupCodes  []string   `json:"-" db:"mfa_backup_codes"`
	Status          UserStatus `json:"status" db:"status"`
	EmailVerified   bool       `json:"email_verified" db:"email_verified"`
	FailedLogins    int        `json:"-" db:"failed_logins"`
	LockedUntil     *time.Time `json:"-" db:"locked_until"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	LastLoginIP     string     `json:"-" db:"last_login_ip"`
	PermissionMask  int64      `json:"-" db:"permission_mask"`
	SidebarCollapsed bool      `json:"sidebar_collapsed" db:"sidebar_collapsed"`
	Language         string     `json:"language" db:"language"`
	MustResetPassword bool      `json:"must_reset_password" db:"must_reset_password"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}
