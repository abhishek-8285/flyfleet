package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "admin123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Generated Hash: %s\n", hash)

	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		fmt.Println("Comparison failed!")
	} else {
		fmt.Println("Comparison successful!")
	}
}
