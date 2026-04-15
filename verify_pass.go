package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash := "$2a$10$B2/XZniBnqEmecEhQGzUn.9ZU5sNpthotIAmxEJcWqGEkodD7Jp9i"
	password := "admin123"
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Printf("Verification failed: %v\n", err)
	} else {
		fmt.Println("Verification successful!")
	}
}
