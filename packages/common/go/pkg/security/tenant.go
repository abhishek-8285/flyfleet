package security

import (
	"context"

	"github.com/google/uuid"
	"github.com/fleet-platform/shared/pkg/middleware"
)

// GetOrgID extracts the organization ID from the context.
func GetOrgID(ctx context.Context) (uuid.UUID, bool) {
	orgIDStr, ok := ctx.Value(middleware.OrgIDKey).(string)
	if !ok {
		return uuid.Nil, false
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return uuid.Nil, false
	}

	return orgID, true
}

// MustGetOrgID extracts the organization ID from the context or panics.
func MustGetOrgID(ctx context.Context) uuid.UUID {
	orgID, ok := GetOrgID(ctx)
	if !ok {
		panic("org_id missing from context")
	}
	return orgID
}
