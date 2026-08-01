package tenant

import (
	"context"
	"testing"
)

func TestTenantContext(t *testing.T) {
	ctx := context.Background()

	// 1. Missing tenant ID from raw context
	tenantID, ok := FromContext(ctx)
	if ok || tenantID != "" {
		t.Errorf("FromContext() on empty context returned tenantID=%q, ok=%v; want empty, false", tenantID, ok)
	}

	// 2. Set tenant ID in context
	expectedTenantID := "00000000-0000-0000-0000-000000000001"
	ctxWithTenant := WithTenantID(ctx, expectedTenantID)

	gotTenantID, ok := FromContext(ctxWithTenant)
	if !ok || gotTenantID != expectedTenantID {
		t.Errorf("FromContext() = %q, ok=%v; want %q, true", gotTenantID, ok, expectedTenantID)
	}
}
