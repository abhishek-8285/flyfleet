package auth

import (
	"testing"
)

func TestParseBearer(t *testing.T) {
	tests := []struct {
		header    string
		wantToken string
		wantErr   bool
	}{
		{header: "Bearer token123", wantToken: "token123", wantErr: false},
		{header: "bearer token123", wantToken: "token123", wantErr: false},
		{header: "Bearer   token123  ", wantToken: "token123", wantErr: false},
		{header: "Basic token123", wantToken: "", wantErr: true},
		{header: "Bearer", wantToken: "", wantErr: true},
		{header: "Bearer   ", wantToken: "", wantErr: true},
		{header: "", wantToken: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got, err := ParseBearer(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBearer(%q) error = %v; wantErr %v", tt.header, err, tt.wantErr)
				return
			}
			if got != tt.wantToken {
				t.Errorf("ParseBearer(%q) = %q; want %q", tt.header, got, tt.wantToken)
			}
		})
	}
}

func TestHashAPIKeyDeterministic(t *testing.T) {
	key := "gn_dev_local_0123456789abcdef0123456789abcdef"
	hash1 := HashAPIKey(key)
	hash2 := HashAPIKey(key)

	if hash1 == "" {
		t.Fatal("HashAPIKey returned empty hash")
	}

	if hash1 != hash2 {
		t.Errorf("HashAPIKey is not deterministic: %q != %q", hash1, hash2)
	}
}

func TestGenerateAPIKey(t *testing.T) {
	raw, prefix, hash, err := GenerateAPIKey("test")
	if err != nil {
		t.Fatalf("GenerateAPIKey() unexpected error: %v", err)
	}

	if raw == "" || prefix == "" || hash == "" {
		t.Errorf("GenerateAPIKey() returned empty outputs: raw=%q, prefix=%q, hash=%q", raw, prefix, hash)
	}

	if HashAPIKey(raw) != hash {
		t.Errorf("HashAPIKey(raw) does not match generated hash")
	}

	// Default environment fallback
	rawDefault, _, _, err := GenerateAPIKey("")
	if err != nil {
		t.Fatalf("GenerateAPIKey('') unexpected error: %v", err)
	}
	if rawDefault[:6] != "gn_dev" {
		t.Errorf("GenerateAPIKey('') environment default failed: %q", rawDefault)
	}
}
