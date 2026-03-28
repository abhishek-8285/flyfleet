package secrets

import (
	"context"
	"os"
	"strings"
)

// SecretFetcher defines the enterprise interface for retrieving sensitive credentials
type SecretFetcher interface {
	Fetch(ctx context.Context, key string) (string, error)
}

// EnterpriseSecretManager handles dynamic secret retrieval (Credstash / KMS)
type EnterpriseSecretManager struct {
	// In a real implementation, this would have the AWS SDK / Credstash client
}

func (m *EnterpriseSecretManager) Fetch(ctx context.Context, value string) (string, error) {
	// Pattern: If the value starts with "env:", we fetch from OS environment (Free & Secure)
	if strings.HasPrefix(value, "env:") {
		envKey := strings.TrimPrefix(value, "env:")
		resolved := os.Getenv(envKey)

		if resolved == "" {
			return "MISSING_ENV_VAR_" + envKey, nil
		}
		return resolved, nil
	}

	// Pattern: Keep credstash support as a mock for now
	if strings.HasPrefix(value, "credstash:") {
		return "MOCK_STASH_" + strings.TrimPrefix(value, "credstash:"), nil
	}

	// Fallback: If it's not a secret reference, return the plain value
	return value, nil
}

// ProcessConfigSecrets is a helper to resolve all sensitive fields in the config
func ResolveSecret(ctx context.Context, value string) string {
	mgr := &EnterpriseSecretManager{}
	resolved, err := mgr.Fetch(ctx, value)
	if err != nil {
		return value // Fallback to original
	}
	return resolved
}
