package crypto

import (
	"encoding/base64"
	"testing"
)

func TestLoadKeyFromEnv(t *testing.T) {
	t.Run("returns nil when env unset", func(t *testing.T) {
		t.Setenv("PROVIDER_ENCRYPTION_KEY", "")
		key, err := LoadKeyFromEnv()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key != nil {
			t.Errorf("expected nil key, got %d bytes", len(key))
		}
	})

	t.Run("returns decoded key", func(t *testing.T) {
		key := make([]byte, 32)
		for i := range key {
			key[i] = byte(i)
		}
		t.Setenv("PROVIDER_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(key))

		got, err := LoadKeyFromEnv()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 32 {
			t.Fatalf("expected 32-byte key, got %d", len(got))
		}
	})

	t.Run("rejects wrong-size key", func(t *testing.T) {
		t.Setenv("PROVIDER_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString([]byte("too-short")))
		_, err := LoadKeyFromEnv()
		if err == nil {
			t.Fatal("expected error for wrong-size key")
		}
	})
}

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}

	t.Run("round-trip", func(t *testing.T) {
		plaintext := []byte(`{"username":"smtp_user","password":"s3cret"}`)

		encrypted, err := Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		if string(encrypted) == string(plaintext) {
			t.Fatal("ciphertext should differ from plaintext")
		}

		decrypted, err := Decrypt(encrypted, key)
		if err != nil {
			t.Fatalf("decrypt failed: %v", err)
		}

		if string(decrypted) != string(plaintext) {
			t.Fatalf("round-trip mismatch: got %q, want %q", decrypted, plaintext)
		}
	})

	t.Run("empty key returns plaintext unchanged", func(t *testing.T) {
		plaintext := []byte("hello")
		result, err := Encrypt(plaintext, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(result) != string(plaintext) {
			t.Fatal("with empty key, Encrypt should return plaintext unchanged")
		}

		decrypted, err := Decrypt(result, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(decrypted) != string(plaintext) {
			t.Fatal("with empty key, Decrypt should return ciphertext unchanged")
		}
	})

	t.Run("empty ciphertext returns nil", func(t *testing.T) {
		result, err := Decrypt(nil, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Fatalf("expected nil, got %v", result)
		}
	})

	t.Run("wrong key fails to decrypt", func(t *testing.T) {
		plaintext := []byte("secret")
		encrypted, err := Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		wrongKey := make([]byte, 32)
		_, err = Decrypt(encrypted, wrongKey)
		if err == nil {
			t.Fatal("expected decryption error with wrong key")
		}
	})
}
