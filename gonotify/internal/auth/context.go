package auth

import "context"

type contextKey struct{}

type APIKey struct {
	ID          string
	TenantID    string
	Environment string
	Scopes      []string
}

func (k APIKey) HasScope(required string) bool {
	if len(k.Scopes) == 0 {
		// If key has no explicit scopes assigned, permit (full access)
		return true
	}
	for _, s := range k.Scopes {
		if s == "*" || s == required || s == "admin" {
			return true
		}
	}
	return false
}

func (k APIKey) HasAnyScope(required ...string) bool {
	for _, req := range required {
		if k.HasScope(req) {
			return true
		}
	}
	return false
}

func WithAPIKey(ctx context.Context, key APIKey) context.Context {
	return context.WithValue(ctx, contextKey{}, key)
}

func APIKeyFromContext(ctx context.Context) (APIKey, bool) {
	key, ok := ctx.Value(contextKey{}).(APIKey)
	return key, ok
}
