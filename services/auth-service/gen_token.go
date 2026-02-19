package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := []byte("your-jwt-secret-key-here")
	claims := jwt.MapClaims{
		"user_id": "16bc21dc-c143-4fe9-94c1-0d70ad1bd705",
		"org_id":  "00000000-0000-0000-0000-000000000001",
		"role":    "super_admin",
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(tokenString)
}
