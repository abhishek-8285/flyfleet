package config

import "time"

// SecurityConfig holds security-related configuration for the IAM domain.
// This is the type referenced by AuthService and PasswordService.
type SecurityConfig struct {
	JWTSecret             string
	JWTAccessTTL          time.Duration
	JWTRefreshTTL         time.Duration
	JWTActionTTL          time.Duration
	BcryptCost            int
	MaxLoginAttempts      int
	LockoutDuration       time.Duration
	PasswordHistoryLength int
}

// IAMConfig holds configuration for the IAM domain within fleet-core.
// This is a subset of the original auth-service config, merged into fleet-core's config system.
type IAMConfig struct {
	JWTSecret             string
	JWTAccessTTL          time.Duration
	JWTRefreshTTL         time.Duration
	JWTActionTTL          time.Duration
	BcryptCost            int
	MaxLoginAttempts      int
	LockoutDuration       time.Duration
	PasswordHistoryLength int
	Email                 EmailConfig
	RateLimit             RateLimitConfig
}

type EmailConfig struct {
	Provider       string
	SendGridAPIKey string
	SMTPHost       string
	SMTPPort       int
	SMTPUser       string
	SMTPPassword   string
	FromAddress    string
	FromName       string
}

type RateLimitConfig struct {
	LoginPerIP    string
	ResetPerEmail string
	ResetPerIP    string
	IdentifyPerIP string
}

// SecurityConfig returns the security subset of IAMConfig.
// Used by AuthService and PasswordService which expect config.SecurityConfig.
func (c IAMConfig) SecurityConfig() SecurityConfig {
	return SecurityConfig{
		JWTSecret:             c.JWTSecret,
		JWTAccessTTL:          c.JWTAccessTTL,
		JWTRefreshTTL:         c.JWTRefreshTTL,
		JWTActionTTL:          c.JWTActionTTL,
		BcryptCost:            c.BcryptCost,
		MaxLoginAttempts:      c.MaxLoginAttempts,
		LockoutDuration:       c.LockoutDuration,
		PasswordHistoryLength: c.PasswordHistoryLength,
	}
}

// DefaultIAMConfig returns sensible defaults
func DefaultIAMConfig() IAMConfig {
	return IAMConfig{
		JWTSecret:             "",
		JWTAccessTTL:          15 * time.Minute,
		JWTRefreshTTL:         7 * 24 * time.Hour,
		JWTActionTTL:          30 * time.Minute,
		BcryptCost:            12,
		MaxLoginAttempts:      5,
		LockoutDuration:       15 * time.Minute,
		PasswordHistoryLength: 5,
		Email: EmailConfig{
			Provider: "noop",
		},
		RateLimit: RateLimitConfig{
			LoginPerIP:    "10/15m",
			ResetPerEmail: "3/1h",
			ResetPerIP:    "5/1h",
			IdentifyPerIP: "20/15m",
		},
	}
}
