package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "Password123!"
	hash := "$2a$04$WOojoGDL.4DzIGNd3opDV.phzq0f8tFT8pEQXk2StJ7YvyGPDBahq"
	
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Printf("Password mismatch: %v\n", err)
	} else {
		fmt.Println("Password match!")
	}

	newHash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	fmt.Printf("New hash (cost 12): %s\n", string(newHash))
}
