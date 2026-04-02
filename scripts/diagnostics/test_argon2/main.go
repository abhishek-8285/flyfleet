package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"golang.org/x/crypto/argon2"
	"crypto/subtle"
)

const (
	Argon2Iterations  = 3
	Argon2Memory      = 64 * 1024
	Argon2Parallelism = 4
	Argon2SaltLength  = 16
	Argon2KeyLength   = 32
)

func generateArgon2Hash(password string) (string, error) {
	salt := make([]byte, Argon2SaltLength)
	rand.Read(salt)
	hash := argon2.IDKey([]byte(password), salt, uint32(Argon2Iterations), uint32(Argon2Memory), uint8(Argon2Parallelism), uint32(Argon2KeyLength))
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, Argon2Memory, Argon2Iterations, Argon2Parallelism, b64Salt, b64Hash), nil
}

func comparePassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format: %d parts", len(parts))
	}
	salt, _ := base64.RawStdEncoding.DecodeString(parts[4])
	decodedHash, _ := base64.RawStdEncoding.DecodeString(parts[5])
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, uint32(len(decodedHash)))
	return subtle.ConstantTimeCompare(hash, decodedHash) == 1, nil
}

func main() {
	pass := "Password123!"
	h, err := generateArgon2Hash(pass)
	if err != nil {
		fmt.Println("Error generating:", err)
		return
	}
	fmt.Println("Hash:", h)
	ok, err := comparePassword(pass, h)
	if err != nil {
		fmt.Println("Error comparing:", err)
		return
	}
	fmt.Println("Match:", ok)
}
