package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash := []byte("$2a$10$J4XbQEebM3lZlbrlyN4nAOc2XzO4vHA5DFIc0iVycghLgtzRrzUPC")
	pass := []byte("password")

	err := bcrypt.CompareHashAndPassword(hash, pass)
	if err != nil {
		fmt.Println("Mismatch:", err)
	} else {
		fmt.Println("Match!")
	}
}
