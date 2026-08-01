package user

import (
	"context"
	"testing"
)

func TestUserContext(t *testing.T) {
	ctx := context.Background()

	_, ok := FromContext(ctx)
	if ok {
		t.Error("FromContext() on empty context should return false")
	}

	usr := AuthenticatedUser{
		ID:             "usr_123",
		TenantID:       "tenant_456",
		ExternalUserID: "ext_789",
	}

	ctxWithUser := WithAuthenticatedUser(ctx, usr)
	gotUser, ok := FromContext(ctxWithUser)

	if !ok || gotUser.ID != usr.ID {
		t.Errorf("FromContext() = %v, ok=%v; want %v, true", gotUser, ok, usr)
	}
}

func TestNewRepository(t *testing.T) {
	repo := NewRepository(nil)
	if repo == nil {
		t.Fatal("NewRepository(nil) returned nil")
	}

	// Test EnsureUser with unmarshalable attributes
	badAttr := map[string]any{
		"unsupported": make(chan int),
	}
	_, err := repo.EnsureUser(context.Background(), "tenant_1", "ext_1", "email@example.com", "+14155552671", badAttr)
	if err == nil {
		t.Error("EnsureUser expected JSON marshal error for bad attribute, got nil")
	}
}
