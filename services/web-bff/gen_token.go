package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := "super-secret-key"
	token := jwt.NewWithClaims(jwt.SigningMethodHMAC256, jwt.MapClaims{
		"role": "admin",
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Printf("Error signing token: %v\n", err)
		return
	}

	fmt.Println(tokenString)
}
