package utils

import "testing"

func TestEncryptDecrypt_WithNormalizedKeyLengths(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{
			name: "exact_32_byte_key",
			key:  "12345678901234567890123456789012",
		},
		{
			name: "non_standard_key_length_is_hashed",
			key:  "short-key",
		},
	}

	plaintext := `{"api_key":"secret-api-key-123"}`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ENCRYPTION_KEY", tt.key)

			encrypted, err := Encrypt(plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}
			if encrypted == plaintext {
				t.Fatalf("expected encrypted payload to differ from plaintext")
			}

			decrypted, err := Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}
			if decrypted != plaintext {
				t.Fatalf("unexpected decrypted value: got %q want %q", decrypted, plaintext)
			}
		})
	}
}
