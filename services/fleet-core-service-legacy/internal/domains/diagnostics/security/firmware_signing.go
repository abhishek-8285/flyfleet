package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateFirmwareKeypair creates a new Ed25519 keypair for firmware signing.
func GenerateFirmwareKeypair() (pub string, priv string, err error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return hex.EncodeToString(pubKey), hex.EncodeToString(privKey), nil
}

// SignFirmware signs the firmware data using the private key.
func SignFirmware(data []byte, privateKeyHex string) (string, error) {
	privKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	if len(privKeyBytes) != ed25519.PrivateKeySize {
		return "", fmt.Errorf("invalid private key size")
	}

	sig := ed25519.Sign(privKeyBytes, data)
	return hex.EncodeToString(sig), nil
}

// VerifyFirmware checks if the signature is valid for the given data and public key.
func VerifyFirmware(data []byte, signatureHex string, publicKeyHex string) (bool, error) {
	pubKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return false, fmt.Errorf("invalid public key: %w", err)
	}

	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false, fmt.Errorf("invalid signature: %w", err)
	}

	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return false, fmt.Errorf("invalid public key size")
	}

	return ed25519.Verify(pubKeyBytes, data, sigBytes), nil
}
