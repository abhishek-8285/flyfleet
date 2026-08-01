package tenant

import "context"

type contextKey struct{}

func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, contextKey{}, tenantID)
}

func FromContext(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(contextKey{}).(string)
	return tenantID, ok
}
