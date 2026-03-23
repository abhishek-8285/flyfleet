package service

import "testing"

func TestSanitizeAuditChanges_RedactsSensitiveKeysRecursively(t *testing.T) {
	input := map[string]any{
		"status":  "ok",
		"api_key": "plain-secret",
		"nested": map[string]any{
			"authorization": "Bearer token-value",
			"safe":          "value",
		},
		"array": []any{
			map[string]any{"password": "p@ss"},
			"plain",
		},
	}

	out := sanitizeAuditChanges(input)

	if out["status"] != "ok" {
		t.Fatalf("expected non-sensitive key to remain unchanged, got %v", out["status"])
	}
	if out["api_key"] != redactedAuditValue {
		t.Fatalf("expected api_key to be redacted, got %v", out["api_key"])
	}

	nested, ok := out["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map, got %T", out["nested"])
	}
	if nested["authorization"] != redactedAuditValue {
		t.Fatalf("expected nested authorization to be redacted, got %v", nested["authorization"])
	}
	if nested["safe"] != "value" {
		t.Fatalf("expected nested safe key unchanged, got %v", nested["safe"])
	}

	array, ok := out["array"].([]any)
	if !ok || len(array) != 2 {
		t.Fatalf("expected array with 2 elements, got %T len=%d", out["array"], len(array))
	}
	firstMap, ok := array[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first array item to be map, got %T", array[0])
	}
	if firstMap["password"] != redactedAuditValue {
		t.Fatalf("expected array nested password redacted, got %v", firstMap["password"])
	}
}

func TestSanitizeAuditChanges_ReturnsNilForNilInput(t *testing.T) {
	if sanitizeAuditChanges(nil) != nil {
		t.Fatal("expected nil output for nil input")
	}
}
