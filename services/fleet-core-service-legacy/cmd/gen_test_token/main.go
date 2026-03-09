package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"org_id":  "00000000-0000-0000-0000-000000000001",
		"user_id": "e1a510a4-d48c-4b54-bb9e-191de2067d3d",
		"role":    "admin",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte("runtime-integration-secret"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Print(tokenString)
}
