package auth

import (
	"context"
	"testing"
)

func TestAPIKeyScopes(t *testing.T) {
	// Empty scopes = full access
	emptyKey := APIKey{}
	if !emptyKey.HasScope("events:write") {
		t.Error("emptyKey.HasScope() should return true (full access)")
	}

	// Wildcard scope
	wildcardKey := APIKey{Scopes: []string{"*"}}
	if !wildcardKey.HasScope("templates:write") {
		t.Error("wildcardKey.HasScope() should return true for '*'")
	}

	// Admin scope
	adminKey := APIKey{Scopes: []string{"admin"}}
	if !adminKey.HasScope("providers:write") {
		t.Error("adminKey.HasScope() should return true for 'admin'")
	}

	// Explicit scopes matching
	scopedKey := APIKey{Scopes: []string{"events:write", "notifications:read"}}
	if !scopedKey.HasScope("events:write") {
		t.Error("scopedKey.HasScope('events:write') should return true")
	}
	if scopedKey.HasScope("templates:write") {
		t.Error("scopedKey.HasScope('templates:write') should return false")
	}

	// HasAnyScope true
	if !scopedKey.HasAnyScope("templates:write", "notifications:read") {
		t.Error("scopedKey.HasAnyScope() should return true for matching scope in slice")
	}

	// HasAnyScope false
	if scopedKey.HasAnyScope("templates:write", "providers:write") {
		t.Error("scopedKey.HasAnyScope() should return false when no scopes match")
	}
}

func TestAPIKeyContext(t *testing.T) {
	ctx := context.Background()

	_, ok := APIKeyFromContext(ctx)
	if ok {
		t.Error("APIKeyFromContext() on empty context should return false")
	}

	key := APIKey{
		ID:          "key_123",
		TenantID:    "tenant_456",
		Environment: "production",
		Scopes:      []string{"events:write"},
	}

	ctxWithKey := WithAPIKey(ctx, key)
	gotKey, ok := APIKeyFromContext(ctxWithKey)
	if !ok || gotKey.ID != key.ID {
		t.Errorf("APIKeyFromContext() = %v, ok=%v; want %v, true", gotKey, ok, key)
	}
}
