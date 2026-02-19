package service

import "strings"

const redactedAuditValue = "***REDACTED***"

var sensitiveAuditKeyHints = []string{
	"password",
	"secret",
	"token",
	"api_key",
	"apikey",
	"auth",
	"authorization",
	"credential",
	"private_key",
}

func sanitizeAuditChanges(changes map[string]any) map[string]any {
	if changes == nil {
		return nil
	}
	sanitized := make(map[string]any, len(changes))
	for k, v := range changes {
		if isSensitiveAuditKey(k) {
			sanitized[k] = redactedAuditValue
			continue
		}
		sanitized[k] = sanitizeAuditValue(v)
	}
	return sanitized
}

func sanitizeAuditValue(v any) any {
	switch typed := v.(type) {
	case map[string]any:
		return sanitizeAuditChanges(typed)
	case map[string]string:
		out := make(map[string]any, len(typed))
		for k, value := range typed {
			if isSensitiveAuditKey(k) {
				out[k] = redactedAuditValue
				continue
			}
			out[k] = value
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i := range typed {
			out[i] = sanitizeAuditValue(typed[i])
		}
		return out
	default:
		return v
	}
}

func isSensitiveAuditKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	for _, hint := range sensitiveAuditKeyHints {
		if strings.Contains(normalized, hint) {
			return true
		}
	}
	return false
}
