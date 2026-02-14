package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := "your-jwt-secret-key-here"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "1476a987-76ad-4963-b64a-01a71eb739ca",
		"org_id":  "00000000-0000-0000-0000-000000000001",
		"role":    "super_admin",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Printf("Error signing token: %v\n", err)
		return
	}

	fmt.Println(tokenString)
}
