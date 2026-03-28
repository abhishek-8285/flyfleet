package security

import (
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TenantValidator provides reusable tenant ownership validation
type TenantValidator struct {
	logger *zap.Logger
}

func NewTenantValidator(logger *zap.Logger) *TenantValidator {
	return &TenantValidator{logger: logger}
}

// GetTenantID extracts and parses tenant ID from request context
func (v *TenantValidator) GetTenantID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(ctxkeys.TenantIDKey); val != nil {
		if s, ok := val.(string); ok {
			if id, err := uuid.Parse(s); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

// RequireTenant returns error if tenant context is missing
func (v *TenantValidator) RequireTenant(r *http.Request) error {
	tenantID := v.GetTenantID(r)
	if tenantID == uuid.Nil {
		v.logger.Warn("missing tenant context",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method))
		return &SecurityError{Message: "Organization context missing", StatusCode: http.StatusUnauthorized}
	}
	return nil
}

// RequireRole checks if user has required role
func (v *TenantValidator) RequireRole(r *http.Request, allowedRoles ...string) error {
	roleVal := r.Context().Value(ctxkeys.RoleKey)
	if roleVal == nil {
		return &SecurityError{Message: "Access denied", StatusCode: http.StatusForbidden}
	}
	role, ok := roleVal.(string)
	if !ok {
		return &SecurityError{Message: "Access denied", StatusCode: http.StatusForbidden}
	}
	for _, allowed := range allowedRoles {
		if role == allowed {
			return nil
		}
	}
	return &SecurityError{Message: "Access denied", StatusCode: http.StatusForbidden}
}

// SecurityError represents a security validation error
type SecurityError struct {
	Message    string
	StatusCode int
}

func (e *SecurityError) Error() string {
	return e.Message
}
