package auth

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Login metrics
	LoginAttemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_login_attempts_total",
		Help: "Total number of login attempts",
	}, []string{"tenant_id", "method", "result"})

	LoginSuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_login_success_total",
		Help: "Total number of successful logins",
	}, []string{"tenant_id", "method", "user_type"})

	LoginDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "auth_login_duration_seconds",
		Help:    "Login flow duration in seconds",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 8), // 100ms to ~25s
	}, []string{"tenant_id", "method"})

	// JWT metrics
	TokenValidationsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auth_token_validations_total",
		Help: "Total JWT token validations",
	})

	TokenValidationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "auth_token_validation_duration_seconds",
		Help:    "JWT token validation latency",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
	}, []string{"tenant_id"})

	TokenRefreshesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_token_refreshes_total",
		Help: "Total token refresh attempts",
	}, []string{"tenant_id", "result"})

	// MFA metrics
	MFAAttemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_mfa_attempts_total",
		Help: "Total MFA verification attempts",
	}, []string{"tenant_id", "method", "result"})

	MFASuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_mfa_success_total",
		Help: "Total successful MFA verifications",
	}, []string{"tenant_id", "method"})

	// Session metrics
	ActiveSessions = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "auth_active_sessions",
		Help: "Current number of active user sessions",
	}, []string{"tenant_id", "user_type"})

	SessionCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_sessions_created_total",
		Help: "Total sessions created",
	}, []string{"tenant_id", "user_type"})

	SessionExpiredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_sessions_expired_total",
		Help: "Total sessions that expired naturally",
	}, []string{"tenant_id"})

	SessionRevokedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_sessions_revoked_total",
		Help: "Total sessions revoked (logout, admin action, security)",
	}, []string{"tenant_id", "reason"})

	// RBAC metrics
	PermissionChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_permission_checks_total",
		Help: "Total permission/access control checks",
	}, []string{"tenant_id", "permission", "result"})

	RoleAssignmentsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_role_assignments_total",
		Help: "Total role assignments to users",
	}, []string{"tenant_id", "role", "action"})

	// Rate limiting metrics
	RateLimitHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_rate_limit_hits_total",
		Help: "Total rate limit enforcement hits",
	}, []string{"tenant_id", "endpoint", "limit_type"})

	RateLimitExceededTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_rate_limit_exceeded_total",
		Help: "Total requests blocked due to rate limits",
	}, []string{"tenant_id", "endpoint", "limit_type"})

	// Security metrics
	FailedLoginRate = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_failed_login_attempts_total",
		Help: "Total failed login attempts (for brute force detection)",
	}, []string{"tenant_id", "username_hash", "reason"})

	AccountLockoutsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_account_lockouts_total",
		Help: "Total accounts locked due to security policy",
	}, []string{"tenant_id", "reason"})

	SuspiciousActivityDetected = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_suspicious_activity_total",
		Help: "Total suspicious authentication patterns detected",
	}, []string{"tenant_id", "activity_type", "severity"})

	// Password reset metrics
	PasswordResetRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_password_reset_requests_total",
		Help: "Total password reset requests",
	}, []string{"tenant_id", "channel"})

	PasswordResetSuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_password_reset_success_total",
		Help: "Total successful password resets",
	}, []string{"tenant_id"})

	// Registration metrics
	RegistrationAttemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_registration_attempts_total",
		Help: "Total user registration attempts",
	}, []string{"tenant_id", "user_type", "result"})

	RegistrationSuccessTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_registration_success_total",
		Help: "Total successful user registrations",
	}, []string{"tenant_id", "user_type"})

	// API key metrics
	APIKeyCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_api_keys_created_total",
		Help: "Total API keys created",
	}, []string{"tenant_id", "user_type"})

	APIKeyRevokedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_api_keys_revoked_total",
		Help: "Total API keys revoked",
	}, []string{"tenant_id", "reason"})
)

// Global metrics instance
var GlobalMetrics = struct {
	LoginAttemptsTotal         *prometheus.CounterVec
	LoginSuccessTotal          *prometheus.CounterVec
	LoginDuration              *prometheus.HistogramVec
	TokenValidationsTotal      prometheus.Counter
	TokenValidationDuration    *prometheus.HistogramVec
	TokenRefreshesTotal        *prometheus.CounterVec
	MFAAttemptsTotal           *prometheus.CounterVec
	MFASuccessTotal            *prometheus.CounterVec
	ActiveSessions             *prometheus.GaugeVec
	SessionCreatedTotal        *prometheus.CounterVec
	SessionExpiredTotal        *prometheus.CounterVec
	SessionRevokedTotal        *prometheus.CounterVec
	PermissionChecksTotal      *prometheus.CounterVec
	RoleAssignmentsTotal       *prometheus.CounterVec
	RateLimitHitsTotal         *prometheus.CounterVec
	RateLimitExceededTotal     *prometheus.CounterVec
	FailedLoginRate            *prometheus.CounterVec
	AccountLockoutsTotal       *prometheus.CounterVec
	SuspiciousActivityDetected *prometheus.CounterVec
	PasswordResetRequestsTotal *prometheus.CounterVec
	PasswordResetSuccessTotal  *prometheus.CounterVec
	RegistrationAttemptsTotal  *prometheus.CounterVec
	RegistrationSuccessTotal   *prometheus.CounterVec
	APIKeyCreatedTotal         *prometheus.CounterVec
	APIKeyRevokedTotal         *prometheus.CounterVec
}{
	LoginAttemptsTotal:         LoginAttemptsTotal,
	LoginSuccessTotal:          LoginSuccessTotal,
	LoginDuration:              LoginDuration,
	TokenValidationsTotal:      TokenValidationsTotal,
	TokenValidationDuration:    TokenValidationDuration,
	TokenRefreshesTotal:        TokenRefreshesTotal,
	MFAAttemptsTotal:           MFAAttemptsTotal,
	MFASuccessTotal:            MFASuccessTotal,
	ActiveSessions:             ActiveSessions,
	SessionCreatedTotal:        SessionCreatedTotal,
	SessionExpiredTotal:        SessionExpiredTotal,
	SessionRevokedTotal:        SessionRevokedTotal,
	PermissionChecksTotal:      PermissionChecksTotal,
	RoleAssignmentsTotal:       RoleAssignmentsTotal,
	RateLimitHitsTotal:         RateLimitHitsTotal,
	RateLimitExceededTotal:     RateLimitExceededTotal,
	FailedLoginRate:            FailedLoginRate,
	AccountLockoutsTotal:       AccountLockoutsTotal,
	SuspiciousActivityDetected: SuspiciousActivityDetected,
	PasswordResetRequestsTotal: PasswordResetRequestsTotal,
	PasswordResetSuccessTotal:  PasswordResetSuccessTotal,
	RegistrationAttemptsTotal:  RegistrationAttemptsTotal,
	RegistrationSuccessTotal:   RegistrationSuccessTotal,
	APIKeyCreatedTotal:         APIKeyCreatedTotal,
	APIKeyRevokedTotal:         APIKeyRevokedTotal,
}
