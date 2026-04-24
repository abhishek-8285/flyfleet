package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os/exec"
	"strings"
	"testing"
)

func TestHMACMatch(t *testing.T) {
	secret := "FF-SECRET-2026"
	message := "550e8400-e29b-41d4-a716-446655440000550e8400-e29b-41d4-a716-446655440001"

	// Go calculation
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	goSig := hex.EncodeToString(mac.Sum(nil))

	// OpenSSL calculation (simulating bash script)
	cmd := exec.Command("bash", "-c", "echo -n '"+message+"' | openssl dgst -sha256 -hmac '"+secret+"' | sed 's/^.* //'")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run openssl: %v", err)
	}
	bashSig := strings.TrimSpace(string(out))

	t.Logf("Go Sig:   %s", goSig)
	t.Logf("Bash Sig: %s", bashSig)

	if goSig != bashSig {
		t.Errorf("Signature mismatch! Go is lowercase, Bash is likely uppercase or formatted differently.")
	}
}
