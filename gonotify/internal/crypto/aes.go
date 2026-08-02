package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

// LoadKeyFromEnv reads PROVIDER_ENCRYPTION_KEY from the environment.
// The key must be a base64-encoded 32-byte AES key.
// Returns nil (with no error) if the env var is not set or empty,
// indicating that encryption is disabled.
func LoadKeyFromEnv() ([]byte, error) {
	raw := os.Getenv("PROVIDER_ENCRYPTION_KEY")
	if raw == "" {
		return nil, nil
	}

	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode PROVIDER_ENCRYPTION_KEY: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("PROVIDER_ENCRYPTION_KEY must decode to 32 bytes, got %d", len(key))
	}

	return key, nil
}

// Encrypt encrypts plaintext using AES-GCM with the given key.
// The returned ciphertext has the nonce prepended.
// If key is empty, returns plaintext unchanged (encryption disabled).
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) == 0 {
		return plaintext, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM mode: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts ciphertext (nonce + ciphertext) using AES-GCM with the given key.
// If key is empty, returns ciphertext unchanged.
// If ciphertext is empty, returns nil.
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, nil
	}

	if len(key) == 0 {
		return ciphertext, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM mode: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}
