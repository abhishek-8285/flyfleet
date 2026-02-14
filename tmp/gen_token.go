package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := []byte("your-jwt-secret-key-here")
	
	claims := jwt.MapClaims{
		"user_id": "00000000-0000-0000-0000-000000000001",
		"org_id":  "00000000-0000-0000-0000-000000000001",
		"email":   "admin@flyfleet.io",
		"role":    "super_admin",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		fmt.Printf("Error signing token: %v\n", err)
		return
	}

	fmt.Println(tokenString)
}
