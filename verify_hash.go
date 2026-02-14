package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash := "$2b$12$k2EpictQdcdbr7tR9N8LdeuosO8ZTzDwir2K3cOc7esxoKS5aFFzm"
	password := "admin123"
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Printf("Verification failed: %v\n", err)
	} else {
		fmt.Println("Verification successful!")
	}
}
