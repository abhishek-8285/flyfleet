package main

import (
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
	"crypto/hmac"
)

func main() {
	password := "password"
	// Example parts from a typical hash
	salt, _ := base64.RawStdEncoding.DecodeString("MTIzNDU2Nzg5MDEyMzQ1Ng") // 16 bytes dummy
	
	iterations := uint32(3)
	memory := uint32(64 * 1024)
	parallelism := uint8(4)
	keyLength := uint32(32)

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, parallelism, b64Salt, b64Hash)

	fmt.Println("Generated:", encoded)
	
	// Simulating comparePassword logic
	decodedHash, _ := base64.RawStdEncoding.DecodeString(b64Hash)
	match := hmac.Equal(hash, decodedHash)
	fmt.Printf("Self-match test: %v\n", match)
}
