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
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	Name            string     `json:"name"`
	Phone           string     `json:"phone,omitempty"`
	AvatarURL       string     `json:"avatar_url,omitempty"`
	PasswordHash    string     `json:"-"`
	MFAEnabled      bool       `json:"mfa_enabled"`
	MFASecret       string     `json:"-"`
	MFABackupCodes  []string   `json:"-"`
	Status          UserStatus `json:"status"`
	EmailVerified   bool       `json:"email_verified"`
	FailedLogins    int        `json:"-"`
	LockedUntil     *time.Time `json:"-"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP     string     `json:"-"`
	PermissionMask  int64      `json:"-"`
	SidebarCollapsed bool      `json:"sidebar_collapsed"`
	Language         string     `json:"language"`
	MustResetPassword bool      `json:"must_reset_password"`
	Roles            []string   `json:"roles"`
	CreatedAt       time.Time  `json:"created_at"`

	UpdatedAt       time.Time  `json:"updated_at"`
}
