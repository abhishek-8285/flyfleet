package main

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

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
	pass := "Abhishek@123"
	hash := "$argon2id$v=19$m=65536,t=3,p=4$VBpRoRoX1vbgc7Ux5+hRkQ$1QCMF72G9El+ndDo6kjE1hwKeAk/cNp+JzXY4C3Cjlk"
	ok, err := comparePassword(pass, hash)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Match:", ok)
}
