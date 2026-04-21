package security

import (
	"context"

	"github.com/google/uuid"
	"github.com/fleet-platform/shared/pkg/middleware"
)

// GetTenantID extracts the organization ID from the context.
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	tenantIDStr, ok := ctx.Value(middleware.TenantIDKey).(string)
	if !ok {
		return uuid.Nil, false
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return uuid.Nil, false
	}

	return tenantID, true
}

// MustGetTenantID extracts the organization ID from the context or panics.
func MustGetTenantID(ctx context.Context) uuid.UUID {
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		panic("tenant_id missing from context")
	}
	return tenantID
}
